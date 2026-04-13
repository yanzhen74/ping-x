# TCP 协议测试

<cite>
**本文引用的文件**
- [main.go](file://main.go)
- [root.go](file://cmd/root.go)
- [send.go](file://cmd/send.go)
- [recv.go](file://cmd/recv.go)
- [config_cmd.go](file://cmd/config_cmd.go)
- [config.go](file://internal/config/config.go)
- [protocol.go](file://internal/proto/protocol.go)
- [ssm.go](file://internal/proto/ssm.go)
- [packet.go](file://internal/packet/packet.go)
- [stats.go](file://internal/stats/stats.go)
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
8. [故障排除指南](#故障排除指南)
9. [结论](#结论)
10. [附录](#附录)

## 简介
本文件面向使用 TCP 协议测试功能的用户与开发者，系统阐述 Ping-X 中 TCP 测试的设计与实现思路。内容覆盖 TCP 协议基础（面向连接、可靠传输、三次握手等）、Ping-X 的配置模型与命令行参数、发送/接收模式的运行流程、统计采集与报告、以及性能优化与故障排除建议。

## 项目结构
该项目采用模块化分层设计：
- CLI 层：通过 Cobra 提供命令入口（send/recv/config），负责参数解析与配置加载。
- 配置层：定义测试配置结构与校验逻辑，支持从 YAML 文件批量加载。
- 协议抽象层：定义 Sender/Receiver 接口与工厂方法，按协议类型创建具体实现。
- 数据包与统计层：定义探测包格式与统计收集器，支撑 RTT 计算与报告生成。
- 主程序入口：初始化版本与执行 CLI。

```mermaid
graph TB
subgraph "CLI 层"
ROOT["root.go<br/>根命令与全局标志"]
SEND["send.go<br/>发送模式命令"]
RECV["recv.go<br/>接收模式命令"]
CFGCMD["config_cmd.go<br/>生成示例配置"]
end
subgraph "配置层"
CFG["config.go<br/>TestConfig/FileConfig/校验/解析"]
end
subgraph "协议抽象层"
PROTO["protocol.go<br/>Sender/Receiver 工厂"]
SSM["ssm.go<br/>SSM 占位实现"]
end
subgraph "数据包与统计层"
PKT["packet.go<br/>探测包结构与序列化"]
STATS["stats.go<br/>统计收集器"]
end
MAIN["main.go<br/>主程序入口"]
MAIN --> ROOT
ROOT --> SEND
ROOT --> RECV
ROOT --> CFGCMD
SEND --> CFG
RECV --> CFG
PROTO --> CFG
PROTO --> STATS
PKT --> STATS
CFGCMD --> CFG
```

图表来源
- [main.go:1-11](file://main.go#L1-L11)
- [root.go:14-54](file://cmd/root.go#L14-L54)
- [send.go:11-91](file://cmd/send.go#L11-L91)
- [recv.go:11-72](file://cmd/recv.go#L11-L72)
- [config_cmd.go:10-101](file://cmd/config_cmd.go#L10-L101)
- [config.go:11-125](file://internal/config/config.go#L11-L125)
- [protocol.go:11-51](file://internal/proto/protocol.go#L11-L51)
- [ssm.go:11-21](file://internal/proto/ssm.go#L11-L21)
- [packet.go:17-88](file://internal/packet/packet.go#L17-L88)
- [stats.go:10-110](file://internal/stats/stats.go#L10-L110)

章节来源
- [main.go:1-11](file://main.go#L1-L11)
- [root.go:14-54](file://cmd/root.go#L14-L54)
- [send.go:11-91](file://cmd/send.go#L11-L91)
- [recv.go:11-72](file://cmd/recv.go#L11-L72)
- [config_cmd.go:10-101](file://cmd/config_cmd.go#L10-L101)
- [config.go:11-125](file://internal/config/config.go#L11-L125)
- [protocol.go:11-51](file://internal/proto/protocol.go#L11-L51)
- [ssm.go:11-21](file://internal/proto/ssm.go#L11-L21)
- [packet.go:17-88](file://internal/packet/packet.go#L17-L88)
- [stats.go:10-110](file://internal/stats/stats.go#L10-L110)

## 核心组件
- 配置模型与校验
  - TestConfig 定义了测试项的关键字段：协议、模式、目标/组/源、端口、绑定地址、接口、发送次数、间隔、超时、负载大小、TTL 等。
  - 校验规则确保协议合法、模式为 send/recv、端口范围有效；send 模式下 TCP/UDP 需要目标地址，多播/SSM 需要组地址；recv 模式下多播/SSM 需要组地址；时间格式需可解析。
- 协议抽象与工厂
  - Sender/Receiver 接口定义了统一的发送/接收行为。
  - NewSender/NewReceiver 工厂按协议类型返回对应实现；当前 TCP 的具体实现标记为占位，尚未完成实际网络交互。
- 探测包与统计
  - Probe 定义了固定头部（魔数、序号、时间戳、负载长度）与可选负载，支持序列化/反序列化与 RTT 计算。
  - Collector 提供并发安全的统计记录与报告生成，包括发送/接收/丢包、RTT 最小/平均/最大。
- CLI 命令
  - send/recv 子命令支持从命令行参数或配置文件加载配置，并进行基本校验与打印。
  - config 子命令输出示例 YAML，覆盖 TCP/UDP/多播/SSM 的典型用法。

章节来源
- [config.go:11-125](file://internal/config/config.go#L11-L125)
- [protocol.go:11-51](file://internal/proto/protocol.go#L11-L51)
- [ssm.go:11-21](file://internal/proto/ssm.go#L11-L21)
- [packet.go:17-88](file://internal/packet/packet.go#L17-L88)
- [stats.go:10-110](file://internal/stats/stats.go#L10-L110)
- [send.go:11-91](file://cmd/send.go#L11-L91)
- [recv.go:11-72](file://cmd/recv.go#L11-L72)
- [config_cmd.go:10-101](file://cmd/config_cmd.go#L10-L101)

## 架构总览
下图展示了 TCP 测试在 Ping-X 中的高层架构：CLI 解析参数 → 加载配置 → 校验 → 通过协议工厂创建 TCP Sender/Receiver → 使用探测包与统计器完成测试 → 输出结果。

```mermaid
graph TB
CLI["CLI 命令<br/>send/recv/config"] --> CFG["配置加载与校验"]
CFG --> FACTORY["协议工厂<br/>NewSender/NewReceiver"]
FACTORY --> TCP_SEND["TCP Sender<br/>占位实现"]
FACTORY --> TCP_RECV["TCP Receiver<br/>占位实现"]
TCP_SEND --> PKT["探测包<br/>Probe"]
TCP_RECV --> PKT
PKT --> STATS["统计收集器<br/>Collector"]
STATS --> REPORT["统计报告"]
```

图表来源
- [send.go:34-91](file://cmd/send.go#L34-L91)
- [recv.go:29-72](file://cmd/recv.go#L29-L72)
- [config.go:34-125](file://internal/config/config.go#L34-L125)
- [protocol.go:21-51](file://internal/proto/protocol.go#L21-L51)
- [packet.go:17-88](file://internal/packet/packet.go#L17-L88)
- [stats.go:10-110](file://internal/stats/stats.go#L10-L110)

## 详细组件分析

### 配置模型与校验（TestConfig）
- 字段说明
  - 名称、协议、模式、目标主机、组地址、源地址、端口、绑定地址、网络接口、发送次数、发送间隔、超时、负载大小、TTL。
- 校验要点
  - 协议必须为 tcp/udp/multicast/ssm。
  - 模式必须为 send 或 recv。
  - 端口范围 1–65535。
  - send 模式：
    - TCP/UDP 要求目标地址非空；
    - 多播/SSM 要求组地址非空。
  - recv 模式：
    - 多播/SSM 要求组地址非空。
  - 时间格式（interval/timeout）需可解析。
- 默认值
  - 未设置间隔时采用 1 秒；
  - 未设置超时时采用 3 秒。

```mermaid
flowchart TD
Start(["进入 Validate"]) --> CheckProto["校验协议是否为 tcp/udp/multicast/ssm"]
CheckProto --> ProtoOK{"协议合法？"}
ProtoOK --> |否| ErrProto["返回错误：协议非法"]
ProtoOK --> |是| CheckMode["校验模式是否为 send/recv"]
CheckMode --> ModeOK{"模式合法？"}
ModeOK --> |否| ErrMode["返回错误：模式非法"]
ModeOK --> |是| CheckPort["校验端口范围 1..65535"]
CheckPort --> PortOK{"端口合法？"}
PortOK --> |否| ErrPort["返回错误：端口非法"]
PortOK --> ModeType{"模式类型"}
ModeType --> |send| CheckSend["send 模式校验"]
ModeType --> |recv| CheckRecv["recv 模式校验"]
CheckSend --> TargetOrGroup["TCP/UDP 需要目标；多播/SSM 需要组"]
TargetOrGroup --> SendOK{"send 校验通过？"}
SendOK --> |否| ErrSend["返回错误：缺少必要字段"]
SendOK --> |是| ParseDur["解析 interval/timeout"]
CheckRecv --> RecvOK["recv 校验通过"]
ParseDur --> DurOK{"格式可解析？"}
DurOK --> |否| ErrDur["返回错误：时间格式非法"]
DurOK --> Done(["返回成功"])
```

图表来源
- [config.go:49-125](file://internal/config/config.go#L49-L125)

章节来源
- [config.go:11-125](file://internal/config/config.go#L11-L125)

### 协议抽象与工厂（Sender/Receiver）
- 接口职责
  - Sender.Send(ctx, cfg, stat)：执行发送逻辑，记录统计。
  - Receiver.Receive(ctx, cfg)：执行接收逻辑。
- 工厂方法
  - NewSender 根据协议返回 TCPSender、UDPSender、MulticastSender、SSMSender；
  - NewReceiver 根据协议返回 TCPReceiver、UDPReceiver、MulticastReceiver、SSMReceiver。
- 当前状态
  - TCP 的具体实现为占位，尚未接入网络交互；SSM 同样为占位。

```mermaid
classDiagram
class Sender {
+Send(ctx, cfg, stat) error
}
class Receiver {
+Receive(ctx, cfg) error
}
class TCPSender
class TCPReceiver
class UDPSender
class UDPReceiver
class MulticastSender
class MulticastReceiver
class SSMSender
class SSMReceiver
Sender <|.. TCPSender
Sender <|.. UDPSender
Sender <|.. MulticastSender
Sender <|.. SSMSender
Receiver <|.. TCPReceiver
Receiver <|.. UDPReceiver
Receiver <|.. MulticastReceiver
Receiver <|.. SSMReceiver
```

图表来源
- [protocol.go:11-51](file://internal/proto/protocol.go#L11-L51)
- [ssm.go:11-21](file://internal/proto/ssm.go#L11-L21)

章节来源
- [protocol.go:11-51](file://internal/proto/protocol.go#L11-L51)
- [ssm.go:11-21](file://internal/proto/ssm.go#L11-L21)

### 探测包与统计（Probe/Collector）
- 探测包（Probe）
  - 固定头部包含魔数、序号、时间戳、负载长度；
  - 负载为可变长度字节序列，用于校验与填充；
  - 支持序列化/反序列化、RTT 计算、ACK 包构造。
- 统计收集器（Collector）
  - 并发安全地记录发送/接收/丢包；
  - 维护 RTT 列表与最小/最大/平均值；
  - 生成人类可读的统计报告。

```mermaid
sequenceDiagram
participant S as "发送端"
participant P as "探测包"
participant R as "接收端"
participant C as "统计收集器"
S->>P : "创建 Probe(含序号/时间戳/负载)"
S->>R : "发送 Probe"
R->>P : "接收并校验"
R->>C : "RecordReceive(RTT)"
R->>S : "发送 ACK(仅头部)"
S->>C : "RecordSend/RecordReceive(若收到 ACK)"
C-->>S : "生成统计报告"
```

图表来源
- [packet.go:17-88](file://internal/packet/packet.go#L17-L88)
- [stats.go:10-110](file://internal/stats/stats.go#L10-L110)

章节来源
- [packet.go:17-88](file://internal/packet/packet.go#L17-L88)
- [stats.go:10-110](file://internal/stats/stats.go#L10-L110)

### CLI 命令与运行流程
- send 子命令
  - 支持从命令行参数或配置文件加载；
  - 校验通过后打印当前配置摘要；
  - 实际协议调用预留（后续任务中实现）。
- recv 子命令
  - 支持从命令行参数或配置文件加载；
  - 校验通过后打印监听配置摘要；
  - 实际协议调用预留（后续任务中实现）。
- config 子命令
  - 输出包含 TCP/UDP/多播/SSM 的示例配置，便于快速上手。

```mermaid
sequenceDiagram
participant U as "用户"
participant CLI as "CLI"
participant CFG as "配置加载"
participant VAL as "配置校验"
participant RUN as "运行逻辑"
U->>CLI : "ping-x send/recv --flags"
CLI->>CFG : "加载配置(文件或参数)"
CFG->>VAL : "Validate()"
VAL-->>CLI : "通过/失败"
CLI->>RUN : "执行 send/recv 逻辑"
RUN-->>U : "输出结果/统计"
```

图表来源
- [send.go:34-91](file://cmd/send.go#L34-L91)
- [recv.go:29-72](file://cmd/recv.go#L29-L72)
- [config_cmd.go:87-101](file://cmd/config_cmd.go#L87-L101)

章节来源
- [send.go:11-91](file://cmd/send.go#L11-L91)
- [recv.go:11-72](file://cmd/recv.go#L11-L72)
- [config_cmd.go:10-101](file://cmd/config_cmd.go#L10-L101)

## 依赖分析
- 外部依赖
  - Cobra：命令行框架；
  - Viper：配置管理；
  - YAML：配置文件解析；
  - Go 标准库：时间、并发、编码等。
- 模块间耦合
  - CLI 层依赖配置层与协议抽象层；
  - 协议抽象层依赖配置层与统计层；
  - 数据包层与统计层相互独立但被上层复用。

```mermaid
graph LR
GO_MOD["go.mod"] --> COBRA["github.com/spf13/cobra"]
GO_MOD --> VIPER["github.com/spf13/viper"]
GO_MOD --> YAML["gopkg.in/yaml.v2"]
MAIN["main.go"] --> ROOT_CMD["cmd/root.go"]
ROOT_CMD --> SEND_CMD["cmd/send.go"]
ROOT_CMD --> RECV_CMD["cmd/recv.go"]
SEND_CMD --> CFG["internal/config/config.go"]
RECV_CMD --> CFG
CFG --> PROTO["internal/proto/protocol.go"]
PROTO --> STATS["internal/stats/stats.go"]
PROTO --> PKT["internal/packet/packet.go"]
```

图表来源
- [go.mod:1-25](file://go.mod#L1-L25)
- [main.go:1-11](file://main.go#L1-L11)
- [root.go:3-6](file://cmd/root.go#L3-L6)
- [send.go:7-8](file://cmd/send.go#L7-L8)
- [recv.go:7-8](file://cmd/recv.go#L7-L8)
- [config.go:3-8](file://internal/config/config.go#L3-L8)
- [protocol.go:3-8](file://internal/proto/protocol.go#L3-L8)
- [stats.go:3-8](file://internal/stats/stats.go#L3-L8)
- [packet.go:3-8](file://internal/packet/packet.go#L3-L8)

章节来源
- [go.mod:1-25](file://go.mod#L1-L25)

## 性能考虑
- 负载大小与 MTU
  - 探测包包含固定头部与可变负载，增大负载会增加 CPU 开销与网络开销。建议根据链路特性调整 size，避免超过路径 MTU 导致分片。
- 发送间隔与速率
  - interval 过短会增加 CPU 与网络压力；过长则降低检测灵敏度。结合 timeout 与 count 合理设置。
- 并发与锁
  - 统计收集器使用互斥锁保护共享状态，避免高并发下的竞争条件。在高频场景下，可考虑减少统计频率或使用更细粒度的采样策略。
- 超时与重传
  - 超时时间影响 RTT 测量与丢包判断。在网络延迟较大或抖动明显时，适当增大 timeout 可减少误判。
- 接口与绑定
  - 在多网卡环境中，合理设置 bind 与 iface 可避免路由与 NAT 引起的异常。

## 故障排除指南
- 协议/模式/端口错误
  - 现象：启动时报错“协议非法/模式非法/端口非法”。
  - 处理：检查配置中的 proto/mode/port 是否符合要求。
- 缺少必要字段
  - 现象：send 模式下 TCP/UDP 报告缺少目标地址；多播/SSM 报告缺少组地址。
  - 处理：补齐 target/group 字段。
- 时间格式错误
  - 现象：interval/timeout 格式不被识别。
  - 处理：使用可被解析的时间格式（如 1s、2.5s、100ms）。
- 配置文件加载失败
  - 现象：提示无法读取/解析配置文件。
  - 处理：检查文件路径与权限，确认 YAML 格式正确。
- TCP 功能未实现
  - 现象：TCP 模式下提示“占位实现，尚未完成”。
  - 处理：等待后续版本实现，或参考 UDP/多播/SSM 的配置作为替代方案。

章节来源
- [config.go:49-125](file://internal/config/config.go#L49-L125)
- [send.go:34-91](file://cmd/send.go#L34-L91)
- [recv.go:29-72](file://cmd/recv.go#L29-L72)
- [ssm.go:14-20](file://internal/proto/ssm.go#L14-L20)

## 结论
Ping-X 的 TCP 协议测试功能基于清晰的配置模型与协议抽象层设计，具备良好的扩展性。当前 TCP 的具体实现处于占位阶段，后续可在现有接口与数据包/统计设施基础上完成网络交互与测试闭环。通过合理的配置与参数调优，用户可以在不同网络环境下进行可靠的连通性与性能评估。

## 附录

### TCP 协议特点与工作原理
- 面向连接
  - 建立连接需要三次握手，释放连接需要四次挥手。
- 可靠传输
  - 通过序号、确认应答、重传、流量控制与拥塞控制保证可靠性。
- 传输控制
  - 拥有端口、窗口、校验和等机制，确保有序、无差错的数据交付。
- 适用场景
  - 对可靠性要求高的应用，如 Web、数据库、文件传输等。

### TCP 测试配置选项详解
- 基本参数
  - 协议：tcp（当前占位实现）。
  - 模式：send（发送端）、recv（接收端）。
  - 目标主机：TCP/UDP 下必填。
  - 组地址：多播/SSM 下必填。
  - 源地址：SSM 下必填。
  - 端口：必填，范围 1–65535。
  - 绑定地址：本地绑定地址，默认 0.0.0.0。
  - 网络接口：多播/SSM 场景可指定。
- 控制参数
  - 发送次数：0 表示持续发送。
  - 发送间隔：时间格式，如 1s、2.5s。
  - 超时时间：单包超时，如 3s。
  - 负载大小：字节数，决定探测包总长度。
  - TTL：多播场景建议 16，其他场景由系统自动选择。

章节来源
- [config.go:11-125](file://internal/config/config.go#L11-L125)

### 使用场景与配置示例
- 客户端连接测试（send）
  - 目标：业务服务器 IP，端口：服务监听端口，模式：send，发送次数与间隔视需求设置。
- 服务器监听验证（recv）
  - 目标：本机监听地址，端口：服务监听端口，模式：recv，验证是否可接收探测包并回显确认。
- 多播/SSM 场景
  - 参考示例配置中的多播与 SSM 条目，分别在发送端与接收端配置对应的组与源地址。

章节来源
- [config_cmd.go:22-85](file://cmd/config_cmd.go#L22-L85)