package main

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/jergoo/go-grpc-tutorial/protos/ping" // 引入编译生成的包
)

// Ping 单次请求-响应模式
func Ping() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 初始化metadata
	md := metadata.New(map[string]string{"ts": strconv.FormatInt(time.Now().UnixMicro(), 10)})
	md.Set("key", "v1", "v2")

	// md := metadata.Pairs("key", "v1", "key", "v2")

	// md := make(metadata.MD)
	// md.Set("key", "v1", "v2")

	// context包装metadata
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx = metadata.AppendToOutgoingContext(ctx, "k3", "v3")

	// 实例化客户端并调用
	client := pb.NewPingPongClient(conn)

	var reponseMD metadata.MD
	res, err := client.Ping(ctx, &pb.PingRequest{Value: "ping"}, grpc.Header(&reponseMD))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got response md: %v", reponseMD)
	log.Println(res.Value)
}

// MultiPong 服务端流模式
func MultiPong() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 初始化metadata
	md := metadata.New(map[string]string{"ts": strconv.FormatInt(time.Now().UnixMicro(), 10)})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 实例化客户端并调用
	client := pb.NewPingPongClient(conn)
	// 获得对 stream 对象的引用
	stream, err := client.MultiPong(ctx, &pb.PingRequest{Value: "ping"})
	if err != nil {
		log.Fatal(err)
	}

	// 读取响应metadata
	mdResponse, err := stream.Header()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("get response md: %v", mdResponse)

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
