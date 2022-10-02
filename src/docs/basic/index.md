# 入门

---

这个部分通过创建一个简单的服务说明 go gRPC 的基本使用方法和场景，并介绍 protobuf 的基本语法。

## 环境准备

### protobuf 编译器

项目地址：[google/protobuf](https://github.com/google/protobuf)

这里使用`brew`工具安装，也可以下载编译好的可执行文件：

```sh
$ brew install protobuf
```

执行`protoc`命令查看当前版本：

```sh
$ protoc --version
libprotoc 3.21.6
```

### go 编译插件

项目地址：
* [protobuf-go](https://github.com/protocolbuffers/protobuf-go)
* [grpc-go](https://github.com/grpc/grpc-go)

安装：
```sh
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

$ export PATH="$PATH:$(go env GOPATH)/bin"
```
