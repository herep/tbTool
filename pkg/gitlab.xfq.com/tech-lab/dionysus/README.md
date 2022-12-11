# README

# Dionysus

## 项目简介

本项目是基于 [cobra](https://github.com/spf13/cobra) 开发的，目的是让需要 golang 开发的业务可以快速起飞。

项目分为两部分：

- cmd:  程序运行模式，现有 `gin` `ctl` `healthy` 三种模式
- pkg: 各种运行模式中，公用的外部组件或者工具集合

**注意：提交代码，需要本地 `make lint` 通过， 因为在push的时候，gitlab的ci不通过将会MR无法合并。**

## 快速起步

新建一个 go 文件，并引入 dionysus

```go
import "gitlab.xfq.com/tech-lab/dionysus/cmd/gincmd"

func main() {
	gc := gincmd.New()
    dionysus.Start("example", gc)
}
```

使用 `go run` 命令启动, 就可以看到 dionysus 支持的所有命令了，相关命令的设置请参照下面对应 cmd说明

```bash
$ go run YourFile.go

Usage:
  example [command]

Available Commands:
  ctl         Run as ctl mod
  gin         Run as gin web server
  healthy     Check service healthy status
  help        Help about any command

Flags:
  -c, --config string      config path
  -e, --endpoints string   the etcd endpoints
  -h, --help               help for example
  -l, --log string         log file path; default console output
  -n, --name string        the project name

Use "example [command] --help" for more information about a command.
```

## cmd 说明

### 1. gincmd

gincmd 是对 [gin 框架](https://gin-gonic.com/) 的封装，这里按照我们现有 IT 结构进行了简单封装

快速起步：

业务启动前的初始化动作，调用`RegPreRunFunc()`来实现。退出前的清理动作，调用 `RegPostRunFunc()`

调用`RegPreRunFunc(value string, priority cmd.Priority, f func() error) error`。参数说明，value为该step的名字；priority为优先级，值越小，优先级越高。 fn为需要执行的初始化函数。注意fn初始化函数失败，整个程序启动会失败。


整个dionysus的启动流程如下： 首先执行 RegPreRunFunc 里面注册的初始化动作，依次注册logger,metrics,RateLimit,Timeout的中间件，然后注册用户配置的全局中间件。 然后依次进行注册的每个group的初始化，依次进行group的局部中间件，head路由，get路由，post路由的注册。

```go
func main() {
    gc := gincmd.New()
    gc.RegGroups(RegisterDemoRouter())
    dionysus.Start("example", gc)
}

func RegisterDemoRouter() *gincmd.CustomGroup {
	// 建立路由组
	demogroup := gincmd.NewGroup("/demogroup")

	// (可选)注册群组中间件，所有群组下的路由规则在执行handler之前，都会首先执行middler中间件
	demogroup.RegMiddle(demoMiddle)

	// 目前只支持 GET, POST和HEAD三种

	// 在该路由下进行一些设置header的操作
	demogroup.RegRouterHead("/demohead", demoHead)

	// 在该路由群组下绑定路由，hander的默认超时时间是5秒，如果超时的话会返回504错误。
	// hander的超时时间可以通过环境变量GAPI_REQUEST_TIMEOUT配置，uint类型，单位是秒。
	demogroup.RegRouterGet("/demoroute", demoGetHandler)
    return demogroup
}

func demoMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("this middle")
		c.Next()
	}
}

func demoHead(c *gin.Context) pkg.Render {
	c.Header("contentType", "application/json")
	// head的http method可以直接返回nil。
	return nil
}

func demoGetHandler(c *gin.Context) pkg.Render {
	name := conf.Get("test_config.child_config.name")
	logger.Info("this is info log")
	// 非head的method如果直接返回 nil的话，那么http会报500错误
	return pkg.String{
		Format: "hola %s %d",
		Data:   []interface{}{name, 2},
	}
}
```

可以通过如下命令运行和测试demo, 首先启动 gin 服务

```bash
$ go run example/example.go gin
```

再打开一个 shell

```bash
$ curl  http://127.0.0.1:8080/demogroup/user\?user_id\=1002

{"code":20000,"data":{"user_name":"小强"},"msg":"OK"}
```

### 2. ctlcmd

ctl cmd 是对有生命周期的命令行程序的抽象封装。其生命周期和执行流程设计如下：

```bash
prerun stage                     run stage                       post run stage
+-----------------+              +-------------------+        +--------------------------------------------------+
|                 |              |                   |        |                                                  |
| +-------------+ |              |  +--------------+ |        |  +-----------------+                             |
| | parse flag  | |              |  |              +------------>+ stuck at select +-----------------+           |
| +-----+-------+ |              |  |              | |        |  +--------+--------+                 |           |
|       |         |       +-------->+ go (runFunc) | |        |           +                          +           |
|       |         |       |      |  |              | |        |        os.Signal                user finish      |
|  flags|in ctx   |       |      |  |              | |        |           +                          |           |
|       |         |       |      |  +--------------+ |        |           |                          |           |
|       |         |       |      |                   |        |           v                          v           |
|       v         |       |      |                   |        |  +--------+--------+  succeed  +-----+--------+  |
| +-----+-------+ |       |      |                   |        |  |  shutdownFunc   +---------->+  postFunc    |  |
| |  preFunc    | |       |      |                   |        |  +--------+--------+           +-----+--------+  |
| +-----+-------+ |       |      |                   |        |           |                          |           |
|       |         |       |      |                   |        |           |                          |code=0     |
|       v         |       |      |                   |        |           |                          v           |
| +-----+-------+ |       |      |                   |        |           |  timeout code=1    +-----+--------+  |
| | go (healthy)+---------+      |                   |        |           +------------------->+ os.Exit(code)|  |
| +-------------+ |              |                   |        |                                +--------------+  |
|                 |              |                   |        |                                                  |
+-----------------+              +-------------------+        +--------------------------------------------------+
```

快速开始：

```bash
func main() {

	activate.RegPreFunc(func(ctx context.Context) error {
		fmt.Println("1") // 运行前的初始化
		return nil
	})

	activate.RegRunFunc(func(ctx context.Context) {
		fmt.Println("2") // 主逻辑
	})

	activate.RegShutdownFunc(func(ctx context.Context) {
		fmt.Println("3") // 接到系统信号时的打断逻辑
	})

	activate.RegPostFunc(func(ctx context.Context) error {
		fmt.Println("4") // 退出时的逻辑
		return nil
	})

	activate.StartEngine(activate.WithCmdUse("dioCtl"))
}
```

运行

```bash
$ go run example/ctl/main.go ctl

1
2
4
```

这里输出了 `1 2 4` ，**只有程序在运行中，被系统信号打断**才会触发 shutdown, 输出 `3` ，详细请参照上面的流程图

## pkg 说明
pkg 是用于不同业务之间，可以支持快速开发的技术组件，例如 `redis` `orm`。开发新组件请参照 [PKG开发规范](pkg/README.md)

### conf

-c, –config string, 或者用环境变量GAPI_CONFIG配置。两种方式同时配置时，环境变量配置生效。指定读取配置文件的路径，请使用绝对路径+文件名。

### logger

-l, –log string 指定日志输出路径，默认输出到终端。或者用环境变量GAPI_LOG配置。两种方式同时配置时，环境变量配置生效。 日志格式统一处理，业务开发务必使用gitlab.xfq.com/tech-lab/dionysus/pkg/logger 包。 Fatal、Panic级别会导致exit，当前不暴露该方法

### addr

-a, –addr string 指定http server的启动地址，默认:8080。或者用环境变量GAPI_ADDR配置。两种方式同时配置时，环境变量配置生效。

### middle timeout

注意hander的默认超时时间是10秒，可以通过环境变量GAPI_REQUEST_TIMEOUT或http的header Request_Timeout进行配置，其中http的header Request_Timeout只有在小于环境变量GAPI_REQUEST_TIMEOUT（当前默认值10s）的值时才生效。并且优先级顺序为default < env < request_header。

### grpool

如果相关需要创建goroutine，请使用此包会进行自动recover。可以直接调用`grpool.Submit`使用默认的pool来运行你的func (默认共享5万个goroutine池)。也可以自己调用可以自己调用`pool := grpool.NewPool, pool.Submit`来运行你的func。

### metrics

服务默认启用metrics，默认为:9120. 可通过环境变量GAPI_METRICS_ADDR或者flag -m, —metricsAddr指定metrics地址。

当指定的gapi地址与metrics地址相同时，处于安全考虑，此时将不启动metrics服务。


## 测试镜像制作说明
该镜像的Dockerfile位于 https://gitlab.xfq.com/tech-lab/micro-gen/blob/gapi/Dockerfile gapi分支。比如我需要增加安装redis,etcd组件时，
则需要在Dockerfile中增加redis,etcd的安装，其它组件可根据需要自行安装。
```bash
RUN echo 'http://mirrors.aliyun.com/alpine/v3.9/main' > /etc/apk/repositories \
    && apk add --no-cache bash protobuf curl git python3 openssh make gcc libc-dev redis\
    && rm -rf /var/cache/apk/* /tmp/*

# etcd
RUN export ETCD_TAR=etcd-v3.4.7-linux-amd64.tar.gz \
    && curl -OL https://github.com/etcd-io/etcd/releases/download/v3.4.7/$ETCD_TAR \
    && mkdir -p /tmp/etcd-download-test \
    && tar xzvf $ETCD_TAR -C /tmp/etcd-download-test --strip-components=1 \
    && mv /tmp/etcd-download-test/etcd /usr/local/bin/ \
    && mv /tmp/etcd-download-test/etcdctl /usr/local/bin/ \
    && rm -rf /tmp/etcd-download-test/ \
    && rm -f $ETCD_TAR
```
然后重新构建镜像：
```bash
docker build -t reg.weipaitang.com/micro/golangci .
```
构建完成后将镜像推送到镜像仓库中
```bash
docker push reg.weipaitang.com/micro/golangci
```
在镜像里将redis,etcd组件做进镜像后，在 https://gitlab.xfq.com/tech-lab/dionysus/blob/master/hack/entry.sh 脚本中增加启动步骤：
```bash
redis-server /etc/redis.conf >/dev/null 2>&1 &
etcd >/dev/null 2>&1 &
```


