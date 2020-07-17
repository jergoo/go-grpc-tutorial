# Hello gRPC

按照惯例，这里也从一个Hello项目开始，本项目定义了一个Hello Service，客户端发送包含字符串名字的请求，服务端返回Hello消息。

**流程：** 

1. 编写`.proto`描述文件
2. 编译生成`.pb.go`文件
3. 服务端实现约定的接口并提供服务
4. 客户端按照约定调用`.pb.go`文件中的方法请求服务

**项目结构：**

```
|—— hello/
	|—— client/
    	|—— main.go   // 客户端
	|—— server/
    	|—— main.go   // 服务端
|—— proto/
	|—— hello/
		|—— hello.proto   // proto描述文件
		|—— hello.pb.go   // proto编译后文件
```


**Step1：编写描述文件：hello.proto**

```protobuf
syntax = "proto3"; // 指定proto版本
package hello;     // 指定默认包名

// 指定golang包名
option go_package = "hello";

// 定义Hello服务
service Hello {
	// 定义SayHello方法
	rpc SayHello(HelloRequest) returns (HelloResponse) {}
}

// HelloRequest 请求结构
message HelloRequest {
	string name = 1;
}

// HelloResponse 响应结构
message HelloResponse {
    string message = 1;
}
```

`hello.proto`文件中定义了一个Hello Service，该服务包含一个`SayHello`方法，同时声明了`HelloRequest`和`HelloResponse`消息结构用于请求和响应。客户端使用`HelloRequest`参数调用`SayHello`方法请求服务端，服务端响应`HelloResponse`消息。一个最简单的服务就定义好了。


**Step2：编译生成`.pb.go`文件**

```sh
$ cd proto/hello

# 编译hello.proto
$ protoc -I . --go_out=plugins=grpc:. ./hello.proto
```
在当前目录内生成的`hello.pb.go`文件，按照`.proto`文件中的说明，包含服务端接口`HelloServer`描述，客户端接口及实现`HelloClient`，及`HelloRequest`、`HelloResponse`结构体。

> 注意：不要手动编辑该文件


**Step3：实现服务端接口 `server/main.go`**

```go
package main

import (
	"fmt"
	"net"

	pb "github.com/jergoo/go-grpc-example/proto/hello" // 引入编译生成的包
	"golang.org/x/net/context"
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
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	// 实例化grpc Server
	s := grpc.NewServer()

	// 注册HelloService
	pb.RegisterHelloServer(s, HelloService)

	grpclog.Println("Listen on " + Address)
	s.Serve(listen)
}

```

服务端引入编译后的`proto`包，定义一个空结构用于实现约定的接口，接口描述可以查看`hello.pb.go`文件中的`HelloServer`接口描述。实例化grpc Server并注册HelloService，开始提供服务。

运行：

```sh
$ go run main.go
Listen on 127.0.0.1:50052  //服务端已开启并监听50052端口
```

**Step4：实现客户端调用 `client/main.go`**

```go
package main

import (
	pb "github.com/jergoo/go-grpc-example/proto/hello" // 引入proto包
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"
)

func main() {
	// 连接
	conn, err := grpc.Dial(Address, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalln(err)
	}
	defer conn.Close()

	// 初始化客户端
	c := pb.NewHelloClient(conn)

	// 调用方法
	req := &pb.HelloRequest{Name: "gRPC"}
	res, err := c.SayHello(context.Background(), req)

	if err != nil {
		grpclog.Fatalln(err)
	}

	grpclog.Println(res.Message)
}
```

客户端初始化连接后直接调用`hello.pb.go`中实现的`SayHello`方法，即可向服务端发起请求，使用姿势就像调用本地方法一样。


运行：

```sh
$ go run main.go
Hello gRPC.    // 接收到服务端响应
```

如果你收到了"Hello gRPC"的回复，恭喜你已经会使用github.com/jergoo/go-grpc-example/proto/hello了。

> 建议到这里仔细看一看hello.pb.go文件中的内容，对比hello.proto文件，理解protobuf中的定义转换为golang后的结构。
