# 安装

## protobuf 编译器

项目地址：[google/protobuf](https://github.com/google/protobuf)

这里使用`brew`工具安装

```sh
$ brew install protobuf
```

执行`protoc`命令查看当前版本：

```sh
$ protoc --version
libprotoc 3.21.6
```
## Go 编译插件

项目地址：
* [protobuf-go](https://github.com/protocolbuffers/protobuf-go)
* [grpc-go](https://github.com/grpc/grpc-go)

安装：
```sh
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

$ export PATH="$PATH:$(go env GOPATH)/bin"
```

grpc-go
-------

项目地址：[grpc-go](https://github.com/grpc/grpc-go)

* 要求golang版本 >= 1.6

```sh
$ go get -u google.golang.org/grpc
```


golang protobuf
---------------

项目地址：[golang/protobuf](https://github.com/golang/protobuf)

* 要求golang版本 > 1.4
* 使用前要求安装protobuf编译器

```sh
$ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

安装`protoc-gen-go`到 `$GOPATH/bin`。 注意：该目录必须在系统的环境变量`$PATH`中。

如果一路没有问题的话，到此为止，需要的环境都安装好了😀。


编译器使用
--------

使用`protoc`命令编译`.proto`文件,不同语言支持需要指定输出参数，如：

```sh
$ protoc --proto_path=IMPORT_PATH --cpp_out=DST_DIR --java_out=DST_DIR --python_out=DST_DIR --go_out=DST_DIR --ruby_out=DST_DIR --javanano_out=DST_DIR --objc_out=DST_DIR --csharp_out=DST_DIR path/to/file.proto
```

这里详细介绍golang的编译姿势:

* `-I` 参数：指定import路径，可以指定多个`-I`参数，编译时按顺序查找，不指定时默认查找当前目录
* `--go_out` ：golang编译支持，支持以下参数

	* `plugins=plugin1+plugin2` - 指定插件，目前只支持grpc，即：`plugins=grpc`
	* `M` 参数 - 指定导入的.proto文件路径编译后对应的golang包名(不指定本参数默认就是`.proto`文件中`import`语句的路径)
	* `import_prefix=xxx` - 为所有`import`路径添加前缀，主要用于编译子目录内的多个proto文件，这个参数按理说很有用，尤其适用替代一些情况时的`M`参数，但是实际使用时有个蛋疼的问题导致并不能达到我们预想的效果，自己尝试看看吧
	* `import_path=foo/bar` - 用于指定未声明`package`或`go_package`的文件的包名，最右面的斜线前的字符会被忽略
	* 末尾 `:编译文件路径  .proto文件路径(支持通配符)`

完整示例：

```sh
$ protoc -I . --go_out=plugins=grpc,Mfoo/bar.proto=bar,import_prefix=foo/,import_path=foo/bar:. ./*.proto
```












