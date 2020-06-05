package gowatch

import (
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var ignoredFilesRegExps = make([]regexp.Regexp, 0, 8)

func init() {
	ignoredFiles := []string{
		`.#(\w+).go$`,
		`.(\w+).go.swp$`,
		`(\w+).go~$`,
		`(\w+).tmp$`,
		`(\w+).exe$`,
	}
	for _, regex := range ignoredFiles {
		r, err := regexp.Compile(regex)
		if err != nil {
			log.WithFields(logrus.Fields{"regex": regex}).Error("正则表达式错误", err)
			os.Exit(2)
		}
		ignoredFilesRegExps = append(ignoredFilesRegExps, *r)
	}
}

// gowatch 对象
type Watcher struct {
	// 工作路径(绝对路径)
	workPath string
	// 配置
	config Config

	// 监控目录
	paths *hashset.Set
	// 防抖动函数
	debounced func(f func())
	// 文件监听器
	pathWatcher *fsnotify.Watcher
	outPut      string
	cmd         *exec.Cmd
	buildState  sync.Mutex
}

func NewWatcher(workPath string, config *Config) *Watcher {
	watcher := &Watcher{
		workPath:  workPath,
		config:    *config,
		paths:     hashset.New(),
		debounced: NewDebounced(time.Duration(config.Build.Delay) * time.Millisecond),
	}
	pathWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panic("创建监听器失败", err)
		os.Exit(2)
	}
	watcher.pathWatcher = pathWatcher
	go watcher.startWatch()
	return watcher
}

func (th *Watcher) StartWatch() {
	th.addWatchPaths(th.workPath)
	// 除了当前目录，增加额外监听的目录
	for _, wPath := range th.config.Watch.Paths {
		wPath = filepath.Join(th.workPath, wPath)
		th.addWatchPaths(wPath)
	}
	th.addWatch()
	go th.goBuild()
}

func (th *Watcher) startWatch() {
	watching := true
	for watching {
		select {
		case event, ok := <-th.pathWatcher.Events:
			if !ok {
				watching = false
				continue
			}
			fileName := event.Name
			operation := event.Op
			// Windows: 对文件夹本身的变化支持很弱
			switch operation {
			case fsnotify.Create:
			case fsnotify.Write:
			case fsnotify.Remove:
			case fsnotify.Rename:
			case fsnotify.Chmod:
			default:
				log.WithFields(logrus.Fields{"fileName": fileName}).Warn("未知的文件变化")
			}
			// 跳过忽略的文件
			if shouldIgnoreFile(fileName) {
				// log.WithFields(logrus.Fields{"fileName": fileName, "operation": operation}).Debug("跳过忽略的文件")
				continue
			}
			// 忽略不需要监听的文件或目录
			if th.isExcluded(fileName) {
				// log.WithFields(logrus.Fields{"fileName": fileName, "operation": operation}).Debug("忽略不需要监听的文件或目录")
				continue
			}
			// 忽略不关注的文件类型(根据文件后缀判断)
			if !th.checkIfWatchExt(fileName) {
				// log.WithFields(logrus.Fields{"fileName": fileName, "operation": operation}).Debug("忽略不关注的文件类型")
				continue
			}
			log.WithFields(logrus.Fields{"fileName": fileName, "operation": operation}).Info("文件变化")
			go th.debounced(th.goBuild)
		case err, ok := <-th.pathWatcher.Errors:
			if !ok {
				return
			}
			log.Warn("监听文件异常", err)
		}
	}
	log.Error("停止监听文件")
	_ = th.pathWatcher.Close()
	os.Exit(2)
}

// 新增监听的文件夹
func (th *Watcher) addWatchPaths(directory string) {
	if !IsDir(directory) {
		return
	}
	fileInfos, err := ioutil.ReadDir(directory)
	if err != nil {
		log.WithFields(logrus.Fields{"directory": directory}).Warn("读取文件夹失败")
	}
	th.paths.Add(directory)
	for _, fileInfo := range fileInfos {
		// 不是目录
		if !fileInfo.IsDir() {
			continue
		}
		fullPath := filepath.Join(directory, fileInfo.Name())
		// 忽略隐藏目录
		if strings.HasPrefix(fileInfo.Name(), ".") {
			log.WithFields(logrus.Fields{"directory": fullPath}).Debug("忽略隐藏目录")
			continue
		}
		// 忽略vendor目录
		if !th.config.Watch.Vendor && fileInfo.Name() == "vendor" {
			log.WithFields(logrus.Fields{"directory": fullPath}).Debug("忽略vendor目录")
			continue
		}
		// 忽略不需要监听的目录
		if th.isExcluded(path.Join(directory, fileInfo.Name())) {
			log.WithFields(logrus.Fields{"directory": fullPath}).Debug("忽略不需要监听的目录")
			continue
		}
		th.addWatchPaths(fullPath)
	}
	return
}

