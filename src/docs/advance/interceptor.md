# 拦截器

---

在应用开发过程中会有这样的需求，就是在请求执行前后做一些通用的处理逻辑，比如记录日志、tracing、身份认证等，在web框架中一般是使用middleware来实现的，gRPC 在客户端和服务端都支持了拦截器功能，用来处理这种业务需求。

下面基于 PingPong 服务做一些扩展来演示拦截器的使用方法。

**源码目录：**

```
|—- src/
	|-- interceptor/
		|—— client.go // 客户端
		|—— server.go // 服务端
	|—- protos/ping/
		|—— ping.proto   // protobuf描述文件
		|—— ping.pb.go   // protoc编译生成
    	|-- ping_grpc.pb.go // protoc编译生成
```

## 普通拦截器

在客户端和服务端分别实现一个记录请求日志的拦截器，打印请求前后的信息。

### 服务端

```go
// src/interceptor/server.go

...

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

func main() {
	// 以option的方式添加拦截器
	srv := grpc.NewServer(grpc.UnaryInterceptor(serverUnaryInterceptor))

...

```

### 客户端

```golang
// src/inteceptor/client.go

...

// 客户端拦截器 - 记录请求和响应日志
func clientUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// 前置逻辑
	log.Printf("[Client Interceptor] send request: %s", method)

	// 发起请求
	err := invoker(ctx, method, req, reply, cc, opts...)

	// 后置逻辑
	log.Printf("[Client Interceptor] response: %s", reply)

	return err
}

...

func Ping() {
	// 以option方式添加拦截器
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure(), grpc.WithUnaryInterceptor(clientUnaryInterceptor))
	if err != nil {
		log.Fatal(err)
	}

...

```

这里分别定义了 `serverUnaryInterceptor` 和 `clientUnaryInterceptor` 拦截器，函数的签名定义在 `google.golang.org/grpc` 包中，分别为 `UnaryServerInterceptor` 和 `UnaryClientInterceptor`, 在前置逻辑位置可以对请求信息做处理，在后置逻辑位置可以对响应信息做处理。在初始化服务端和客户端连接时以option的形式配置就好了，同时也支持配置多个拦截器。

> 运行结果：
>
> ```sh
> // server
> 2022/09/27 00:00:00 [Server Interceptor] accept request: /protos.PingPong/Ping
> 2022/09/27 00:00:00 [Server Interceptor] response: value:"pong"
>
> // client
> 2022/09/27 00:00:00 [Client Interceptor] send request: /protos.PingPong/Ping
> 2022/09/27 00:00:00 [Client Interceptor] response: value:"pong"
> ```

## 流拦截器

同样实现一个打印请求和响应日志的拦截器，只是函数签名变成了 `grpc.StreamServerInterceptor` 和 `grpc.StreamClientInterceptor`。

### 服务端

```go
// src/interceptor/server.go

...

// 服务端拦截器 - 记录stream请求和响应日志
func serverStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 前置逻辑
	log.Printf("[Server Stream Interceptor] accept request: %s", info.FullMethod)

	// 处理请求
	err := handler(srv, ss)
	return err
}

...

func main() {
	// 以option的方式添加拦截器
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(serverUnaryInterceptor),
		grpc.StreamInterceptor(serverStreamInterceptor),
	}
	srv := grpc.NewServer(opts...)

...

```

以上实现其实和普通拦截器的使用方式没太大区别，但是流的特性在于请求和响应不是一次性处理完成的，而是多次发送和接收数据，所以我们可能需要在发送和接收数据的过程中处理一些公共逻辑，这才是流拦截器特别的地方。我们注意到 `handler` 方法调用的第二个参数是一个 `grpc.ServerStream` 接口类型，这个接口类型包含了 `SendMsg` 和 `RecvMsg` 方法，所以我们可以使用一个自定义类型实现这个接口，对原对象进行包装重写这两个方法，这样就能达到我们的目的了。

```go
// 服务端拦截器 - 记录stream请求和响应日志
func serverStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 前置逻辑
	log.Printf("[Server Stream Interceptor] accept request: %s", info.FullMethod)

	// 处理请求，使用自定义类型包装 ServerStream
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
```

### 客户端

客户端的使用方式和服务端类似，只是对应的数据处理接口类型变成了 `grpc.ClientStream`。

```go
// src/interceptor/client.go

···

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(clientStreamInterceptor),
	}
	conn, err := grpc.Dial("localhost:1234", opts...)

···

func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// 前置逻辑
	log.Printf("[Client Stream Interceptor] send request: %s", method)

	// 请求
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}

	// 自定义类型包装 ClientStream
	return &customClientStream{s}, nil
}

type customClientStream struct {
	grpc.ClientStream
}

func (s *customClientStream) SendMsg(m interface{}) error {
	log.Printf("[Client Stream Interceptor] send: %T", m)
	return s.ClientStream.SendMsg(m)
}

func (s *customClientStream) RecvMsg(m interface{}) error {
	log.Printf("[Client Stream Interceptor] recv: %T", m)
	return s.ClientStream.RecvMsg(m)
}
```


> 运行结果，以 MultiPingPong 方法为例，最后一次输出的recv是结束消息，err == io.EOF
> ```sh
> // server
> 2022/10/02 10:53:11 [Server Stream Interceptor] accept request: /protos.PingPong/MultiPingPong
> 2022/10/02 10:53:11 [Server Stream Interceptor] recv: *ping.PingRequest
> 2022/10/02 10:53:11 [Server Stream Interceptor] recv: *ping.PingRequest
> 2022/10/02 10:53:11 [Server Stream Interceptor] send: *ping.PongResponse
> 2022/10/02 10:53:11 [Server Stream Interceptor] recv: *ping.PingRequest
>
> // client
> 2022/10/02 10:53:11 [Client Stream Interceptor] send request: /protos.PingPong/MultiPingPong
> 2022/10/02 10:53:11 [Client Stream Interceptor] send: *ping.PingRequest
> 2022/10/02 10:53:11 send:ping
> 2022/10/02 10:53:11 [Client Stream Interceptor] send: *ping.PingRequest
> 2022/10/02 10:53:11 send:ping
> 2022/10/02 10:53:12 [Client Stream Interceptor] recv: *ping.PongResponse
> 2022/10/02 10:53:12 recv:pong
> 2022/10/02 10:53:13 [Client Stream Interceptor] recv: *ping.PongResponse
>```

---

> **注意**：在自定义的 `RecvMsg` 方法中，前置位置只能读取消息的类型，无法读取实际数据，因为这个时候接收到的消息还没有解析处理，如果要获取接收消息的实际数据，需要把自定义的处理逻辑放在后置位置，例如：
> ```go
> func (s *customClientStream) RecvMsg(m interface{}) error {
>	err := s.ClientStream.RecvMsg(m)
>	log.Printf("[Client Stream Interceptor] recv: %v", m)
>	return err
> }
>```
