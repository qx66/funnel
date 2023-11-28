# funnel (漏斗)

一个通用的埋点处理系统 (将埋点数据吞吐到Kafka)

## 背景

大部分公司都会有类似的埋点数据收集系统

### 通用做法

Openresty 编写 Lua 脚本，收集对应的数据，然后传到 Kafka 系统中。 (不保证数据不丢失)

### funnel

funnel 是通用做法的一个替代方案，负载均衡器将请求代理到 funnel 服务中，
funnel 服务将数据缓存到本地，然后批量发送到 kafka 服务中。

## 配置

BatchTimeout、ReadTimeout、WriteTimeout 单位为 ms.

important:
    在执行 make config 之后, 一定要将以下代码在 internal/conf/conf.pb.go 文件替换

```golang
type Kafka struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BootstrapServer []string `protobuf:"bytes,1,rep,name=bootstrapServer,proto3" json:"bootstrapServer,omitempty" yaml:"bootstrapServer,omitempty"`
	BatchTimeout    int32    `protobuf:"varint,2,opt,name=batchTimeout,proto3" json:"batchTimeout,omitempty" yaml:"batchTimeout,omitempty"`
	ReadTimeout     int32    `protobuf:"varint,3,opt,name=readTimeout,proto3" json:"readTimeout,omitempty" yaml:"readTimeout,omitempty"`
	WriteTimeout    int32    `protobuf:"varint,4,opt,name=writeTimeout,proto3" json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
}
```

## 收集的数据

收集的数据:

    1. get query
    2. post json
    3. post form
    4. header
    5. cookie

## Usage

```shell
./bin/funnel-linux -configPath=./configs/config.yaml
```