package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/jergoo/go-grpc-tutorial/protos/ping" // 引入编译生成的包
)

// Ping 单次请求-响应模式
func Ping() {
	// 读取服务端证书，并制定对应服务名
	creds, err := credentials.NewClientTLSFromFile("keys/server.crt", "grpc.server")
	if err != nil {
		log.Fatalf("load crt fail: %v", err)
	}

	// 连接配置
	opts := []grpc.DialOption{
		// TLS 认证
		grpc.WithTransportCredentials(creds),
		// Token 认证
		grpc.WithPerRPCCredentials(CustomAuth{Token: "1234567890"}),
	}
	conn, err := grpc.Dial("localhost:1234", opts...)
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

// CustomAuth 自定义认证类型
type CustomAuth struct {
	Token string
}

// GetRequestMetadata 生成认证信息
func (a CustomAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": a.Token,
	}, nil
}

// RequireTransportSecurity 是否开启 TLS
func (a CustomAuth) RequireTransportSecurity() bool {
	return true
}
