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
