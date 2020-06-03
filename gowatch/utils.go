package gowatch

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000", ForceQuote: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}
