package test

import (
	"github.com/Lzw2016/gow/gowatch"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"
)

func Test0101(t *testing.T) {
	demo()
}

func demo() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("D:\\SourceCode\\github\\gowatch")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func Test0102(t *testing.T) {
	wd, _ := os.Getwd()
	filePath := filepath.Join(wd, "../example/gowatch.yml")
	log.Printf("wd=%s | gowatch.yml=%s", wd, filePath)
	config := gowatch.ParseConfig(filePath)
	log.Printf("cfg=%#v", config)
}

func Test0103(t *testing.T) {
	f := func() {
		log.Printf("# ----------------->")
	}
	debounced := gowatch.NewDebounced(100 * time.Millisecond)

	debounced(f)
	time.Sleep(200 * time.Millisecond)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func Test0104(t *testing.T) {
	wd, _ := os.Getwd()
	filePath, _ := filepath.Abs("./")
	logrus.WithFields(logrus.Fields{"path": filePath}).Infof("path")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.WithFields(logrus.Fields{"event": event}).Infof("event")
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.WithFields(logrus.Fields{"err": err}).Warn("err")
			}
		}
	}()
	err = watcher.Add(filepath.Join(wd, "./01_test.go"))
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func Test0105(t *testing.T) {
	//logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000", ForceColors: false})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Debugf("111 %d", 1)
	logrus.Infof("222 %d", 2)
	logrus.Warnf("333 %d", 3)
	logrus.Errorf("444 %d", 4)

	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("a111")

	logrus.WithFields(logrus.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("b222")

	logrus.WithFields(logrus.Fields{
		"omg":    true,
		"number": 100,
	}).Error("c 333")
	//logrus.Panic("c 333", "c 333")
}

func Test0106(t *testing.T) {
	r, _ := regexp.Compile("(\\w+).tmp$")
	logrus.Println("MatchString = ", r.MatchString("D:\\SourceCode\\tmp\\go-test\\main\\main2.go"))
}

func Test0107(t *testing.T) {
	var cmd *exec.Cmd
	go func() {

	}()
	cmd = exec.Command("D:\\SourceCode\\tmp\\go-test\\app.exe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
	time.Sleep(20 * time.Second)
	err := cmd.Process.Signal(syscall.SIGKILL)
	if err != nil {
		log.Printf("%#v", err)
	}
}
