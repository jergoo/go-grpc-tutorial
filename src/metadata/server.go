package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/jergoo/go-grpc-tutorial/protos/ping" // 引入编译生成的包
)

// PingPongServer 实现 pb.PingPongServer 接口
type PingPongServer struct {
	pb.UnimplementedPingPongServer // 兼容性需要，避免未实现server接口全部方法
}

// Ping 单次请求-响应模式
func (s *PingPongServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	// 读取请求metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("Got md: %v", md)
	}

	// 设置响应metadata
	grpc.SetHeader(ctx, metadata.New(map[string]string{"rkey": "rval"}))

	return &pb.PongResponse{Value: "pong"}, nil
}

// MultiPong 服务端流模式
func (s *PingPongServer) MultiPong(req *pb.PingRequest, stream pb.PingPong_MultiPongServer) error {
	// 读取请求metadata
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Printf("Got md: %v", md)
	}

	// 设置响应metadata
	err := stream.SendHeader(metadata.New(map[string]string{"rkey": "rval"}))
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		data := &pb.PongResponse{Value: "pong"}
		// 发送消息
		err := stream.Send(data)
		if err != nil {
			return err
		}
		if err != nil {
			log.Println(err)
		}
	}
	return nil
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
