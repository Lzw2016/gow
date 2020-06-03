# gow
Go程序热编译工具，提升开发效率，支持Windows、Linux、Mac

### 特性
- 支持Windows、Linux、Mac
- 默认只监听.go文件，支持自定义监听文件配置
- 支持自定义设置构建参数
- 支持自定义设置运行参数和环境变量
- 支持延时构建(防止构建抖动)

### 安装使用

```shell script
# 安装
go get github.com/silenceper/gowatch

# 已安装更新
go get -u github.com/silenceper/gowatch

# 使用
gow -v
```

### 命令行参数

- -o : 非必须，go build输出路径(默认是'./')
- -p : 非必须，指定需要build的package（也可以是单个文件）
- -args: 非必须，运行时附加参数。如: -args='-port=8080,-name=demo'
- -v: 非必须，显示版本

例子:

```shell script
gow -o ./bin -p ./main/main.go
```

### 通过配置文件使用(推荐)

大部分情况下，不需要更改配置，直接执行gow命令就能满足的大部分的需要，但是也提供了一些配置用于自定义，在执行目录下创建`gowatch.yml`文件:

```yaml
# 执行的app名字，默认是'app.exe'
app-name: app
# 指定output执行的程序路径，默认是'./'
output: ./
# watch相关参数
watch:
  # 需要追加监听的文件后缀名字，默认是'.go'，
  ext:
    - .go
  # 需要追加监听的目录，默认是'./'
  paths:
    - ../gowatch
    - ../main
  # vendor目录下的文件是否也监听，默认是'false'
  vendor: true
  # 不需要监听的目录
  exclude:
    - main.exe
# build相关参数
build:
  # 构建延时(单位:毫秒),默认是'5000ms'
  delay: 3000
  # 构建时的额外参数
  args:
    - key=value
  # 需要编译的包或文件,多个使用','分隔
  pkg: main/main.go
  # 在go build 时期接收的-tags参数
  tags:
# run相关参数
run:
  # build完成后是否自动运行，默认是'true'
  auto-run: true
  # 执行时的额外参数
  args:
    - port=8080
    - name=demo
  # 执行时追加的环境变量
  envs:
    - k=v
```
