package logx
/*
一天1切，如果文件的尺寸大于阀值，还需要切
*/
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DailyRotateRule struct {
	filename  string
	delimiter string

	//
	rotatedSize int64 //切割尺寸的最大长度
	keepAge     int64 //保留秒数，单位是秒

	//动态
	curSize int64  //当前文件长度  curSize < rotatedSize
	curDate string //当前文件日期,

	//当前文件
	curFile string //每次
}

func (r *DailyRotateRule) Init() {
	r.MarkRotated()

	if r.keepAge == 0 {
		r.keepAge = minKeepAge
	}
	if r.rotatedSize == 0 {
		r.rotatedSize = minRotateSize
	}

	//重新选择当前文件，如果本天最近文件空间还不满，可以选择最近文件，如果没找到或者空间已经满了，就用新文件
	//枚举文件，找到已有文件
	pattern  := fmt.Sprintf("%s%s%s*", r.filename, r.delimiter,r.curDate)
	files, _ := filepath.Glob(pattern)
	if len(files) > 0 {
		sort.Strings(files)

		//查找最后1个是否还有空间
		lastFile := files[len(files) -1 ]
		st,err := os.Stat(lastFile)
		if err == nil {
			if st.Size() < r.rotatedSize {
				r.curFile = lastFile
				return
			}
		}
	}
}

//时间有跨度，就切割文件，如果时间没有跨度，当文件长度大于rotatedSize
func (r *DailyRotateRule) ShallRotate() bool {
	if getNowDate() != r.curDate {
		return true
	} else {
		//
		if r.curSize > r.rotatedSize {
			return true
		} else {
			return false
		}
	}
}

func (r *DailyRotateRule) MarkRotated() {
	var curTime string
	r.curSize  = 0
	r.curDate,curTime = getNowAllTime()
	r.curFile  = fmt.Sprintf("%s%s%s", r.filename, r.delimiter,curTime)
}

func (r *DailyRotateRule) CurrentFile() string {
	return r.curFile
}

func (r *DailyRotateRule) OutdatedFiles() ([]string,error) {
	pattern := fmt.Sprintf("%s%s*", r.filename, r.delimiter)

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil,fmt.Errorf("failed to delete outdated log files, error: %s", err)
	}

	drop_tm := time.Now().Add(-time.Second * time.Duration(r.keepAge))
	var buf strings.Builder
	boundary := drop_tm.Format(baseTimeFormat)
	fmt.Fprintf(&buf, "%s%s%s", filepath.Base(r.filename), r.delimiter, boundary)
	boundaryFile := buf.String()

	var outdates []string
	for _, file := range files {
		if filepath.Base(file) < boundaryFile {
			//再判断文件时间
			st,err := os.Stat(file)
			if err == nil {
				if st.ModTime().Unix() < drop_tm.Unix() {
					outdates = append(outdates, file)
				}
			}
		}
	}

	return outdates,nil
}

func (r *DailyRotateRule) AddSize(size int64) {
	r.curSize += size
}
