# Hello gRPC

这里从一个简单的例子说明 gRPC 的基本使用流程，建议边看边实践。

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
## 实践

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

### Step1：编写服务描述文件

```protobuf
// src/protos/ping/ping.proto
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

定义了一个名为`PingPong`的 service，包含一个`Ping`方法，同时声明了`PingRequest`和`PongResponse`消息结构用于请求和响应。客户端使用`PingRequest`参数调用`Ping`方法请求服务端，服务端响应`PongResponse`消息，一个最简单的服务就定义好了。

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

服务端引入编译生成的包，定义一个`PingPongServer`用于实现约定的接口，接口描述可以查看`ping_grpc.pb.go`文件中的`PingPongServer`接口描述。实例化grpc Server并注册PingPongServer开始提供服务。

### Step4：客户端调用

```go
// src/ping/client.go
package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/jergoo/go-grpc-tutorial/protos/ping" // 引入编译生成的包
)

// Ping 单次请求-响应模式
func Ping() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 实例化客户端并调用
	client := pb.NewPingPongClient(conn)
	res, err := client.Ping(context.Background(), &pb.PingRequest{Value: "ping"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Value)
}
```

客户端初始化连接后直接调用`Ping`方法，即可向服务端发起请求并获取响应，就像调用本地方法一样。

## 总结

到此为止一个简单的gRPC服务就完成了，基本流程如下：

1. 使用protobuf编写服务描述文件，定义消息结构和服务接口
2. 编译proto文件，生成服务端和客户端接口代码
3. 服务端实现`.pb.go`文件中的server接口
4. 客户端直接使用`.pb.go`文件中的client调用接口
