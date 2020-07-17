# HTTP网关

源自coreos的一篇博客 [Take a REST with HTTP/2, Protobufs, and Swagger](https://coreos.com/blog/grpc-protobufs-swagger.html)。

etcd3 API全面升级为gRPC后，同时要提供REST API服务，维护两个版本的服务显然不太合理，所以[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)诞生了。通过protobuf的自定义option实现了一个网关，服务端同时开启gRPC和HTTP服务，HTTP服务接收客户端请求后转换为grpc请求数据，获取响应后转为json数据返回给客户端。

结构如图：

![](../_media/grpc_rest_gateway.png)


## 安装grpc-gateway

```sh
$ go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

```

## 目录结构

```
|—— hello_http/
	|—— client/
    	|—— main.go   // 客户端
	|—— server/
    	|—— main.go   // GRPC服务端
	|—— server_http/
		|—— main.go   // HTTP服务端
|—— proto/
	|—— google       // googleApi http-proto定义
		|—— api
			|—— annotations.proto
			|—— annotations.pb.go
			|—— http.proto
			|—— http.pb.go
	|—— hello_http/
		|—— hello_http.proto   // proto描述文件
		|—— hello_http.pb.go   // proto编译后文件
		|—— hello_http_pb.gw.go // gateway编译后文件
```
这里用到了google官方Api中的两个proto描述文件，直接拷贝不要做修改，里面定义了protocol buffer扩展的HTTP option，为grpc的http转换提供支持。


## 示例代码

**Step 1. 编写proto描述文件：proto/hello_http.proto**

```protobuf
syntax = "proto3";

package hello_http;
option go_package = "hello_http";

import "google/api/annotations.proto";

// 定义Hello服务
service HelloHTTP {
    // 定义SayHello方法
    rpc SayHello(HelloHTTPRequest) returns (HelloHTTPResponse) {
        // http option
        option (google.api.http) = {
            post: "/example/echo"
            body: "*"
        };
    }
}

// HelloRequest 请求结构
message HelloHTTPRequest {
    string name = 1;
}

// HelloResponse 响应结构
message HelloHTTPResponse {
    string message = 1;
}

```
这里在原来的`SayHello`方法定义中增加了http option, POST方式，路由为"/example/echo"。

**Step 2. 编译proto**

```sh
$ cd proto

# 编译google.api
$ protoc -I . --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. google/api/*.proto

# 编译hello_http.proto
$ protoc -I . --go_out=plugins=grpc,Mgoogle/api/annotations.proto=github.com/jergoo/go-grpc-example/proto/google/api:. hello_http/*.proto

# 编译hello_http.proto gateway
$ protoc --grpc-gateway_out=logtostderr=true:. hello_http/hello_http.proto
```

注意这里需要编译google/api中的两个proto文件，同时在编译hello_http.proto时使用`M`参数指定引入包名，最后使用grpc-gateway编译生成`hello_http_pb.gw.go`文件，这个文件就是用来做协议转换的，查看文件可以看到里面生成的http handler，处理proto文件中定义的路由"example/echo"接收POST参数，调用HelloHTTP服务的客户端请求grpc服务并响应结果。

**Step 3: 实现服务端和客户端**

server/main.go和client/main.go的实现与hello项目一致，这里不再说明。

> server_http/main.go

```golang
package main

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	gw "github.com/jergoo/go-grpc-example/proto/hello_http"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// grpc服务地址
	endpoint := "127.0.0.1:50052"
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	// HTTP转grpc
	err := gw.RegisterHelloHTTPHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		grpclog.Fatalf("Register handler err:%v\n", err)
	}

	grpclog.Println("HTTP Listen on 8080")
	http.ListenAndServe(":8080", mux)
}
```

就是这么简单。开启了一个http server，收到请求后根据路由转发请求到对应的RPC接口获得结果。grpc-gateway做的事情就是帮我们自动生成了转换过程的实现。


## 运行结果

依次开启gRPC服务和HTTP服务端：

```sh
$ cd hello_http/server && go run main.go
Listen on 127.0.0.1:50052
```

```sh
$ cd hello_http/server_http && go run main.go
HTTP Listen on 8080
```

调用grpc客户端：

```sh
$ cd hello_http/client && go run main.go
Hello gRPC.

# HTTP 请求
$ curl -X POST -k http://localhost:8080/example/echo -d '{"name": "gRPC-HTTP is working!"}'
{"message":"Hello gRPC-HTTP is working!."}
```

## 升级版服务端

上面的使用方式已经实现了我们最初的需求，[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)项目中提供的示例也是这种使用方式，这样后台需要开启两个服务两个端口。其实我们也可以只开启一个服务，同时提供http和gRPC调用方式。

新建一个项目`hello_http_2`, 基于`hello_tls`项目改造。客户端只要修改调用的proto包地址就可以了，这里我们看服务端的实现：

> hello\_http_2/server/main.go

```golang
package main

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	pb "github.com/jergoo/go-grpc-example/proto/hello_http"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

// 定义helloHTTPService并实现约定的接口
type helloHTTPService struct{}

// HelloHTTPService Hello HTTP服务
var HelloHTTPService = helloHTTPService{}

// SayHello 实现Hello服务接口
func (h helloHTTPService) SayHello(ctx context.Context, in *pb.HelloHTTPRequest) (*pb.HelloHTTPResponse, error) {
	resp := new(pb.HelloHTTPResponse)
	resp.Message = "Hello " + in.Name + "."

	return resp, nil
}

func main() {
	endpoint := "127.0.0.1:50052"
	conn, err := net.Listen("tcp", endpoint)
	if err != nil {
		grpclog.Fatalf("TCP Listen err:%v\n", err)
	}

	// grpc tls server
	creds, err := credentials.NewServerTLSFromFile("../../keys/server.pem", "../../keys/server.key")
	if err != nil {
		grpclog.Fatalf("Failed to create server TLS credentials %v", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHelloHTTPServer(grpcServer, HelloHTTPService)

	// gw server
	ctx := context.Background()
	dcreds, err := credentials.NewClientTLSFromFile("../../keys/server.pem", "server name")
	if err != nil {
		grpclog.Fatalf("Failed to create client TLS credentials %v", err)
	}
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}
	gwmux := runtime.NewServeMux()
	if err = pb.RegisterHelloHTTPHandlerFromEndpoint(ctx, gwmux, endpoint, dopts); err != nil {
		grpclog.Fatalf("Failed to register gw server: %v\n", err)
	}

	// http服务
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	srv := &http.Server{
		Addr:      endpoint,
		Handler:   grpcHandlerFunc(grpcServer, mux),
		TLSConfig: getTLSConfig(),
	}

	grpclog.Infof("gRPC and https listen on: %s\n", endpoint)

	if err = srv.Serve(tls.NewListener(conn, srv.TLSConfig)); err != nil {
		grpclog.Fatal("ListenAndServe: ", err)
	}

	return
}

func getTLSConfig() *tls.Config {
	cert, _ := ioutil.ReadFile("../../keys/server.pem")
	key, _ := ioutil.ReadFile("../../keys/server.key")
	var demoKeyPair *tls.Certificate
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		grpclog.Fatalf("TLS KeyPair err: %v\n", err)
	}
	demoKeyPair = &pair
	return &tls.Config{
		Certificates: []tls.Certificate{*demoKeyPair},
		NextProtos:   []string{http2.NextProtoTLS}, // HTTP2 TLS支持
	}
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

```

gRPC服务端接口的实现没有区别，重点在于HTTP服务的实现。gRPC是基于http2实现的，`net/http`包也实现了http2，所以我们可以开启一个HTTP服务同时服务两个版本的协议，在注册http handler的时候，在方法`grpcHandlerFunc`中检测请求头信息，决定是直接调用gRPC服务，还是使用gateway的HTTP服务。`net/http`中对http2的支持要求开启https，所以这里要求使用https服务。

**步骤**

* 注册开启TLS的grpc服务
* 注册开启TLS的gateway服务，地址指向grpc服务
* 开启HTTP server

### 运行结果

```sh
$ cd hello_http_2/server && go run main.go
gRPC and https listen on: 127.0.0.1:50052
```

```sh
$ cd hello_http_2/client && go run main.go
Hello gRPC.

# HTTP 请求
$ curl -X POST -k https://localhost:50052/example/echo -d '{"name": "gRPC-HTTP is working!"}'
{"message":"Hello gRPC-HTTP is working!."}
```




