# Hello gRPC

从一个简单的示例说明 go gRPC 的基本使用流程，建议边看边操作。实现一个 PingPong 服务，客户端发送 ping 请求，服务端返回 pong 响应。

**项目结构：**

```
|—— src/ping/
	|—— client.go // 客户端
	|—— server.go // 服务端
|—— src/protos/
	|—— ping/
		|—— ping.proto   // protobuf描述文件
		|—— ping.pb.go   // protoc编译生成
		|-- ping_grpc.pb.go // protoc编译生成
```

## 第1步：编写protobuf描述文件

```protobuf
// src/protos/ping/ping.proto
syntax = "proto3"; // 指定proto版本
package ping;     // 指定包名

// 指定go包路径
option go_package = "protos/ping";

// 定义PingPong服务
service PingPong {
	// Ping 发送 ping 请求，接收 pong 响应
    rpc Ping(PingRequest) returns (PongResponse);
}

// PingRequest 请求结构
message PingRequest {
    string value = 1; // value字段为string类型
}

// PongResponse 响应结构
message PongResponse {
    string value = 1; // value字段为string类型
}
```

定义了一个名为`PingPong`的 service，包含一个`Ping`方法，同时声明了`PingRequest`和`PongResponse`消息结构用于请求和响应。客户端使用`PingRequest`参数调用`Ping`方法请求服务端，服务端响应`PongResponse`消息，一个基本的服务就定义好了。

## 第2步：编译protobuf文件

```sh
$ cd src
$ protoc --go_out=. --go-grpc_out=. protos/ping/ping.proto
```
在src目录执行编译命令，会在目录`src/protos/ping`内生成两个文件`ping.pb.go`和`ping_grpc.pb.go`。可以大概看一下这两个文件的内容，`ping.pb.go` 包含了之前定义的两个message相关的结构，`ping_grpc.pb.go`包含了定义的service相关的客户端和服务端接口，**不要修改这两个文件的内容**。

## 第3步：实现服务端接口

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
	pb.UnimplementedPingPongServer
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

服务端引入编译生成的包，定义一个`PingPongServer`用于实现约定的接口，接口描述可以查看`ping_grpc.pb.go`文件中的`PingPongServer`接口。实例化grpc Server并注册PingPongServer开始提供服务。

## 第4步：客户端调用

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
	log.Println(res.Value)
}
```

客户端初始化连接，使用`ping_grpc.pb.go`中的`PingPongClient`实例调用`Ping`方法，即可向服务端发起请求并获取响应，就像调用本地方法一样。

---

> 以上就是一个最基础的gRPC服务，使用非常简单，底层网络细节全部由gRPC处理，开发者只需要关注接口设计，基本流程如下：
> 1. 编写protobuf描述文件，定义消息结构和服务接口
> 2. 编译protobuf文件，生成服务端和客户端接口代码
> 3. 实现`*_grpc.pb.go`文件中描述的服务端接口
> 4. 使用`*_grpc.pb.go`文件中的client调用服务
