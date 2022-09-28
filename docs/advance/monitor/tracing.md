# tracing

---

grpc内置了客户端和服务端的请求追踪，基于`golang.org/x/net/trace`包实现，默认是开启状态，可以查看事件和请求日志，对于基本的请求状态查看调试也是很有帮助的，客户端与服务端基本一致，这里以服务端开启trace server为例，修改hello项目服务端代码：

## 目录结构

```
|—— hello_trace/
	|—— client/
    	|—— main.go   // 客户端
	|—— server/
    	|—— main.go   // 服务端
|—— proto/
	|—— hello/
		|—— hello.proto   // proto描述文件
		|—— hello.pb.go   // proto编译后文件
```

## 示例代码

```go
package main

import (
	"fmt"
	"net"
	"net/http"

	pb "github.com/jergoo/go-grpc-tutorial/proto/hello" // 引入编译生成的包

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"
)

// 定义helloService并实现约定的接口
type helloService struct{}

// HelloService Hello服务
var HelloService = helloService{}

// SayHello 实现Hello服务接口
func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s.", in.Name)

	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	// 实例化grpc Server
	s := grpc.NewServer()

	// 注册HelloService
	pb.RegisterHelloServer(s, HelloService)

	// 开启trace
	go startTrace()

	grpclog.Println("Listen on " + Address)
	s.Serve(listen)
}

func startTrace() {
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
	go http.ListenAndServe(":50051", nil)
	grpclog.Println("Trace listen on 50051")
}
```
这里我们开启一个http服务监听50051端口，用来查看grpc请求的trace信息

运行：

```sh
$ go run main.go

Listen on 127.0.0.1:50052                                                       
Trace listen on 50051

# 进入client目录执行一次客户端请求     
```


## 服务端事件查看

访问：localhost:50051/debug/events，结果如图：

![](../_media/grpc_trace_events.jpg)

可以看到服务端注册的服务和服务正常启动的事件信息。


## 请求日志信息查看

访问：localhost:50051/debug/requests，结果如图：

![](../_media/grpc_trace_requests.jpg)

这里可以显示最近的请求状态，包括请求的服务、参数、耗时、响应，对于简单的状态查看还是很方便的，默认值显示最近10条记录。