func (th *Watcher) isExcluded(path string) bool {
	for _, excludePath := range th.config.Watch.Exclude {
		excludeFullPath, err := filepath.Abs(excludePath)
		if err != nil {
			log.WithFields(logrus.Fields{"excludePath": excludePath}).Warn("读取绝对路径失败", err)
			continue
		}
		fullPath, err := filepath.Abs(path)
		if err != nil {
			log.WithFields(logrus.Fields{"path": path}).Warn("读取绝对路径失败", err)
			break
		}
		if strings.HasPrefix(fullPath, excludeFullPath) {
			return true
		}
	}
	return false
}

func (th *Watcher) addWatch() {
	for _, wPath := range th.paths.Values() {
		fullPath, ok := wPath.(string)
		if !ok {
			continue
		}
		log.WithFields(logrus.Fields{"path": fullPath}).Info("监听文件夹")
		err := th.pathWatcher.Add(fullPath)
		if err != nil {
			log.WithFields(logrus.Fields{"path": fullPath}).Panic("监听文件夹失败", err)
			os.Exit(2)
		}
	}
}

func (th *Watcher) goBuild() {
	th.buildState.Lock()
	defer th.buildState.Unlock()
	log.WithFields(logrus.Fields{}).Info("开始构建...")
	if err := os.Chdir(th.workPath); err != nil {
		log.WithFields(logrus.Fields{"workPath": th.workPath}).Error("切换工作文件夹失败", err)
		return
	}
	// go build -o {Output/AppName} {Args} -tags {Tags} {Pkg}
	cmdName := "go"
	args := []string{"build"}
	th.outPut = filepath.Join(th.config.Output, th.config.AppName)
	if "windows" == runtime.GOOS && !strings.HasSuffix(th.outPut, ".exe") {
		th.outPut = th.outPut + ".exe"
	}
	args = append(args, "-o", th.outPut)
	args = append(args, th.config.Build.Args...)
	if th.config.Build.Tags != "" {
		args = append(args, "-tags", th.config.Build.Tags)
	}
	var packages []string
	if th.config.Build.Pkg != "" {
		packages = strings.Split(th.config.Build.Pkg, ",")
	}
	args = append(args, packages...)
	buildCmd := exec.Command(cmdName, args...)
	buildCmd.Env = append(os.Environ(), "GOGC=off")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	log.WithFields(logrus.Fields{}).Info("构建: ", cmdName, " ", strings.Join(args, " "))
	startTime := time.Now()
	err := buildCmd.Run()
	if err != nil {
		log.WithFields(logrus.Fields{}).Warn("========================== 构建失败 ==========================")
		log.WithFields(logrus.Fields{}).Warn(err)
		return
	}
	endTime := time.Now()
	log.WithFields(logrus.Fields{}).Info("构建成功! 耗时:", endTime.Sub(startTime))
	if th.config.Run.AutoRun {
		go th.restart()
	}
}

func (th *Watcher) restart() {
	log.WithFields(logrus.Fields{"appName": th.outPut}).Info("重启目标程序")
	th.kill()
	th.start()
}

func (th *Watcher) kill() {
	defer func() {
		if err := recover(); err != nil {
			log.WithFields(logrus.Fields{}).Error("停止失败", err)
		}
	}()
	if th.cmd != nil && th.cmd.Process != nil {
		err := th.cmd.Process.Kill()
		if err != nil {
			log.WithFields(logrus.Fields{}).Error("停止失败", err)
		} else {
			log.WithFields(logrus.Fields{}).Info("停止成功")
		}
	}
}

func (th *Watcher) start() {
	appName := th.outPut
	if !strings.HasPrefix(appName, "./") {
		appName = "./" + appName
	}
	// ./app {Args}
	th.cmd = exec.Command(appName)
	th.cmd.Args = append([]string{}, th.config.Run.Args...)
	th.cmd.Env = append(os.Environ(), th.config.Run.Envs...)
	th.cmd.Stdout = os.Stdout
	th.cmd.Stderr = os.Stderr
	log.WithFields(logrus.Fields{"env": th.config.Run.Envs}).Info("启动: ", appName, " ", strings.Join(th.cmd.Args, " "))
	err := th.cmd.Start()
	if err != nil {
		log.WithFields(logrus.Fields{"appName": appName, "args": th.cmd.Args, "env": th.config.Run.Envs}).Info("启动失败", err)
	} else {
		log.WithFields(logrus.Fields{"appName": appName + "\n\n\n"}).Info("启动成功!")

	}
}

func shouldIgnoreFile(filename string) bool {
	for _, regex := range ignoredFilesRegExps {
		if regex.MatchString(filename) {
			return true
		}
		continue
	}
	return false
}

func (th *Watcher) checkIfWatchExt(fileName string) bool {
	for _, ext := range th.config.Watch.FileExt {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}
