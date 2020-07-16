# 编译器使用

## 使用`protoc`命令编译`.proto`文件:

* -I 参数：指定import路径，可以指定多个 -I参数，按顺序查找，默认只查找当前目录
* --go_out ：golang编译支持，支持以下参数:
    * plugins=plugin1+plugin2 - 指定插件，目前只有grpc，即：plugins=grpc
    * M 参数 - 指定导入的.proto文件路径编译后对应的golang包名(不指定本参数默认就是.proto文件中import语句的路径)
    * import_prefix=xxx - 为所有import路径添加前缀，主要用于编译子目录内的多个proto文件，这个参数按理说很有用，尤其适用替代hello_http.proto编译时的M参数，但是实际使用时有个蛋疼的问题，自己尝试看看吧
    * import_path=foo/bar - 用于指定未声明package或go_package的文件的包名，最右面的斜线前的字符会被忽略
    * :编译文件路径  .proto文件路径(支持通配符)
    * 同一个包内包含多个.proto文件时使用通配符同时编译所有文件，单独编译每个文件会存在变量命名冲突，例如编译hello_http.proto那里所示

> 完整示例：

```protobuf
protoc --go_out=plugins=grpc,Mfoo/bar.proto=bar,import_prefix=foo/,import_path=foo/bar:. ./*.proto
```
