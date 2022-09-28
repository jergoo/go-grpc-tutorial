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
func serverLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// 前置逻辑
	log.Printf("[Server] accept request: %s", info.FullMethod)

	// 处理请求
	response, err := handler(ctx, req)

	// 后置逻辑
	log.Printf("[Server] response: %s", response)

	return response, err
}

func main() {
	// 以option的方式添加拦截器
	srv := grpc.NewServer(grpc.UnaryInterceptor(serverLogInterceptor))

...

```

### 客户端

```golang
// src/inteceptor/client.go

...

// 客户端拦截器 - 记录请求和响应日志
func clientLogInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// 前置逻辑
	log.Printf("[Client] send request: %s", method)

	// 发起请求
	err := invoker(ctx, method, req, reply, cc, opts...)

	// 后置逻辑
	log.Printf("[Client] response: %s", reply)

	return err
}

...

func Ping() {
	// 以option方式添加拦截器
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure(), grpc.WithUnaryInterceptor(clientLogInterceptor))
	if err != nil {
		log.Fatal(err)
	}

...

```

这里分别定义了 `serverLogInterceptor` 和 `clientLogInterceptor` 拦截器, 函数的签名定义在 `google.golang.org/grpc` 包中，分别为 `UnaryServerInterceptor` 和 `UnaryClientInterceptor`, 在前置逻辑位置可以对请求信息做处理，在后置逻辑位置可以对响应信息做处理。在初始化服务端和客户端连接时以option的形式配置就好了，同时也支持配置多个拦截器。

> 运行结果：
>
> ```sh
> // server
> 2022/09/27 23:25:36 [Server] accept request: /protos.PingPong/Ping
> 2022/09/27 23:25:36 [Server] response: value:"pong"
>
> // client
> 2022/09/27 23:25:36 [Client] send request: /protos.PingPong/Ping
> 2022/09/27 23:25:36 [Client] response: value:"pong"
> ```

## 流拦截器


### 服务端


### 客户端
