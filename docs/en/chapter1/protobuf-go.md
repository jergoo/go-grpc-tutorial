# Protobuf Go

这里使用一个测试文件对照说明常用结构的protobuf到golang的转换。只说明关键部分代码，详细内容请查看完整文件。示例文件在`proto/test`目录下。


## Package

在proto文件中使用`package`关键字声明包名，默认转换成go中的包名与此一致，如果需要指定不一样的包名，可以使用`go_package`选项：

```protobuf
package test;
option go_package="test";
```

## Message

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

## Service

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
