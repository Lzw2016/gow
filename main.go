package main

import (
	"flag"
	"github.com/Lzw2016/gow/gowatch"
	"os"
	"runtime"
	"strings"
)

var (
	// 配置项
	config *gowatch.Config
	// 自定义输出文件夹
	output string
	// 自定义编译 packages
	buildPkg string
	// 自定义运行时附加参数
	runArgs string
	// 显示版本
	showVersion bool
	// 系统退出消息chan
	exit chan bool
)

func init() {
	flag.StringVar(&output, "o", "", "go build输出路径(默认是'./')")
	flag.StringVar(&buildPkg, "p", "", "go build packages")
	flag.StringVar(&runArgs, "args", "", "运行时附加参数。如: -args='-port=8080,-name=demo'")
	flag.BoolVar(&showVersion, "v", false, "显示版本")
}

func main() {
	flag.Parse()
	// 显示版本号
	if showVersion {
		gowatch.PrintVersion()
		os.Exit(0)
	}
	// 初始化 config
	config = gowatch.ParseConfig("")
	if config.AppName == "" {
		config.AppName = gowatch.DefaultConfig.AppName
	}
	if output != "" {
		config.Output = output
	}
	if !gowatch.IsDir(config.Output) {
		config.Output = gowatch.DefaultConfig.Output
	}
	if runArgs != "" {
		config.Run.Args = strings.Split(runArgs, ",")
	}
	if config.Build.Delay < 0 {
		config.Build.Delay = gowatch.DefaultConfig.Build.Delay
	}
	if buildPkg != "" {
		config.Build.Pkg = buildPkg
	}
	if len(config.Watch.FileExt) <= 0 {
		config.Watch.FileExt = gowatch.DefaultConfig.Watch.FileExt
	}
	// 当前工作路径(绝对路径)
	workPath, _ := os.Getwd()
	watcher := gowatch.NewWatcher(workPath, config)
	watcher.StartWatch()

	for {
		select {
		case <-exit:
			runtime.Goexit()
		}
	}
}
