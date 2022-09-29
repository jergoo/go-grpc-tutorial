#! /bin/bash

# 编译test.proto
protoc -I . --go_out=plugins=grpc:. test/test.proto

# 编译hello.proto
protoc -I . --go_out=plugins=grpc:. hello/hello.proto

# 编译google api，新版编译器可以省略M参数
protoc -I . --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. google/api/*.proto

# 编译hello_http.proto
protoc -I . --go_out=plugins=grpc,Mgoogle/api/annotations.proto=github.com/jergoo/go-grpc-example/proto/google/api:. hello_http/*.proto

# 编译hello_http.proto gateway
protoc --grpc-gateway_out=logtostderr=true:. hello_http/hello_http.proto
