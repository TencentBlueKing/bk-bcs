# blog使用glog改造后的使用说明

## 1. 特性说明

1. 兼容原有的blog日志管理包中的方法，可无缝迁移。
2. 特性变更：CloseLogs函数不再是关闭日志打印，而是将缓存的所有日志刷到对应日志文件中，防止异常退出导致的日志丢失。
3. 特性变更：目前的blog不再有日志翻转功能，日志翻转需要借助第三方工具实现。
4. 新增特性：
   - 日志分级：可利用函数blog.Level进行日志分级，日志类型可为Info, Infof, Infoln这三种，日志级别层级不限制。最终打印的日志级别由启动参数-v来决定，如-v=2，则日志级别小于等于2的均会显示，大于2的即3以上的均不会显示，方便我们自主管理打印的日志级别。
   - 新增格式化日志打印支持，Infof, Warnf, Errorf, Fatalf。可方便、灵活定义日志格式。
   - 新增强制进程退出日志打印函数Fatal与 Fatalf两个函数，可以在进程启动过程检测到不可恢复的异常时强制进程退出。


5. 我们使用的glog在runtime时是不支持debug日志打印功能，这个功能是合并在日志分级功能中的。为了兼容blog原有的Debug功能，我们用自定义的日志级别3级来打印debug日志。这样如果要看debug日志信息，需要在启动时指定--v参数的值>=3。

## 2. 新增的启动标志

​	blog新增以下启动参数，每个参数的意义见具体的解释。

```
  -alsologtostderr
    	log to standard error as well as files
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging

```

## 3. 日志格式说明

```go
I0504 11:10:43.943151   19171 main.go:13] Hi, you have received a message type:Info from breeze.
```

以上面的一条日志信息为例。这条日志分为5部分，分别为:

1. 日志头： I0504, 其中“I”表示这是一条info日志，0504为当前的日期。
2. 日志打印时间： “11:10:43.943151” 为打印这条日志的具体时间，精确到ns。
3. 当前的进程ID： “19171” 为当前进程的ID号。
4. 打印日志所在的文件和行号： “main.go:13” 表示该条日志为main.go文件第13行打印。
5. 具体的日志信息： “Hi, you have received a message type:Info from breeze.”



## 4. 使用示例

```go
package main

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"flag"
)

func main()  {
	flag.Parse()
    // InitLogs() is need.
	blog.InitLogs()
    // CloseLogs() can assure you that you can not lose any log.
	defer blog.CloseLogs()

	blog.Info("Hi, you have received a message type:Info from breeze.")
	blog.Infof("Hi, you have received a message type:%s from %s.", "Infof", "breeze")

	blog.Warn("Hi, you have received a message type:Warn from breeze.")
	blog.Warnf("Hi, you have received a message type:%s from %s.", "Warnf", "breeze")

	blog.Error("Hi, you have received a message type:Error from breeze.")
	blog.Errorf("Hi, you have received a message type:%s from %s.", "Errorf", "breeze")

	blog.Debug("Hi, you have received a message type:Debug from breeze.")

	blog.Level(1).Info("Hi, this is a self-defined log level: 1 from breeze.")
	blog.Level(2).Infof("Hi, this is a self-defined log level: %d from %s.", 2, "breeze")


	blog.Fatal("what? process exit.")
	blog.Info("can not run here.")
}

```

​	编译这个demo，使用`./main`启动时，输出的日志信息如下，从日志中可以看出默认情况下-v参数为0，不显示自定义的分级日志和Debug信息。

```shell
[root@host1 /data/gopath/src]# ./main
I0504 11:10:05.200126   19086 main.go:13] Hi, you have received a message type:Info from breeze.
I0504 11:10:05.200178   19086 main.go:14] Hi, you have received a message type:Infof from breeze.
W0504 11:10:05.200188   19086 main.go:16] Hi, you have received a message type:Warn from breeze.
W0504 11:10:05.200194   19086 main.go:17] Hi, you have received a message type:Warnf from breeze.
E0504 11:10:05.200202   19086 main.go:19] Hi, you have received a message type:Error from breeze.
E0504 11:10:05.200210   19086 main.go:20] Hi, you have received a message type:Errorf from breeze.
F0504 11:10:05.200219   19086 main.go:28] what? process exit.

```

​	指定启动参数`./main -v=3`, 再次执行，输出的日志内容如下，可以看到debug信息与自定义的分级日志均显示。

```shell
[root@host1 /data/gopath/src]# ./main --v=3
I0504 11:50:24.779726   30944 main.go:13] Hi, you have received a message type:Info from breeze.
I0504 11:50:24.779777   30944 main.go:14] Hi, you have received a message type:Infof from breeze.
W0504 11:50:24.779785   30944 main.go:16] Hi, you have received a message type:Warn from breeze.
W0504 11:50:24.779791   30944 main.go:17] Hi, you have received a message type:Warnf from breeze.
E0504 11:50:24.779798   30944 main.go:19] Hi, you have received a message type:Error from breeze.
E0504 11:50:24.779808   30944 main.go:20] Hi, you have received a message type:Errorf from breeze.
I0504 11:50:24.779817   30944 blog.go:69] [Hi, you have received a message type:Debug from breeze.]
I0504 11:50:24.779833   30944 main.go:24] Hi, this is a self-defined log level: 1 from breeze.
I0504 11:50:24.779841   30944 main.go:25] Hi, this is a self-defined log level: 2 from breeze.
F0504 11:50:24.779849   30944 main.go:28] what? process exit.
```

​	另外，以上两次运行最后一条日志`blog.Info("can not run here.")`均未执行，这是因为blog.Fatal()使进程退出了。