# metadata

---

服务间使用 Http 相互调用时，经常会设置一些业务自定义 header 如时间戳、trace信息等，gRPC使用 HTTP/2 协议自然也是支持的，gRPC 通过 `google.golang.org/grpc/metadata` 包内的 `MD` 类型提供相关的功能接口。

**源码目录：**

```
|—- src/
	|-- metadata/
		|—— client.go // 客户端
		|—— server.go // 服务端
	|—- protos/ping/
		|—— ping.proto   // protobuf描述文件
		|—— ping.pb.go   // protoc编译生成
    	|-- ping_grpc.pb.go // protoc编译生成
```

## 类型定义

```go
// MD is a mapping from metadata keys to values. Users should use the following
// two convenience functions New and Pairs to generate MD.
type MD map[string][]string
```

`metadata.MD` 类型的定义非常简单，可以像一个普通的 map 一样直接操作，同时 metadata 包里封装了很多工具方法供我们使用。

```go
// 使用 New 方法创建
md := metadata.New(map[string]string{"k1":"v1", "k2", "v2"})

// 直接使用 make 创建
md := make(metadata.MD)

// 使用 Pairs 方法创建
md := metadata.Pairs("k1", "v1-1", "k1", "v1-2")

// 一些操作
md.Set("key", "v1", "v2")
md.Append("key", "v3")
md.Delete("key")
vals := md.Get("key")

```

## 发送与接收
### 客户端

客户端请求的 metadata 是通过设置 context 使用的，metadata 包提供了两个 context 相关的方法，设置好 context 后直接在调用 rpc 方法时传入即可，示例代码如下：

```go
md := metadata.New(map[string]string{"k1":"v1", "k2", "v2"})

// 使用 NewOutgoingContext 初始化一个新的 context
ctx := metadata.NewOutgoingContext(context.Background(), md)

// 使用 AppendToOutgoingContext 向 context 追加 metadata
ctx = metadata.AppendToOutgoingContext(ctx, "k3", "v3")

```

客户端接收响应中的 metadata 需要区分普通 rpc 和 stream rpc，示例如下：

```go
// 普通 rpc，使用 grpc.Header 方法包装为 CallOption
var md metadata.MD
res, err := client.Ping(ctx, &pb.PingRequest{Value: "ping"}, grpc.Header(&md))

// stream rpc
stream, err := client.MultiPong(context.Background(), &pb.PingRequest{Value: "ping"})
if err != nil {
    log.Fatal(err)
}

// 通过 stream 对象的 Header 方法获取
md, err := stream.Header()
if err != nil {
    log.Fatal(err)
}
```

### 服务端

对应客户端请求的 metadata 是使用 context 设置的，那么服务端在接收时自然也是从 context 中读取，metadata 包中的 `FromIncommingContext` 方法就是用来读取 context 中的 metadata数据的，示例如下：

```go
// unary rpc
func (s *PingPongServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	// 读取请求metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("Got md: %v", md)
	}

...

// stream rpc
func (s *PingPongServer) MultiPingPong(stream pb.PingPong_MultiPingPongServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Printf("Got md: %v", md)
	}

...
```

服务端设置响应的 metadata 也非常简单，只需要调用封装好的 `SetHeader` 或 `SendHeader` 方法即可，示例如下：

```go
// unary rpc
func (s *PingPongServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	// 读取请求metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("Got md: %v", md)
	}

	// SetHeader设置响应 metadata
	grpc.SetHeader(ctx, metadata.New(map[string]string{"rkey": "rval"}))
    // 注意 SendHeader 只能调用一次
    // grpc.SendHeader(ctx, metadata.New(map[string]string{"rkey": "rval"}))

	....

// stream rpc, 调用 stream 的 SetHeader 方法
func (s *PingPongServer) MultiPong(req *pb.PingRequest, stream pb.PingPong_MultiPongServer) error {
	stream.SetHeader(metadata.New(map[string]string{"rkey": "rval"}))

    // 注意 SendHeader 只能调用一次
    // stream.SendHeader(metadata.New(map[string]string{"rkey": "rval"}))
```
---

