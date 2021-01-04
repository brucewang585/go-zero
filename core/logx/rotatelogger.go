package logx

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/brucewang585/go-zero/core/fs"
	"github.com/brucewang585/go-zero/core/lang"
	"github.com/brucewang585/go-zero/core/timex"
)

const (
	dateFormat      = "2006-01-02"
	baseTimeFormat  = "2006-01-02_15-04-05"
	hoursPerDay     = 24
	bufferSize      = 1000
	defaultDirMode  = 0755
	defaultFileMode = 0600
)

var ErrLogFileClosed = errors.New("error: log file closed")

type (	
	RotateRule interface {
		Init()

		//
		ShallRotate() bool
		MarkRotated()
		CurrentFile() string //��ȡ
		OutdatedFiles() ([]string,error)

		//
		AddSize(s int64)
	}

	RotateLogger struct {
		//��̬����
		filename string
		rule     RotateRule
		compress bool

		//��̬����
		curfile        string //filename+���ָ�����+ʱ��(BaseTimeFormat)
		fp             *os.File
		done           chan lang.PlaceholderType
		write_channel  chan []byte
		rotate_channel chan bool

		// can't use threading.RoutineGroup because of cycle import
		waitGroup sync.WaitGroup
		closeOnce sync.Once
	}
)

func DefaultRotateRule(filename, delimiter string, keepAge int64, rotateSize int64) RotateRule {
	return &DailyRotateRule{
		filename:    filename,
		delimiter:   delimiter,
		keepAge:     keepAge,
		rotatedSize:  rotateSize,
	}
}

func NewLogger(filename string, rule RotateRule, compress bool) (*RotateLogger, error) {
	l := &RotateLogger{
		filename:       filename,
		rule:           rule,
		done:           make(chan lang.PlaceholderType),
		write_channel:  make(chan []byte, bufferSize),
		rotate_channel: make(chan bool, 100),
	}
	if err := l.init(); err != nil {
		return nil, err
	}

	l.startWorker()
	return l, nil
}

func (l *RotateLogger) Close() error {
	var err error

	l.closeOnce.Do(func() {
		close(l.done)
		l.waitGroup.Wait()

		if err = l.fp.Sync(); err != nil {
			return
		}

		err = l.fp.Close()
	})

	return err
}

func (l *RotateLogger) Write(data []byte) (int, error) {
	select {
	case l.write_channel <- data:
		return len(data), nil
	case <-l.done:
		log.Println(string(data))
		return 0, ErrLogFileClosed
	}
}

func (l *RotateLogger) init() error {
	//ע�⣬һ��Ҫ���ã���rule��ʼ��
	l.rule.Init()

	//
	l.curfile = l.rule.CurrentFile()
	if _, err := os.Stat(l.curfile); err != nil {
		basePath := path.Dir(l.curfile)
		if _, err = os.Stat(basePath); err != nil {
			if err = os.MkdirAll(basePath, defaultDirMode); err != nil {
				return err
			}
		}

		if l.fp, err = os.Create(l.curfile); err != nil {
			return err
		}
	} else if l.fp, err = os.OpenFile(l.curfile, os.O_APPEND|os.O_WRONLY, defaultFileMode); err != nil {
		return err
	}

	fs.CloseOnExec(l.fp)

	return nil
}

func (l *RotateLogger) startWorker() {
	l.waitGroup.Add(1)
	go func() {
		defer l.waitGroup.Done()

		for {
			select {
			case event := <-l.write_channel:
				l.writeToFile(event)

			case <-l.done:
				return
			}
		}
	}()

	l.waitGroup.Add(1)
	go func() {
		defer l.waitGroup.Done()

		for {
			select {
			case <-l.rotate_channel:
				l.maybeDeleteOutdatedFiles()

			case <-l.done:
				return
			}
		}
	}()
	//������1��
	l.rotate_channel<-true
}

func (l *RotateLogger) writeToFile(v []byte) {
	if l.rule.ShallRotate() {
		//Rotate
		//�رվ��ļ�
		if l.fp != nil {
			l.fp.Close()
			l.fp = nil
		}

		//�ƺ����������������ļ�
		l.rotate_channel <- true

		//
		l.rule.MarkRotated()

		//�����ļ�
		l.curfile = l.rule.CurrentFile()
		if fp, err := os.Create(l.curfile); err == nil {
			l.fp = fp
			fs.CloseOnExec(l.fp)
		}
	}
	if l.fp != nil {
		l.fp.Write(v)
		l.rule.AddSize(int64(len(v)))
	}
}

func (l *RotateLogger) maybeCompressFile(file string) {
	if !l.compress {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			ErrorStack(r)
		}
	}()
	compressLogFile(file)
}

func (l *RotateLogger) maybeDeleteOutdatedFiles() {
	files,_ := l.rule.OutdatedFiles()
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			Errorf("failed to remove outdated file: %s,error:%v", file,err)
		}
	}
}

func compressLogFile(file string) {
	return
	start := timex.Now()
	Infof("compressing log file: %s", file)
	if err := gzipFile(file); err != nil {
		Errorf("compress error: %s", err)
	} else {
		Infof("compressed log file: %s, took %s", file, timex.Since(start))
	}
}

func getNowDate() string {
	return time.Now().Format(dateFormat)
}

func getNowAllTime() (string,string) {
	now := time.Now()
	return now.Format(dateFormat),now.Format(baseTimeFormat)
}

func gzipFile(file string) error {
	in, err := os.Open(file)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(fmt.Sprintf("%s.gz", file))
	if err != nil {
		return err
	}
	defer out.Close()

	w := gzip.NewWriter(out)
	if _, err = io.Copy(w, in); err != nil {
		return err
	} else if err = w.Close(); err != nil {
		return err
	}

	return os.Remove(file)
}
