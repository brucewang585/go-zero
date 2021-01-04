package logx

import (
	"strconv"
)

const (
	minKeepAge = 60
	minRotateSize = 1024//5*1024*1024
	defaultKeepAge = 7*24*3600
)

type LogConf struct {
	ServiceName         string `json:",optional"`
	Mode                string `json:",default=console,options=console|file|volume"`
	Path                string `json:",default=logs"`
	Level               string `json:",default=info,options=info|error|severe"`
	Compress            bool   `json:",optional"`
	KeepDays            int    `json:",optional"`
	StackCooldownMillis int    `json:",default=100"`
	KeepAge             string `json:",optional"` //1d,1h,1m,1s, ��д��λĬ��Ϊ�룬����ֵ������1����
	RotateSize          string `json:",optional"` //����1��1�У�ͬʱ���ļ��������,1g,1m,1k����д��λĬ��Ϊ�ֽڣ�����ֵ������10m
}

func KeepAge2I(s string) int64 {
	if s == "" {
		return minKeepAge
	}

	//
	var age int64
	t,_ := strconv.Atoi(s[:len(s)-1])
	switch s[len(s)-1] {
	case 'd','D':
		age = int64(t)*24*3600
	case 'h','H':
		age = int64(t)*3600
	case 'm','M':
		age = int64(t)*60
	case 's','S':
	default:
	}

	if age < minKeepAge {
		age = minKeepAge
	}
	return age
}

func RotateSize2I(s string) int64 {
	if s == "" {
		return minRotateSize
	}

	//
	var size int64
	t,_ := strconv.Atoi(s[:len(s)-1])
	switch s[len(s)-1] {
	case 'g','G':
		size = int64(t)*1024*1024*1024
	case 'm','M':
		size = int64(t)*1024*1024
	case 'k','K':
		size = int64(t)*1024
	default:
	}

	if size < minRotateSize {
		size = minRotateSize
	}
	return size
}


