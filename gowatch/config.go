package gowatch

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

var configFile = "./gowatch.yml"

type WatchCfg struct {
	// 需要追加监听的文件后缀名字，默认是'.go'，
	FileExt []string `yaml:"ext"`
	// 需要追加监听的目录，默认是'./'
	Paths []string `yaml:"paths"`
	// vendor目录下的文件是否也监听，默认是'false'
	Vendor bool `yaml:"vendor"`
	// 不需要监听的目录
	Exclude []string `yaml:"exclude"`
}

type BuildCfg struct {
	// 构建延时(单位:毫秒),默认是'5000ms'
	Delay int64 `yaml:"delay"`
	// 构建时的额外参数
	Args []string `yaml:"args"`
	// 需要编译的包或文件,多个使用','分隔
	Pkg string `yaml:"pkg"`
	// 在go build 时期接收的-tags参数
	Tags string `yaml:"tags"`
}

type RunCfg struct {
	// build完成后是否自动运行，默认是'true'
	AutoRun bool `yaml:"auto-run"`
	// 执行时的额外参数
	Args []string `yaml:"args"`
	// 执行时追加的环境变量
	Envs []string `yaml:"envs"`
}

type Config struct {
	// 执行的app名字，默认是'app.exe'
	AppName string `yaml:"app-name"`
	// 指定output执行的程序路径，默认是'./'
	Output string `yaml:"output"`
	// watch相关参数
	Watch WatchCfg
	// build相关参数
	Build BuildCfg
	// run相关参数
	Run RunCfg
}

// 默认值
var DefaultConfig = Config{
	AppName: "app",
	Output:  "./",
	Watch: WatchCfg{
		FileExt: []string{".go"},
		Paths:   []string{},
		Vendor:  false,
		Exclude: []string{},
	},
	Build: BuildCfg{
		Delay: 5000,
		Args:  []string{},
	},
	Run: RunCfg{
		AutoRun: true,
		Args:    []string{},
		Envs:    []string{},
	},
}

// Yml解析默认值
func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawConfig Config
	defaultValue := rawConfig(DefaultConfig)
	if err := unmarshal(&defaultValue); err != nil {
		return err
	}
	*s = Config(defaultValue)
	return nil
}

// 解析配置文件
func ParseConfig(filePath string) *Config {
	if filePath == "" || !IsFile(filePath) {
		filePath = configFile
	}
	config := &Config{}
	filename, _ := filepath.Abs(filePath)
	if !IsFile(filename) {
		log.WithFields(logrus.Fields{"filename": filename}).Info("配置文件不存在,使用默认配置")
		return config
	}
	log.WithFields(logrus.Fields{"filename": filename}).Info("成功读取配置文件")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic("读取配置文件失败：", err)
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Panic("Yaml配置文件读取失败：", err)
	}
	return config
}
