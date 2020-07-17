# Interceptor

grpc服务端和客户端都提供了interceptor功能，功能类似middleware，很适合在这里处理验证、日志等流程。

在自定义Token认证的示例中，认证信息是由每个服务中的方法处理并认证的，如果有大量的接口方法，这种姿势就太不优雅了，每个接口实现都要先处理认证信息。这个时候interceptor就可以用来解决了这个问题，在请求被转到具体接口之前处理认证信息，一处认证，到处无忧。
在客户端，我们增加一个请求日志，记录请求相关的参数和耗时等等。修改hello_token项目实现：


## 目录结构

```
|—— hello_interceptor/
	|—— client/
    	|—— main.go   // 客户端
	|—— server/
    	|—— main.go   // 服务端
|—— keys/             // 证书目录
	|—— server.key
	|—— server.pem
|—— proto/
	|—— hello/
		|—— hello.proto   // proto描述文件
		|—— hello.pb.go   // proto编译后文件
```

## 示例代码

**Step 1. 服务端interceptor：**
> hello_interceptor/server/main.go

```go
package main

import (
	"fmt"
	"net"

	pb "github.com/jergoo/go-grpc-example/proto/hello"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"       // grpc 响应状态码
	"google.golang.org/grpc/credentials" // grpc认证包
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata" // grpc metadata包
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"
)

// 定义helloService并实现约定的接口
type helloService struct{}

// HelloService Hello服务
var HelloService = helloService{}

// SayHello 实现Hello服务接口
func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s.", in.Name)

	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	// TLS认证
	creds, err := credentials.NewServerTLSFromFile("../../keys/server.pem", "../../keys/server.key")
	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
	}

	opts = append(opts, grpc.Creds(creds))

	// 注册interceptor
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	// 实例化grpc Server
	s := grpc.NewServer(opts...)

	// 注册HelloService
	pb.RegisterHelloServer(s, HelloService)

	grpclog.Println("Listen on " + Address + " with TLS + Token + Interceptor")

	s.Serve(listen)
}

// auth 验证Token
func auth(ctx context.Context) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
	}

	var (
		appid  string
		appkey string
	)

	if val, ok := md["appid"]; ok {
		appid = val[0]
	}

	if val, ok := md["appkey"]; ok {
		appkey = val[0]
	}

	if appid != "101010" || appkey != "i am key" {
		return grpc.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
	}

	return nil
}

// interceptor 拦截器
func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	err := auth(ctx)
	if err != nil {
		return nil, err
	}
	// 继续处理请求
	return handler(ctx, req)
}
```

**Step 2. 实现客户端interceptor：**

> hello_intercepror/client/main.go

```golang
package main

import (
	"time"

	pb "github.com/jergoo/go-grpc-example/proto/hello" // 引入proto包

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials" // 引入grpc认证包
	"google.golang.org/grpc/grpclog"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"

	// OpenTLS 是否开启TLS认证
	OpenTLS = true
)

// customCredential 自定义认证
type customCredential struct{}

// GetRequestMetadata 实现自定义认证接口
func (c customCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appid":  "101010",
		"appkey": "i am key",
	}, nil
}

// RequireTransportSecurity 自定义认证是否开启TLS
func (c customCredential) RequireTransportSecurity() bool {
	return OpenTLS
}

func main() {
	var err error
	var opts []grpc.DialOption

	if OpenTLS {
		// TLS连接
		creds, err := credentials.NewClientTLSFromFile("../../keys/server.pem", "server name")
		if err != nil {
			grpclog.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// 指定自定义认证
	opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))
	// 指定客户端interceptor
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor))

	conn, err := grpc.Dial(Address, opts...)
	if err != nil {
		grpclog.Fatalln(err)
	}
	defer conn.Close()

	// 初始化客户端
	c := pb.NewHelloClient(conn)

	// 调用方法
	req := &pb.HelloRequest{Name: "gRPC"}
	res, err := c.SayHello(context.Background(), req)
	if err != nil {
		grpclog.Fatalln(err)
	}

	grpclog.Println(res.Message)
}

// interceptor 客户端拦截器
func interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	grpclog.Printf("method=%s req=%v rep=%v duration=%s error=%v\n", method, req, reply, time.Since(start), err)
	return err
}

```

## 运行结果

```sh
$ cd hello_inteceptor/server && go run main.go
Listen on 127.0.0.1:50052 with TLS + Token + Interceptor
```

```sh
$ cd hello_inteceptor/client && go run main.go
method=/hello.Hello/SayHello req=name:"gRPC"  rep=message:"Hello gRPC."  duration=33.879699ms error=<nil>

Hello gRPC.
```


**项目推荐：**  [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)

这个项目对interceptor进行了封装，支持多个拦截器的链式组装，对于需要多种处理的地方使用起来会更方便些。

