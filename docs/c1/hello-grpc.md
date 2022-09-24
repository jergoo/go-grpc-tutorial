# Hello gRPC

这里从一个最简单的例子说明 gRPC 的基本使用流程，建议边看边实践。

## 环境准备

### Protobuf 编译器

项目地址：[google/protobuf](https://github.com/google/protobuf)

这里使用`brew`工具安装

```sh
$ brew install protobuf
```

执行`protoc`命令查看当前版本：

```sh
$ protoc --version
libprotoc 3.21.6
```
### Go 编译插件

项目地址：
* [protobuf-go](https://github.com/protocolbuffers/protobuf-go)
* [grpc-go](https://github.com/grpc/grpc-go)

安装：
```sh
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

$ export PATH="$PATH:$(go env GOPATH)/bin"
```

## 实现

实现一个PingPong Service，客户端发送 ping 请求，服务端返回 pong 响应。

**项目结构：**

```
|—— src/ping/
	|—— client.go // 客户端
	|—— server.go // 服务端
|—— src/protos/
	|—— ping/
		|—— ping.proto   // proto描述文件
		|—— ping.pb.go   // proto编译生成
		|-- ping_grpc.pb.go // proto编译生成
```


### Step1：编写接口描述文件：ping.proto

```protobuf
syntax = "proto3"; // 指定proto版本
package ping;     // 指定包名

// 指定go包名
option go_package = "protos/ping";

// 定义PingPong服务
service PingPong {
	rpc Ping(PingRequest) returns (PongResponse);
}

// PingRequest 请求结构
message PingRequest {
	string value = 1;
}

// PongResponse 响应结构
message PongResponse {
    string value = 1;
}
```

这里大概看一下`ping.proto`文件的内容，里面定义了一个名为`PingPong`的 service，包含一个`Ping`方法，同时声明了`PingRequest`和`PongResponse`消息结构用于请求和响应。客户端使用`PingRequest`参数调用`Ping`方法请求服务端，服务端响应`PongResponse`消息，一个最简单的服务就定义好了。


## Step2：编译proto文件

```sh
$ cd src
$ protoc --go_out=. --go-grpc_out=. protos/ping/ping.proto
```
在src目录执行编译命令，会在目录`src/protos/ping`内生成两个文件`ping.pb.go`和`ping_grpc.pb.go`。可以大概看一下这两个文件的内容，`ping.pb.go` 包含了之前定义的两个message相关的结构，`ping_grpc.pb.go`包含了定义的service相关的客户端和服务端接口，**注意不要手动修改这两个文件**。

### Step3：实现服务端接口

```go
// src/ping/server.go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/jergoo/go-grpc-tutorial/protos/ping" // 引入编译生成的包
)

// PingPongServer 实现 pb.PingPongServer 接口
type PingPongServer struct {
	pb.UnsafePingPongServer
}

// Ping 单次请求-响应模式
func (s *PingPongServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	return &pb.PongResponse{Value: "pong"}, nil
}

// 启动server
func main() {
	srv := grpc.NewServer()
	// 注册 PingPongServer
	pb.RegisterPingPongServer(srv, &PingPongServer{})
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listen on 1234")
	srv.Serve(lis)
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
	pb "github.com/jergoo/go-grpc-tutorial/proto/hello" // 引入proto包
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

如果你收到了"Hello gRPC"的回复，恭喜你已经会使用github.com/jergoo/go-grpc-tutorial/proto/hello了。

> 建议到这里仔细看一看hello.pb.go文件中的内容，对比hello.proto文件，理解protobuf中的定义转换为golang后的结构。
