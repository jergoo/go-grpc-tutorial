# gRPC 流

---

从其名称可以理解，流就是持续不断的传输。有一些业务场景请求或者响应的数据量比较大，不适合使用普通的 RPC 调用通过一次请求-响应处理，一方面是考虑数据量大对请求响应时间的影响，另一方面业务场景的设计不一定需要一次性处理完所有数据，这时就可以使用流来分批次传输数据。gRPC支持单向流和双向流，只需要在 service 的 rpc 方法描述中通过 `stream` 关键字指定启用流特性就好了，下面通过两个示例来说明使用方法。

**源码目录：**

```
|—- src/
	|-- ping/
		|—— client.go // 客户端
		|—— server.go // 服务端
	|—- protos/ping/
		|—— ping.proto   // protobuf描述文件
		|—— ping.pb.go   // protoc编译生成
    	|-- ping_grpc.pb.go // protoc编译生成
```

## 单向流

单向流是指客户端和服务端只有一端开启流特性，这里的单向特指**发送数据的方向**。

* 当服务端开启流时，客户端和普通 RPC 调用一样通过一次请求发送数据，服务端通过流分批次响应。
* 当客户端开启流时，客户端通过流分批次发送请求数据，服务端接完所有数据后统一响应一次。

### 服务端流

定义一个 `MultiPong` 方法，在服务端开启流，功能是接收到客户端的请求后响应10次 pong 消息。

```protobuf
...

service PingPong {
    // 服务端流模式，在响应消息前添加 stream 关键字
    rpc MultiPong(PingRequest) returns (stream PongResponse);
}

...
```

**服务端实现**：第二个参数为 stream 对象的引用，可以通过它的 `Send` 方法发送数据。

```go
// src/ping/server.go

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
```

**客户端实现**：请求方式和普通 RPC 没有区别，重点关注对响应数据流的处理，通过一个 for 循环接收数据直到结束。

```go
// src/ping/client.go

func MultiPong() {

    ...
    
    // 获得对 stream 对象的引用
	stream, err := client.MultiPong(context.Background(), &pb.PingRequest{Value: "ping"})
	if err != nil {
		log.Fatal(err)
	}

	// 循环接收响应数据流
	for {
		msg, err := stream.Recv()
		if err != nil {
            // 数据结束
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		log.Println(msg.Value)
	}
}
```

### 客户端流

定义一个 `MultiPing` 方法，在客户端开启流，功能是持续发送多个 ping 请求，服务端统一响应一次。

```protobuf
...

service PingPong {
	// 客户端流模式，在请求消息前添加 stream 关键字
	rpc MultiPing(stream PingRequest) returns (PongResponse);
}

...
```

**服务端实现**：只有一个参数为 stream 对象的引用，可以通过它的 `Recv` 方法接收数据。使用 `SendAndClose` 方法关闭流并响应，服务端可以根据需要提前关闭。

```go
// src/ping/server.go

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
```

**客户端实现**：调用 `MultiPing` 方法时不再指定请求参数，而是通过返回的 stream 对象的 `Send` 分批发送数据。

```go
// src/ping/client.go

func MultiPing() {
    ...

    // 调用并得到 stream 对象
    stream, err := client.MultiPing(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 发送数据
	for i := 0; i < 6; i++ {
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
```

> 运行结果：
> ```sh
> // 发送5个ping
> 2022/09/27 00:00:00 got 5 ping
> 
> // 发送10个ping
> 2022/09/27 00:00:00 ping enough, max 5
> ```

## 双向流

双向流是指客户端在发送数据和服务端响应数据的过程中都启用流特性，实际上单向流只是双向流的特例，有了上面的基础，双向流就很好理解了。

定义一个 `MultiPingPong` 方法，在客户端和服务端都开启流，功能是服务端每接收到两个 ping 就响应一次 pong。

```protobuf
...

service PingPong {
	// 双向流模式
	rpc MultiPingPong(stream PingRequest) returns (stream PongResponse);
}

...
```

**服务端实现**：同样通过 stream 的 `Recv` 和 `Send` 方法接收和发送数据。

```go
// src/ping/server.go

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
```

**客户端实现**：这里在另外一个 goroutine 里处理接收数据的逻辑来演示同时发送和接收数据。

```go
// src/ping/client.go
func MultiPingPong() {
    ...

	stream, err := client.MultiPingPong(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 在另一个goroutine中处理接收数据
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

        // 延时一段时间发送，等待响应结果
		time.Sleep(500 * time.Millisecond)
	}

	// 结束发送
	stream.CloseSend()
	// 等待接收完成
	<-c
}

```

> 运行结果：
> ```sh
> 2022/09/29 09:09:59 send:ping
> 2022/09/29 09:10:00 send:ping
> 2022/09/29 09:10:00 recv:pong
> 2022/09/29 09:10:00 send:ping
> 2022/09/29 09:10:01 send:ping
> 2022/09/29 09:10:01 recv:pong
> 2022/09/29 09:10:01 send:ping
> 2022/09/29 09:10:02 send:ping
> 2022/09/29 09:10:02 recv:pong
> ```

---
