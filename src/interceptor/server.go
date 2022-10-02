package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"

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

// MultiPong 服务端流模式
func (s *PingPongServer) MultiPong(req *pb.PingRequest, stream pb.PingPong_MultiPongServer) error {
	for i := 0; i < 10; i++ {
		data := &pb.PongResponse{Value: "pong"}
		// 发送消息
		err := stream.Send(data)
		if err != nil {
			return err
		}
	}
	return nil
}

// MultiPing 客户端流模式
func (s *PingPongServer) MultiPing(stream pb.PingPong_MultiPingServer) error {
	msgs := []string{}
	for {
		// 提前结束接收消息
		if len(msgs) > 5 {
			return stream.SendAndClose(&pb.PongResponse{Value: "ping enough, max 5"})
		}

		msg, err := stream.Recv()
		if err != nil {
			// 客户端消息结束，返回响应信息
			if err == io.EOF {
				return stream.SendAndClose(&pb.PongResponse{Value: fmt.Sprintf("got %d ping", len(msgs))})
			}
			return err
		}
		msgs = append(msgs, msg.Value)
	}
}

// MultiPingPong 双向流模式
func (s *PingPongServer) MultiPingPong(stream pb.PingPong_MultiPingPongServer) error {
	msgs := []string{}
	for {
		// 接收消息
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		msgs = append(msgs, msg.Value)

		// 每收到两个消息响应一次
		if len(msgs)%2 == 0 {
			err = stream.Send(&pb.PongResponse{Value: "pong"})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// type UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)

// 服务端拦截器 - 记录请求和响应日志
func serverUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// 前置逻辑
	log.Printf("[Server Interceptor] accept request: %s", info.FullMethod)

	// 处理请求
	response, err := handler(ctx, req)

	// 后置逻辑
	log.Printf("[Server Interceptor] response: %s", response)

	return response, err
}

// type StreamServerInterceptor func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error

// 服务端拦截器 - 记录stream请求和响应日志
func serverStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 前置逻辑
	log.Printf("[Server Stream Interceptor] accept request: %s", info.FullMethod)

	// 处理请求，使用自定义 ServerStream
	err := handler(srv, &customServerStream{ss})
	return err
}

type customServerStream struct {
	grpc.ServerStream
}

func (s *customServerStream) SendMsg(m interface{}) error {
	log.Printf("[Server Stream Interceptor] send: %T", m)
	return s.ServerStream.SendMsg(m)
}

func (s *customServerStream) RecvMsg(m interface{}) error {
	log.Printf("[Server Stream Interceptor] recv: %T", m)
	return s.ServerStream.RecvMsg(m)
}

// 启动server
func main() {
	// 以option的方式添加拦截器
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(serverUnaryInterceptor),
		grpc.StreamInterceptor(serverStreamInterceptor),
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
