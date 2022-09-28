# Protobuf

---

gRPC推荐使用proto3，这里只介绍基本语法，重点关注如何描述一个服务，更多详细语法及高级特性请查看[官方文档](https://developers.google.com/protocol-buffers/)


## 基本结构

一个 protobuf 描述文件以`.proto`做为文件后缀，基本由三部分构成：

* 头部区域声明版本、包名、导入包及文件级别的option等信息
* 定义 service 及其 rpc 方法描述
* 定义 message/enum 等自义定数据类型

**示例：**

```protobuf
// -------------------- 头部区域 ----------------------------
syntax = "proto3";                        // 指定proto版本号，最新版使用proto3
import "some other package";              // 导入其它包
package pacakge_name;                     // 指定包名

option go_package = "go package path";    // go package 文件选项

// -------------------- 服务描述 ----------------------------

service ServiceName {
    rpc FuncName(Request) returns (Response);
}

// -------------------- 自定义数据类型 ----------------------------

message Request {
    string value = 1;
}

message Response {
    string value = 1;
}

```

**规范**

* 除结构定义外的语句以分号结尾，结构定义包括：message、service、enum
* message 命名采用驼峰命名方式，字段命名采用小写字母加下划线分隔方式
* enums 类型名采用驼峰命名方式，字段命名采用大写字母加下划线分隔方式
* service 名称与 rpc 方法名统一采用驼峰式命名
* 支持以 `//` 开头的单行注释

## 头部区域

首行要求明确声明使用的 protobuf 版本

```protobuf
syntax = "proto3"; 
```

### 导入

可以使用import语句导入使用其它 protobuf 描述文件中声明的类型，protoc 编译器会在编译命令中 `-I / --proto_path`参数指定的目录中查找导入的文件，如果没有指定该参数，默认在当前目录中查找。

**示例：**
```protobuf
syntax = "proto3"; 
import "google/protobuf/wrappers.proto"; // 导入其它包

...

service SomeService {
    // 使用包路径引用导入包的类型
    rpc getInfo(google.protobuf.StringValue) returns (Response);
}

```

### 包名

在`.proto`文件中使用`package`声明包名，避免命名冲突。

```
syntax = "proto3";
package foo.bar;
message Open {...}
```
在其他的消息格式定义中可以使用**包名+消息名**的方式来使用类型，如：

```
message Foo {
    ...
    foo.bar.Open open = 1;
    ...
}
```

## message

一个 `message` 定义描述了一个消息格式，是一个复合类型，和编程语言的结构体类似，protobuf 内置了一些基本类型，使用基本类型和其它复合类型组合定义一个 `message` 类型。

* 字段声明格式：`[类型] [字段名] = [Tag];`
* 所有的字段需要前置声明数据类型，除了 protobuf 内置基本类型也可以是其它 message 类型
* 每个字段都有一个**唯一的数值标签**，这些标签用于标识该字段在消息中的二进制格式，使用中的类型不应该随意改动
* 可以针对 message 和字段添加注释，注释内容会同步到编译生成的源码文件中
* 可以在类型名前使用 `repeated` 关键词，声明该字段为数组类型

**示例：**

```protobuf
// SearchRequest 搜索请求
message SearchRequest {
    string keyword = 1;     // 查询关键词
    int32  page_no = 2;     // 页码
    int32  page_size = 3;   // 数量
    repeated int32 arr = 4; // 数组
}
```

### 基本类型

| source   | 整型                      | 浮点            | 布尔 | 字符串 | 字节数组 |
| -------- | ------------------------- | --------------- | ---- | ------ | -------- |
| protobuf | int32/uint32/int64/uint64 | float/double    | bool | string | bytes    |
| go       | int32/uint32/int64/uint64 | float32/float64 | bool | string | []byte   |

### enum类型

当定义一个 message 时，想要一个字段只能是一个预定义好的值列表内的一个值，就需要用到enum类型了。**注意：每个enum定义的第一个元素值必须是0**

**示例：**

```protobuf
message Response {
  string value = 1;
  Status status = 2; // 使用Status类型
}

enum Status {
    OK = 0;
    FAIL = 1;
}
```

### map类型

proto3支持map类型声明: `map<key_type, value_type> field_name = N;`

* `key_type`类型可以是内置的基本类型(除浮点类型和`bytes`)
* `value_type`可以是除map以外的任意类型
* map字段不支持`repeated`属性
* 不要依赖map类型的字段顺序

**示例：**
```protobuf
message Project {...}
map<string, Project> projects = 1;
```

## service

service 描述一个RPC服务的接口，使用 `rpc` 关键字描述方法的签名，方法支持单次请求-响应(unary)和 stream 模式。protoc编译器会根据所选择的不同语言生成服务接口代码。生成的接口代码作为客户端与服务端的约定，服务端必须实现定义的所有接口方法，客户端直接调用同名方法向服务端发起请求。


**示例：**

```protobuf
service ServiceName {
    rpc Single (Request) returns (Response);                  // unary
    rpc ServerStream (Request) returns (stream Response);     // server stream
    rpc ClientStream (stream Request) returns (Response);     // client stream
    rpc BiStream (stream Request) returns (stream Response);  // bidirectional stream
}
```

## 编译

通过定义好的 `.proto` 文件生成各种语言的代码，需要安装编译器 `protoc` 及对应语言的插件。参考Github项目[google/protobuf](https://github.com/google/protobuf)安装编译器.

**示例命令：**
```sh
protoc --proto_path=IMPORT_PATH --cpp_out=DST_DIR --java_out=DST_DIR --python_out=DST_DIR --go_out=DST_DIR --go-grpc_out=DST_DIR --ruby_out=DST_DIR --javanano_out=DST_DIR --objc_out=DST_DIR --csharp_out=DST_DIR path/to/file.proto
```

## Protobuf⇢Go转换

我们定义一个示例文件对照说明常用结构的 protobuf 到 go 的转换，只说明关键部分代码，详细内容请查看完整文件。

```protobuf
// src/protos/example/example.proto

syntax = "proto3"; // 指定proto版本
package example; // 指定包名
option go_package="protos/example"; // 指定go包路径

// ExampleService 示例
service ExampleService {
    // Single 单次请求响应模式
    rpc Single(Request) returns (Response);
    // ServerStream 服务端流模式
    rpc ServerStream(Request) returns (stream Response);
    // ClientStream 客户端流模式
    rpc ClientStream(stream Request) returns (Response);
    // BiStream 双向流模式
    rpc BiStream(stream Request) returns (stream Response);
}

// Request 请求结构
message Request {
    string value = 1;
}

// Response 响应结构
message Response {
    string valuee = 1;
}

// Msg message 数据类型示例
message Msg {
    int32 i32 = 1;
    int64 i64 = 2;
    float f32 = 3;
    double f64  = 4;
    string str = 5;
    bool boolean = 6;
    bytes byteArr = 7;
    map<string, string> dict = 8;
    Status status = 9;
    EmbMsg embMsg = 10;
    repeated int64 intArr = 11;
}

message EmbMsg {
    string value = 1;
}

// Status 枚举
enum Status {
    OK = 0;
    FAIL = 1;
}

```

**编译：**
```sh
> cd src
> protoc --go_out=. --go-grpc_out=. ./protos/example/example.proto
```

### package

在proto文件中使用 `package` 关键字声明包名，默认转换成go中的包名与此一致。这里使用 `go_package` 选项用于控制编译结果文件的保存路径，这个路径会和编译命令中的`--go_out=.` 选项的路径拼接。比如这里当前目录是 `src`, 编译结果输出路径为 `./protos/example`。

```protobuf
package example; // 指定包名
option go_package="protos/example"; // 指定go包路径
```

### message

protobuf 中的 `message` 对应 go 中的 `struct`，全部使用驼峰命名规则，编译结果文件为 `{proto file name}.pb.go`。

```go
// src/protos/example/example.pb.go

// Msg message 数据类型示例
type Msg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	I32     int32             `protobuf:"varint,1,opt,name=i32,proto3" json:"i32,omitempty"`
	I64     int64             `protobuf:"varint,2,opt,name=i64,proto3" json:"i64,omitempty"`
	F32     float32           `protobuf:"fixed32,3,opt,name=f32,proto3" json:"f32,omitempty"`
	F64     float64           `protobuf:"fixed64,4,opt,name=f64,proto3" json:"f64,omitempty"`
	Str     string            `protobuf:"bytes,5,opt,name=str,proto3" json:"str,omitempty"`
	Boolean bool              `protobuf:"varint,6,opt,name=boolean,proto3" json:"boolean,omitempty"`
	ByteArr []byte            `protobuf:"bytes,7,opt,name=byteArr,proto3" json:"byteArr,omitempty"`
	Dict    map[string]string `protobuf:"bytes,8,rep,name=dict,proto3" json:"dict,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Status  Status            `protobuf:"varint,9,opt,name=status,proto3,enum=example.Status" json:"status,omitempty"`
	EmbMsg  *EmbMsg           `protobuf:"bytes,10,opt,name=embMsg,proto3" json:"embMsg,omitempty"`
	IntArr  []int64           `protobuf:"varint,11,rep,packed,name=intArr,proto3" json:"intArr,omitempty"`
}

type EmbMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

```
除了会生成对应的结构外，还会有些工具方法，如字段的getter:

```go
func (x *Msg) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_OK
}
```

枚举类型会生成对应名称的常量，同时会有两个map方便使用：

```go
// Status 枚举
type Status int32

