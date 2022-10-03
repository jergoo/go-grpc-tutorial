package main

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	pb "github.com/jergoo/go-grpc-tutorial/protos/ping" // 引入编译生成的包
)

// PingPongServer 实现 pb.PingPongServer 接口
type PingPongServer struct {
	pb.UnimplementedPingPongServer // 兼容性需要，避免未实现server接口全部方法
}

// Ping 单次请求-响应模式
func (s *PingPongServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	return &pb.PongResponse{Value: "pong"}, nil
}

// 服务端拦截器 - token 认证
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("authorization missing")
	}

	var token string
	if auth, ok := md["authorization"]; ok {
		token = auth[0]
	}
	if token != "1234567890" {
		return nil, grpc.Errorf(codes.Unauthenticated, "token invalid")
	}

	// 处理请求
	return handler(ctx, req)
}

// 启动server
func main() {
	// 加载证书文件
	creds, err := credentials.NewServerTLSFromFile("keys/server.crt", "keys/server.key")
	if err != nil {
		log.Fatalf("load crt fail:%v", err)
	}
	opts := []grpc.ServerOption{
		grpc.Creds(creds),                      // TLS
		grpc.UnaryInterceptor(authInterceptor), // Token
	}

	srv := grpc.NewServer(opts...)
	// 注册 PingPongServer
	pb.RegisterPingPongServer(srv, &PingPongServer{})
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listen on 1234")
	srv.Serve(lis)
}
