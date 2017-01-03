package main

import (
	pb "git.vodjk.com/go-grpc/example/proto" // 引入proto包

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials" // 引入grpc认证包
	"google.golang.org/grpc/grpclog"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"
)

func main() {
	// TLS连接
	creds, err := credentials.NewClientTLSFromFile("../../keys/server.pem", "server name")
	if err != nil {
		grpclog.Fatalf("Failed to create TLS credentials %v", err)
	}
	conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(creds))

	if err != nil {
		grpclog.Fatalln(err)
	}

	defer conn.Close()

	// 初始化客户端
	c := pb.NewHelloClient(conn)

	// 调用方法
	reqBody := new(pb.HelloRequest)
	reqBody.Name = "gRPC"
	r, err := c.SayHello(context.Background(), reqBody)

	if err != nil {
		grpclog.Fatalln(err)
	}

	grpclog.Println(r.Message)
}
