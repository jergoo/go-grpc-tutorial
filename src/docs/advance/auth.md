# 安全认证

---

服务开发中需要考虑服务的安全性，如连接是否加密，用户请求是否有权限等，gRPC 支持基于 TLS 的认证保证连接的安全性，通知也支持基于 token 的认证方式，用于对用户做权限认证。

**源码目录：**

```
|—- src/
	|-- auth/
		|-- keys/     // 证书目录
		|—— client.go // 客户端
		|—— server.go // 服务端
	|—- protos/ping/
		|—— ping.proto   // protobuf描述文件
		|—— ping.pb.go   // protoc编译生成
    	|-- ping_grpc.pb.go // protoc编译生成
```


## TLS认证

首先需要准备服务端证书，在 keys 目录存放证书文件。

```sh
$ cd src/auth/keys
# 生成证书，-config 替换为对应系统的openssl配置文件目录
$ openssl req -newkey rsa:2048 -x509 -nodes -sha256 -days 3650 \
    -keyout server.key -new -out server.crt \
    -subj /CN=grpc.server -reqexts SAN -extensions SAN \
    -config <(cat /System/Library/OpenSSL/openssl.cnf \
        <(printf '[SAN]\nsubjectAltName=DNS:grpc.server'))
```

生成证书文件后，我们就可以在服务端通过 `ServerOption` 开启 TLS 了，示例如下：

```go
// src/auth/server.go

func main() {
	creds, err := credentials.NewServerTLSFromFile("keys/server.crt", "keys/server.key")
	if err != nil {
		log.Fatalf("load crt fail:%v", err)
	}
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
	}

	srv := grpc.NewServer(opts...)

...
```

类似的，在客户端使用 `DialOption` 开启 TLS，示例如下：

```golang
func Ping() {
	// 读取服务端证书，并制定对应服务名
	cred, err := credentials.NewClientTLSFromFile("keys/server.crt", "go-grpc-tutorial")
	if err != nil {
		log.Fatalf("load crt fail: %v", err)
	}

	// 连接配置
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(cred),
	}
	conn, err := grpc.Dial("localhost:1234", opts...)
	if err != nil {
		log.Fatal(err)
	}

	...
```

## Token 认证

TLS 认证是针对连接的安全加密方式，实际应用中还需要针对每个用户请求进行认证，常用的方式就是基于 token 认证，gRPC 使用 `grpc.PerRPCCredentials` 接口对此提供了支持。

```go
// PerRPCCredentials defines the common interface for the credentials which need to
// attach security information to every RPC (e.g., oauth2).
type PerRPCCredentials interface {
	// GetRequestMetadata gets the current request metadata, refreshing
	// tokens if required. This should be called by the transport layer on
	// each request, and the data should be populated in headers or other
	// context. If a status code is returned, it will be used as the status
	// for the RPC. uri is the URI of the entry point for the request.
	// When supported by the underlying implementation, ctx can be used for
	// timeout and cancellation. Additionally, RequestInfo data will be
	// available via ctx to this call.
	// TODO(zhaoq): Define the set of the qualified keys instead of leaving
	// it as an arbitrary string.
	GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
	// RequireTransportSecurity indicates whether the credentials requires
	// transport security.
	RequireTransportSecurity() bool
}

```

客户端示例如下：

```go
// src/auth/client.go

func Ping() {
	// 增加认证 Dial Option
	conn, err := grpc.Dial("localhost:1234", grpc.WithPerRPCCredentials(CustomAuth{Token: "1234567890"}))
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

// CustomAuth 自定义认证类型
type CustomAuth struct {
	Token string
}

// GetRequestMetadata 生成认证信息
func (a CustomAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": a.Token,
	}, nil
}

// RequireTransportSecurity 是否开启 TLS
func (a CustomAuth) RequireTransportSecurity() bool {
	return false
}
```

这里定义了一个 `CustomAuth` 类型，并实现了 `grpc.PerRPCCredentials` 接口的两个方法，通过 `grpc.WithPerRPCCredentials` 方法转换为 `DialOption` 类型初始化连接，这样每次 rpc 调用时 token 信息会通过请求的metadata 传输到服务端。

既然是通过 metadata 传输 token 信息，那么服务端认证就非常简单了，可以实现一个拦截器统一处理请求中的 token，示例如下：

```golang
// src/auth/server.go
...

// 服务端拦截器 - token 认证
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("authorization missing")
	}

	var token string
	if auth, ok := md["authorization"]; ok {
		token = auth[0]
	}
	if token != "1234567890" {
		return nil, grpc.Errorf(codes.Unauthenticated, "token invalid")
	}

	// 处理请求
	return handler(ctx, req)
}

// 启动server
func main() {
	srv := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))

...
```

以上就是基于 token 的认证方法，还是比较简单的，实际应用中可以根据自己的业务需求生成不同类型的 token，`google.golang.org/grpc/credentials/oauth`包也对 oauth2 和 jwt 提供了支持，感兴趣可以看一下 `oauth.NewOauthAccess(token *oauth2.Token)` 和 `oauth.NewJWTAccessFromKey(jsonKey []byte)` 方法，实际上也是实现了 `grpc.PerRPCCredentials` 接口。

---