const (
	Status_OK   Status = 0
	Status_FAIL Status = 1
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "OK",
		1: "FAIL",
	}
	Status_value = map[string]int32{
		"OK":   0,
		"FAIL": 1,
	}
)
```

### service

针对 service 的编译是由 `protoc-gen-go-grpc` 插件参与处理，编译结果文件为 `{proto file name}_grpc.pb.go`。代码中包含服务端和客户端接口的定义，客户端接口已经自动实现了，直接供客户端使用者调用，服务端接口需要由服务提供方实现。

```go
// 客户端接口
type ExampleServiceClient interface {
	// Single 单次请求响应模式
	Single(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// ServerStream 服务端流模式
	ServerStream(ctx context.Context, in *Request, opts ...grpc.CallOption) (ExampleService_ServerStreamClient, error)
	// ClientStream 客户端流模式
	ClientStream(ctx context.Context, opts ...grpc.CallOption) (ExampleService_ClientStreamClient, error)
	// BiStream 双向流模式
	BiStream(ctx context.Context, opts ...grpc.CallOption) (ExampleService_BiStreamClient, error)
}

// 服务端接口
type ExampleServiceServer interface {
	// Single 单次请求响应模式
	Single(context.Context, *Request) (*Response, error)
	// ServerStream 服务端流模式
	ServerStream(*Request, ExampleService_ServerStreamServer) error
	// ClientStream 客户端流模式
	ClientStream(ExampleService_ClientStreamServer) error
	// BiStream 双向流模式
	BiStream(ExampleService_BiStreamServer) error
	mustEmbedUnimplementedExampleServiceServer()
}
```

## 参考文档

* [Language Guide(proto3)](https://developers.google.com/protocol-buffers/docs/proto3)
* [Protocol Buffer Encoding](https://developers.google.com/protocol-buffers/docs/encoding)
* [Go Generated Code Guide](https://developers.google.com/protocol-buffers/docs/reference/go-generated)
