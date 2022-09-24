# Protobuf语法

gRPC推荐使用proto3，这里介绍基本语法，更多详细使用方式建议查看[官方文档](https://developers.google.com/protocol-buffers/)

## Message定义

一个`message`类型定义描述了一个请求或响应的消息格式，可以包含多种类型字段。

例如定义一个搜索请求的消息格式`SearchRequest`，每个请求包含查询字符串、页码、每页数目。每个字段声明以分号结尾。

```protobuf
syntax = "proto3";

message SearchRequest {
    string query = 1;
    int32  page_number = 2;
    int32  result_per_page = 3;
}
```

> **首行要求明确声明使用的protobuf版本为`proto3`，如果不声明，编译器默认使用`proto2`。**


一个`.proto`文件中可以定义多个message，一般用于同时定义多个相关的message，例如在同一个.proto文件中同时定义搜索请求和响应消息：

```protobuf
syntax = "proto3";

message SearchRequest {
    string query = 1;
    int32  page_number = 2;
    int32  result_per_page = 3;
}

message SearchResponse {
	...
}
```

### 字段类型声明

所有的字段需要前置声明数据类型，上面的示例指定了两个数值类型和一个字符串类型。除了基本的标量类型还有复合类型，如枚举、map、数组、甚至其它message类型等。后面会依次说明。


### 分配Tags

消息的定义中，每个字段都有一个**唯一的数值标签**。这些标签用于标识该字段在消息中的二进制格式，使用中的类型不应该随意改动。其中，[1-15]内的标识在编码时占用一个字节，包含标识和字段类型。[16-2047]之间的标识符占用2个字节。建议为频繁出现的消息元素分配[1-15]间的标签。如果考虑到以后可能或扩展频繁元素，可以预留一些。

最小的标识符可以从1开始，最大到2<sup>29</sup> - 1，或536,870,911。不可以使用[19000－19999]之间的标识符， Protobuf协议实现中预留了这些标识符。在.proto文件中使用这些预留标识号，编译时会报错。


### 字段规则

* 单数形态：一个message内同名单数形态的字段不能超过一个
* repeated：前置`repeated`关键词，声明该字段为数组类型

### 添加注释

支持双斜线`//`单行注释

```protobuf
syntax = "proto3";              // 协议版本声明

// SearchRequest 搜索请求消息
message SearchRequest {
    string query = 1;           // 查询字符串
    int32  page_number = 2;     // 页码
    int32  result_per_page = 3; // 每页条数
}
```

## Service定义

如果想要将消息类型用在RPC(远程方法调用)系统中，可以在`.proto`文件中定义一个RPC服务接口，protoc编译器会根据所选择的不同语言生成服务接口代码。例如，想要定义一个RPC服务并具有一个方法，该方法接收`SearchRequest`并返回一个`SearchResponse`，此时可以在`.proto`文件中进行如下定义：

```
service SearchService {
    rpc Search (SearchRequest) returns (SearchResponse) {}
}
```

生成的接口代码作为客户端与服务端的约定，服务端必须实现定义的所有接口方法，客户端直接调用同名方法向服务端发起请求。

## 数据类型

### 基本类型
| .proto   | Go      |
| -------- | ------- |
| double   | float64 |
| float    | float32 |
| int32    | int32   |
| int64    | int64   |
| uint32   | uint32  |
| uint64   | uint64  |
| sint32   | int32   |
| sint64   | int64   |
| fixed32  | uint32  |
| fixed64  | uint64  |
| sfixed32 | int32   |
| sfixed64 | int64   |
| bool     | bool    |
| string   | string  |
| bytes    | []byte  |

> 序列化时的编码规则请参考 [Protocol Buffer Encoding](https://developers.google.com/protocol-buffers/docs/encoding).

### 默认值

* 字符串类型默认为空字符串
* 字节类型默认为空字节
* 布尔类型默认false
* 数值类型默认为0值
* enums类型默认为第一个定义的枚举值，必须是0

针对不同语言的默认值的具体行为参考 [generated code guide](https://developers.google.com/protocol-buffers/docs/reference/overview)

### 枚举(Enum) 

当定义一个message时，想要一个字段只能是一个预定义好的值列表内的一个值，就需要用到enum类型了。

示例：这里定义一个名为`Corpus`的enum类型值，并且指定一个字段为`Corpus`类型

```protobuf
message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
  // 定义enum类型
  enum Corpus {
    UNIVERSAL = 0;
    WEB = 1;
    IMAGES = 2;
    LOCAL = 3;
    NEWS = 4;
    PRODUCTS = 5;
    VIDEO = 6;
  }
  Corpus corpus = 4; // 使用Corpus作为字段类型
}
```

> **注意：每个enum定义的第一个元素值必须是0**


还可以通过给不同的enum元素赋相同的值来定义别名，要求设置`allow_alias`选项的值为`true`，否则会报错。

```protobuf
// 正确示例
enum EnumAllowingAlias {
  option allow_alias = true; // 开启allow_alias选项
  UNKNOWN = 0;
  STARTED = 1;
  RUNNING = 1;  // RUNNING和STARTED互为别名
}

// 错误示例
enum EnumNotAllowingAlias {
  UNKNOWN = 0;
  STARTED = 1;
  RUNNING = 1;  // 未开启allow_alias选项，编译会报错
}
```

> enum类型值同样支持文件级定义和message内定义，引用方式与message嵌套一致

### 使用其它Message

可以使用其它message类型作为字段类型。

例如，在`SearchResponse`中包含`Result`类型的消息，可以在相同的`.proto`文件中定义`Result`消息类型，然后在`SearchResponse`中指定字段类型为`Result`:

```protobuf
message SearchResponse {
    repeated Result results = 1;
}

message Result {
    string url = 1;
    string title = 2;
    repeated string snippets = 3;
}
```

### Message嵌套

上面已经介绍过可以嵌套使用另一个message作为字段类型，其实还可以在一个message内部定义另一个message类型，作为子级message。

示例：`Result`类型在`SearchResponse`类型中定义并直接使用，作为`results`字段的类型

```protobuf
message SearchResponse {
    message Result {
        string url = 1;
        string title = 2;
        repeated string snippets = 3;
    }
    repeated Result results = 1;
}
```

内部声明的message类型名称只可在内部直接使用，在外部引用需要前置父级message名称,如`Parent.Type`：

```protobuf
message SomeOtherMessage {
    SearchResponse.Result result = 1;
}
```

支持多层嵌套：

```
message Outer {                // Level 0
    message MiddleAA {         // Level 1
        message Inner {        // Level 2
            int64 ival = 1;
            bool  booly = 2;
        }
    }
    message MiddleBB {         // Level 1
        message Inner {        // Level 2
            int32 ival = 1;
            bool  booly = 2;
        }
    }
}
```


### Map类型

proto3支持map类型声明:

```protobuf
map<key_type, value_type> map_field = N;

message Project {...}
map<string, Project> projects = 1;

```

* `key_type`类型可以是内置的标量类型(除浮点类型和`bytes`)
* `value_type`可以是除map以外的任意类型
* map字段不支持`repeated`属性
* 不要依赖map类型的字段顺序

## 导入(import)

可以使用import语句导入使用其它描述文件中声明的类型：

```protobuf
import "others.proto";
```

protoc编译器会在编译命令中 `-I / --proto_path`参数指定的目录中查找导入的文件，如果没有指定该参数，默认在当前目录中查找。


## 包(Packages)

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

## 选项(Options)

在定义.proto文件时可以标注一系列的options。Options并不改变整个文件声明的含义，但却可以影响特定环境下处理方式。完整的可用选项可以查看`google/protobuf/descriptor.proto`.

一些选项是文件级别的，意味着它可以作用于顶层作用域，不包含在任何消息内部、enum或服务定义中。一些选项是消息级别的，可以用在消息定义的内部。

以下是一些常用的选项：

* `java_package` (file option)：指定生成java类所在的包，如果在.proto文件中没有明确的声明java_package，会使用默认包名。不需要生成java代码时不起作用
* `java_outer_classname` (file option)：指定生成Java类的名称，如果在.proto文件中没有明确声明java_outer_classname，生成的class名称将会根据.proto文件的名称采用驼峰式的命名方式进行生成。如（foo_bar.proto生成的java类名为FooBar.java）,不需要生成java代码时不起任何作用
* `objc_class_prefix` (file option): 指定Objective-C类前缀，会前置在所有类和枚举类型名之前。没有默认值，应该使用3-5个大写字母。注意所有2个字母的前缀是Apple保留的。
* `go_package` (file option): 指定生成go代码的包名，也会影响生成文件的路径


## 基本规范

*  描述文件以`.proto`做为文件后缀，除结构定义外的语句以分号结尾
	* 结构定义包括：message、service、enum
	* rpc方法定义结尾的分号可有可无

* Message命名采用驼峰命名方式，字段命名采用小写字母加下划线分隔方式
		
	```protobuf
	message SongServerRequest {
	    required string song_name = 1;
	}
	```

* Enums类型名采用驼峰命名方式，字段命名采用大写字母加下划线分隔方式

	```protobuf
	enum Foo {
	    FIRST_VALUE = 1;
	    SECOND_VALUE = 2;
	}
	```

* Service名称与RPC方法名统一采用驼峰式命名


## 编译

通过定义好的`.proto`文件生成Java, Python, C++, Go, Ruby, JavaNano, Objective-C, or C# 代码，需要安装编译器`protoc`。参考Github项目[google/protobuf](https://github.com/google/protobuf)安装编译器.

运行命令：
```
protoc --proto_path=IMPORT_PATH --cpp_out=DST_DIR --java_out=DST_DIR --python_out=DST_DIR --go_out=DST_DIR --go-grpc_out=DST_DIR --ruby_out=DST_DIR --javanano_out=DST_DIR --objc_out=DST_DIR --csharp_out=DST_DIR path/to/file.proto
```

具体语言的编译实例请参考[详细文档](https://developers.google.com/protocol-buffers/docs/reference/overview)。

## 更多

* [Any类型](https://developers.google.com/protocol-buffers/docs/proto3#any)
* [Oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof)
* [自定义Options](https://developers.google.com/protocol-buffers/docs/proto.html#customoptions)


## Protobuf⇢Go转换

这里使用一个测试文件对照说明常用结构的protobuf到golang的转换。只说明关键部分代码，详细内容请查看完整文件。示例文件在`proto/test`目录下。

### Package

在proto文件中使用`package`关键字声明包名，默认转换成go中的包名与此一致，如果需要指定不一样的包名，可以使用`go_package`选项：

```protobuf
package test;
option go_package="test";
```

### Message

proto中的`message`对应go中的`struct`，全部使用驼峰命名规则。嵌套定义的`message`，`enum`转换为go之后，名称变为`Parent_Child`结构。

示例proto：

```protobuf
// Test 测试
message Test {
    int32 age = 1;
    int64 count = 2;
    double money = 3;
    float score = 4;
    string name = 5;
    bool fat = 6;
    bytes char = 7;
    // Status 枚举状态
    enum Status {
        OK = 0;
        FAIL = 1;
    }
    Status status = 8;
    // Child 子结构
    message Child {
        string sex = 1;
    }
    Child child = 9;
	map<string, string> dict = 10;
}
```

转换结果：

```go
// Status 枚举状态
type Test_Status int32

const (
	Test_OK   Test_Status = 0
	Test_FAIL Test_Status = 1
)

// Test 测试
type Test struct {
	Age    int32       `protobuf:"varint,1,opt,name=age" json:"age,omitempty"`
	Count  int64       `protobuf:"varint,2,opt,name=count" json:"count,omitempty"`
	Money  float64     `protobuf:"fixed64,3,opt,name=money" json:"money,omitempty"`
	Score  float32     `protobuf:"fixed32,4,opt,name=score" json:"score,omitempty"`
	Name   string      `protobuf:"bytes,5,opt,name=name" json:"name,omitempty"`
	Fat    bool        `protobuf:"varint,6,opt,name=fat" json:"fat,omitempty"`
	Char   []byte      `protobuf:"bytes,7,opt,name=char,proto3" json:"char,omitempty"`
	Status Test_Status `protobuf:"varint,8,opt,name=status,enum=test.Test_Status" json:"status,omitempty"`
	Child  *Test_Child `protobuf:"bytes,9,opt,name=child" json:"child,omitempty"`
	Dict   map[string]string `protobuf:"bytes,10,rep,name=dict" json:"dict,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

// Child 子结构
type Test_Child struct {
	Sex string `protobuf:"bytes,1,opt,name=sex" json:"sex,omitempty"`
}
```
除了会生成对应的结构外，还会有些工具方法，如字段的getter:

```go
func (m *Test) GetAge() int32 {
	if m != nil {
		return m.Age
	}
	return 0
}
```

枚举类型会生成对应名称的常量，同时会有两个map方便使用：

```go
var Test_Status_name = map[int32]string{
	0: "OK",
	1: "FAIL",
}
var Test_Status_value = map[string]int32{
	"OK":   0,
	"FAIL": 1,
}
```

### Service

定义一个简单的Service，`TestService`有一个方法`Test`，接收一个`Request`参数，返回`Response`：

```protobuf
// TestService 测试服务
service TestService {
    // Test 测试方法
	rpc Test(Request) returns (Response) {};
}

// Request 请求结构
message Request {
	string name = 1;
}

// Response 响应结构
message Response {
    string message = 1;
}
```

转换结果：

```go
// 客户端接口
type TestServiceClient interface {
	// Test 测试方法
	Test(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

// 服务端接口
type TestServiceServer interface {
	// Test 测试方法
	Test(context.Context, *Request) (*Response, error)
}
```

生成的go代码中包含该Service定义的接口，客户端接口已经自动实现了，直接供客户端使用者调用，服务端接口需要由服务提供方实现。
