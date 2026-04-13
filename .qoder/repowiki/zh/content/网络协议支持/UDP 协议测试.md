# UDP 协议测试

<cite>
**本文引用的文件**
- [main.go](file://main.go)
- [root.go](file://cmd/root.go)
- [config.go](file://internal/config/config.go)
- [protocol.go](file://internal/proto/protocol.go)
- [packet.go](file://internal/packet/packet.go)
- [stats.go](file://internal/stats/stats.go)
- [send.go](file://cmd/send.go)
- [recv.go](file://cmd/recv.go)
- [config_cmd.go](file://cmd/config_cmd.go)
- [version.go](file://cmd/version.go)
- [go.mod](file://go.mod)
</cite>

## 目录
1. [简介](#简介)
2. [项目结构](#项目结构)
3. [核心组件](#核心组件)
4. [架构总览](#架构总览)
5. [详细组件分析](#详细组件分析)
6. [依赖分析](#依赖分析)
7. [性能考虑](#性能考虑)
8. [故障排查指南](#故障排查指南)
9. [结论](#结论)
10. [附录](#附录)

## 简介
本文件面向“UDP 协议测试”功能，系统阐述其工作原理与实现机制。基于当前仓库的结构与实现，UDP 测试通过命令行子命令以发送/接收两种模式运行，结合统一的配置模型与探测包格式，完成对 UDP 连通性的验证与统计。同时，文档解释 UDP 的无连接、不可靠、低开销等特性，并提供配置项说明、典型使用场景、性能优化建议与调试技巧。

## 项目结构
项目采用模块化分层组织：
- cmd 层：命令入口与子命令（send/recv/config/version），负责参数解析与流程编排
- internal/config：配置模型与校验逻辑
- internal/packet：探测包结构与序列化/反序列化
- internal/proto：协议抽象与工厂方法（含 UDP Sender/Receiver 占位）
- internal/stats：统计收集与报告
- 根目录：程序入口与版本注入

```mermaid
graph TB
subgraph "命令层(cmd)"
ROOT["root.go<br/>根命令与全局标志"]
SEND["send.go<br/>发送模式子命令"]
RECV["recv.go<br/>接收模式子命令"]
CFG["config_cmd.go<br/>生成示例配置"]
VER["version.go<br/>版本命令"]
end
subgraph "内部模块(internal)"
CONF["config.go<br/>配置模型与校验"]
PKT["packet.go<br/>探测包结构与序列化"]
PROTO["protocol.go<br/>协议接口与工厂"]
STATS["stats.go<br/>统计收集器"]
end
MAIN["main.go<br/>入口与版本注入"]
MAIN --> ROOT
ROOT --> SEND
ROOT --> RECV
ROOT --> CFG
ROOT --> VER
SEND --> CONF
RECV --> CONF
SEND --> PROTO
RECV --> PROTO
SEND --> STATS
SEND --> PKT
RECV --> PKT
```

图表来源
- [main.go:1-11](file://main.go#L1-L11)
- [root.go:14-25](file://cmd/root.go#L14-L25)
- [send.go:11-16](file://cmd/send.go#L11-L16)
- [recv.go:11-16](file://cmd/recv.go#L11-L16)
- [config.go:11-27](file://internal/config/config.go#L11-L27)
- [packet.go:17-24](file://internal/packet/packet.go#L17-L24)
- [protocol.go:11-19](file://internal/proto/protocol.go#L11-L19)
- [stats.go:10-20](file://internal/stats/stats.go#L10-L20)

章节来源
- [main.go:1-11](file://main.go#L1-L11)
- [root.go:14-25](file://cmd/root.go#L14-L25)

## 核心组件
- 配置模型与校验：统一的 TestConfig 结构体承载协议、模式、目标、端口、绑定、接口、计数、间隔、超时、包大小、TTL 等字段；提供 Validate 与默认间隔/超时解析
- 探测包格式：固定头部包含魔数、序号、时间戳、载荷长度，支持序列化/反序列化与 RTT 计算
- 协议工厂：按协议类型创建发送/接收器接口实例（UDP 占位已就绪）
- 统计收集：记录发送/接收/丢包与 RTT 列表，计算最小/最大/平均 RTT 与丢包率
- 命令行：send/recv 子命令解析参数并进行配置校验；config 子命令输出示例配置；version 输出版本

章节来源
- [config.go:11-27](file://internal/config/config.go#L11-L27)
- [config.go:49-100](file://internal/config/config.go#L49-L100)
- [packet.go:17-24](file://internal/packet/packet.go#L17-L24)
- [packet.go:42-72](file://internal/packet/packet.go#L42-L72)
- [protocol.go:21-35](file://internal/proto/protocol.go#L21-L35)
- [stats.go:10-20](file://internal/stats/stats.go#L10-L20)
- [stats.go:58-76](file://internal/stats/stats.go#L58-L76)
- [send.go:21-32](file://cmd/send.go#L21-L32)
- [recv.go:21-27](file://cmd/recv.go#L21-L27)
- [config_cmd.go:22-75](file://cmd/config_cmd.go#L22-L75)

## 架构总览
UDP 测试在命令层与内部模块之间形成清晰边界：命令层负责输入解析与流程控制，内部模块负责协议抽象、数据包格式与统计。发送/接收模式共享同一配置模型与探测包格式，通过协议工厂选择具体实现。

```mermaid
graph TB
CLI_SEND["CLI 发送模式<br/>send.go"] --> CFG_LOAD["配置加载/校验<br/>config.go"]
CLI_RECV["CLI 接收模式<br/>recv.go"] --> CFG_LOAD
CFG_LOAD --> FACTORY["协议工厂<br/>protocol.go"]
FACTORY --> UDP_SEND["UDP 发送器占位<br/>protocol.go"]
FACTORY --> UDP_RECV["UDP 接收器占位<br/>protocol.go"]
UDP_SEND --> PKT_ENC["探测包编码<br/>packet.go"]
UDP_RECV --> PKT_DEC["探测包解码<br/>packet.go"]
UDP_SEND --> STATS_COL["统计收集<br/>stats.go"]
UDP_RECV --> STATS_COL
```

图表来源
- [send.go:34-91](file://cmd/send.go#L34-L91)
- [recv.go:29-72](file://cmd/recv.go#L29-L72)
- [config.go:49-100](file://internal/config/config.go#L49-L100)
- [protocol.go:21-51](file://internal/proto/protocol.go#L21-L51)
- [packet.go:42-72](file://internal/packet/packet.go#L42-L72)
- [stats.go:29-49](file://internal/stats/stats.go#L29-L49)

## 详细组件分析

### 配置模型与校验（UDP 关键字段）
- 协议 proto：支持 tcp/udp/multicast/ssm
- 模式 mode：send/recv
- 目标 target：仅 udp/tcp send 必填
- 组 group：仅 multicast/ssm send/recv 必填
- 端口 port：1-65535
- 绑定 bind：本地绑定地址
- 接口 iface：多播时使用
- 计数 count：发送次数（0 表示持续）
- 间隔 interval：发送间隔（支持时长格式）
- 超时 timeout：单包超时（支持时长格式）
- 包大小 size：探测包载荷大小（bytes）
- TTL：TTL（0 自动选择）

```mermaid
classDiagram
class TestConfig {
+string Name
+string Proto
+string Mode
+string Target
+string Group
+string Source
+int Port
+string Bind
+string Iface
+int Count
+string Interval
+string Timeout
+int Size
+int TTL
+Validate() error
+GetInterval() time.Duration
+GetTimeout() time.Duration
}
```

图表来源
- [config.go:11-27](file://internal/config/config.go#L11-L27)
- [config.go:49-100](file://internal/config/config.go#L49-L100)

章节来源
- [config.go:11-27](file://internal/config/config.go#L11-L27)
- [config.go:49-100](file://internal/config/config.go#L49-L100)

### 探测包格式与处理
- 固定头部：魔数、序号、时间戳、载荷长度
- 载荷：按 size 动态填充
- 编码/解码：大端序二进制序列化/反序列化
- RTT 计算：基于发送时间戳
- 确认包：保留原始序号与时间戳，载荷为空

```mermaid
classDiagram
class Probe {
+uint32 Magic
+uint32 Seq
+int64 Timestamp
+uint16 Size
+[]byte Payload
+Marshal() []byte,error
+Unmarshal([]byte) *Probe,error
+RTT() time.Duration
+MakeAck() *Probe
}
```

图表来源
- [packet.go:17-24](file://internal/packet/packet.go#L17-L24)
- [packet.go:42-72](file://internal/packet/packet.go#L42-L72)

章节来源
- [packet.go:17-24](file://internal/packet/packet.go#L17-L24)
- [packet.go:42-72](file://internal/packet/packet.go#L42-L72)

### 协议工厂与 UDP 占位
- NewSender/NewReceiver：按协议类型返回对应实现
- UDP 已在工厂中注册（占位待实现）
- 其他协议（TCP/Multicast/SSM）亦有占位

```mermaid
classDiagram
class Sender {
<<interface>>
+Send(ctx, cfg, stat) error
}
class Receiver {
<<interface>>
+Receive(ctx, cfg) error
}
class UDPSender
class UDPReceiver
Sender <|.. UDPSender
Receiver <|.. UDPReceiver
```

图表来源
- [protocol.go:11-19](file://internal/proto/protocol.go#L11-L19)
- [protocol.go:21-51](file://internal/proto/protocol.go#L21-L51)

章节来源
- [protocol.go:21-51](file://internal/proto/protocol.go#L21-L51)

### 统计收集器
- 记录发送/接收/丢包
- 维护 RTT 列表，计算最小/最大/平均 RTT
- 生成统计报告字符串

```mermaid
classDiagram
class Collector {
-Mutex mu
+int Sent
+int Received
+int Lost
+[]Duration RTTs
+Duration MinRTT
+Duration MaxRTT
+Duration SumRTT
+RecordSend()
+RecordReceive(Duration)
+RecordLoss()
+LossRate() float64
+AvgRTT() Duration
+Report(string,string) string
}
```

图表来源
- [stats.go:10-20](file://internal/stats/stats.go#L10-L20)
- [stats.go:58-76](file://internal/stats/stats.go#L58-L76)
- [stats.go:78-93](file://internal/stats/stats.go#L78-L93)

章节来源
- [stats.go:10-20](file://internal/stats/stats.go#L10-L20)
- [stats.go:58-76](file://internal/stats/stats.go#L58-L76)
- [stats.go:78-93](file://internal/stats/stats.go#L78-L93)

### 发送模式流程（UDP）
```mermaid
sequenceDiagram
participant U as "用户"
participant CMD as "send.go"
participant CFG as "config.go"
participant FAC as "protocol.go"
participant PKT as "packet.go"
participant ST as "stats.go"
U->>CMD : "执行 send 子命令"
CMD->>CFG : "构建/加载配置并校验"
CFG-->>CMD : "校验结果"
CMD->>FAC : "NewSender('udp')"
FAC-->>CMD : "返回 UDPSender 占位"
CMD->>PKT : "构造探测包(含序号/时间戳/载荷)"
CMD->>ST : "记录发送"
CMD-->>U : "开始发送(占位逻辑)"
```

图表来源
- [send.go:34-91](file://cmd/send.go#L34-L91)
- [config.go:49-100](file://internal/config/config.go#L49-L100)
- [protocol.go:21-35](file://internal/proto/protocol.go#L21-L35)
- [packet.go:26-40](file://internal/packet/packet.go#L26-L40)
- [stats.go:29-34](file://internal/stats/stats.go#L29-L34)

### 接收模式流程（UDP）
```mermaid
sequenceDiagram
participant U as "用户"
participant CMD as "recv.go"
participant CFG as "config.go"
participant FAC as "protocol.go"
participant PKT as "packet.go"
participant ST as "stats.go"
U->>CMD : "执行 recv 子命令"
CMD->>CFG : "构建/加载配置并校验"
CFG-->>CMD : "校验结果"
CMD->>FAC : "NewReceiver('udp')"
FAC-->>CMD : "返回 UDPReceiver 占位"
CMD->>PKT : "解码探测包并校验魔数/长度"
CMD->>PKT : "生成确认包"
CMD->>ST : "记录接收/RTT"
CMD-->>U : "等待/处理(占位逻辑)"
```

图表来源
- [recv.go:29-72](file://cmd/recv.go#L29-L72)
- [config.go:49-100](file://internal/config/config.go#L49-L100)
- [protocol.go:37-51](file://internal/proto/protocol.go#L37-L51)
- [packet.go:53-72](file://internal/packet/packet.go#L53-L72)
- [stats.go:36-49](file://internal/stats/stats.go#L36-L49)

### UDP 配置选项详解
- 协议 proto：udp
- 模式 mode：send 或 recv
- 目标 target：udp/tcp send 必填
- 端口 port：必填（1-65535）
- 绑定 bind：本地绑定地址（默认 0.0.0.0）
- 接口 iface：多播时使用
- 计数 count：发送次数（0 表示持续）
- 间隔 interval：发送间隔（支持时长格式）
- 超时 timeout：单包超时（支持时长格式）
- 包大小 size：探测包载荷大小（bytes）
- TTL：TTL（0 自动选择：unicast=64, multicast=16）

章节来源
- [send.go:21-32](file://cmd/send.go#L21-L32)
- [recv.go:21-27](file://cmd/recv.go#L21-L27)
- [config.go:11-27](file://internal/config/config.go#L11-L27)
- [config.go:62-65](file://internal/config/config.go#L62-L65)

### 使用场景与配置示例
- 无连接通信测试（点对点 UDP）
  - 发送端：proto=udp, mode=send, target=远端IP, port=远端端口, size=64
  - 接收端：proto=udp, mode=recv, port=监听端口
- 广播/多播测试（组播）
  - 接收端：proto=multicast, mode=recv, group=多播组, port=端口, iface=网卡
  - 发送端：proto=multicast, mode=send, group=多播组, port=端口, iface=网卡, count=20
- SSM（源特定组播）测试
  - 接收端：proto=ssm, mode=recv, group=组播组, source=源地址, port=端口, iface=网卡

示例配置片段路径
- [config_cmd.go:22-75](file://cmd/config_cmd.go#L22-L75)

章节来源
- [config_cmd.go:22-75](file://cmd/config_cmd.go#L22-L75)

### UDP 协议特点
- 无连接：无需握手建立连接，直接发送/接收
- 不可靠：不保证到达、顺序与去重
- 低开销：无连接状态维护，报头更短
- 适用场景：实时性要求高、可容忍少量丢包的应用（如音视频、监控）

## 依赖分析
- 外部依赖：Cobra/Viper/YAML 等，用于命令行与配置管理
- 内部耦合：命令层依赖配置与协议工厂；协议工厂依赖配置与统计；探测包与统计被发送/接收流程复用

```mermaid
graph LR
SEND["send.go"] --> CONF["config.go"]
SEND --> PROTO["protocol.go"]
SEND --> PKT["packet.go"]
SEND --> STATS["stats.go"]
RECV["recv.go"] --> CONF
RECV --> PROTO
RECV --> PKT
RECV --> STATS
CFG_CMD["config_cmd.go"] --> CONF
```

图表来源
- [send.go:34-91](file://cmd/send.go#L34-L91)
- [recv.go:29-72](file://cmd/recv.go#L29-L72)
- [config.go:49-100](file://internal/config/config.go#L49-L100)
- [protocol.go:21-51](file://internal/proto/protocol.go#L21-L51)
- [packet.go:42-72](file://internal/packet/packet.go#L42-L72)
- [stats.go:29-49](file://internal/stats/stats.go#L29-L49)
- [config_cmd.go:22-75](file://cmd/config_cmd.go#L22-L75)

章节来源
- [go.mod:5-24](file://go.mod#L5-L24)

## 性能考虑
- 包大小 size：增大载荷会增加网络与 CPU 开销，需权衡吞吐与延迟
- 间隔 interval：过短会导致 CPU/带宽压力增大，过长影响检测灵敏度
- 计数 count：持续发送会占用资源，建议按场景设定上限
- TTL：合理设置可避免不必要的路由跳数
- 统计开销：RTT 计算与数组追加应避免在高频路径中产生阻塞

## 故障排查指南
- 配置校验失败
  - 协议非法：确保 proto 为 tcp/udp/multicast/ssm
  - 模式非法：mode 必须为 send 或 recv
  - 端口越界：port 必须在 1-65535
  - send 必填项缺失：udp/tcp send 需要 target；multicast/ssm send 需要 group
  - recv 必填项缺失：multicast/ssm recv 需要 group
  - 时间格式错误：interval/timeout 必须是合法时长格式
- 探测包异常
  - 魔数不匹配：检查发送/接收两端探测包格式是否一致
  - 数据过短：确认包头长度与载荷长度计算正确
- 网络问题
  - 绑定/接口：确认 bind 与 iface 配置正确
  - 多播：确认组播路由与 IGMP 配置
- 版本与输出
  - 使用 version 命令核对版本
  - 使用 config 子命令生成示例配置进行对照

章节来源
- [config.go:49-100](file://internal/config/config.go#L49-L100)
- [packet.go:53-72](file://internal/packet/packet.go#L53-L72)
- [version.go:9-19](file://cmd/version.go#L9-L19)
- [config_cmd.go:22-75](file://cmd/config_cmd.go#L22-L75)

## 结论
本项目为 UDP 协议测试提供了清晰的命令行接口与统一的配置模型，探测包格式与统计模块完善，协议工厂预留了扩展空间。当前 UDP 的发送/接收实现为占位，后续可在相应接口中接入实际网络操作，即可完成端到端的 UDP 连通性测试与性能评估。

## 附录
- 命令行参考
  - 发送模式：send 子命令，支持 proto/target/group/port/bind/iface/count/interval/timeout/size/ttl
  - 接收模式：recv 子命令，支持 proto/bind/port/group/source/iface
  - 生成示例配置：config 子命令，输出 tests 数组示例
  - 版本：version 子命令，打印版本信息
- 参考实现路径
  - 发送流程：[send.go:34-91](file://cmd/send.go#L34-L91)
  - 接收流程：[recv.go:29-72](file://cmd/recv.go#L29-L72)
  - 配置模型与校验：[config.go:11-100](file://internal/config/config.go#L11-L100)
  - 探测包格式：[packet.go:17-88](file://internal/packet/packet.go#L17-L88)
  - 协议工厂：[protocol.go:21-51](file://internal/proto/protocol.go#L21-L51)
  - 统计收集：[stats.go:10-110](file://internal/stats/stats.go#L10-L110)