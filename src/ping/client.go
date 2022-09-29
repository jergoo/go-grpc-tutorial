package main

import (
	"context"
	"io"
	"log"
	"time"

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

// MultiPong 服务端流模式
func MultiPong() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 实例化客户端并调用
	client := pb.NewPingPongClient(conn)
	// 获得对 stream 对象的引用
	stream, err := client.MultiPong(context.Background(), &pb.PingRequest{Value: "ping"})
	if err != nil {
		log.Fatal(err)
	}

	// 循环接收数据流
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		log.Println(msg.Value)
	}
}

// MultiPing 客户端流模式
func MultiPing() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 实例化客户端并调用
	client := pb.NewPingPongClient(conn)
	stream, err := client.MultiPing(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 发送数据
	for i := 0; i < 5; i++ {
		data := &pb.PingRequest{Value: "ping"}
		err = stream.Send(data)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 发送结束并获取服务端响应
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res.Value)
}

// MultiPingPong 双向流模式
func MultiPingPong() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 实例化客户端并调用
	client := pb.NewPingPongClient(conn)
	stream, err := client.MultiPingPong(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 在另一个goroutine中接收数据
	c := make(chan struct{})
	go func(stream pb.PingPong_MultiPingPongClient, c chan struct{}) {
		defer func() {
			c <- struct{}{}
		}()
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			log.Printf("recv:%s\n", msg.Value)
		}
	}(stream, c)

	// 发送数据
	for i := 0; i < 6; i++ {
		data := &pb.PingRequest{Value: "ping"}
		err = stream.Send(data)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("send:%s\n", data.Value)
		time.Sleep(500 * time.Millisecond)
	}

	// 结束发送
	stream.CloseSend()
	// 等待接收完成
	<-c
}
