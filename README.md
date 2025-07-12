# WebSocket 广播器

一个用Go实现的WebSocket消息广播中继服务器，支持灵活的规则配置和消息转换。

## 项目结构

```
websocket_broadcaster/
├── main.go                    # 主程序入口
├── go.mod                     # Go模块定义
├── go.sum                     # 依赖校验文件
├── configs/                   # 配置文件目录
│   ├── config.yaml           # 默认配置文件
│   ├── config.example.yaml   # 配置示例文件
│   └── config.simple.yaml    # 简化配置文件
├── internal/                  # 内部包（私有）
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── message/              # 消息处理
│   │   └── message.go
│   └── connection/           # 连接管理
│       └── connection.go
├── pkg/                      # 公共包
│   └── utils/               # 工具函数
│       └── utils.go
├── main_test.go              # 单元测试
├── build.sh                  # Linux/Mac构建脚本
├── build.bat                 # Windows构建脚本
├── Makefile                  # Make构建配置
└── README.md                 # 项目文档
```

## 模块说明

### 🏗️ main.go
程序入口，负责：
- 启动HTTP服务器
- 加载配置文件
- 优雅关闭处理

### � internal/ 内部包

#### �📨 internal/message/
消息相关的结构和方法：
- `Message` 和 `Body` 结构体定义
- 消息内容获取、克隆、验证等方法
- 时间戳更新功能

#### ⚙️ internal/config/
配置文件处理：
- 配置结构体定义（规则、群组、服务器等）
- YAML配置文件加载
- 配置验证和辅助方法

#### 🔗 internal/connection/
WebSocket连接管理：
- 连接管理器实现
- WebSocket升级和处理
- 消息广播逻辑
- 规则匹配和转换

### 📁 pkg/ 公共包

#### 🛠️ pkg/utils/
通用工具函数：
- 模式匹配函数
- 字符串数组操作
- 消息类型验证

### 🧪 main_test.go
单元测试：
- 核心功能测试
- 边界情况验证
- 性能测试（可扩展）

## 功能特性

- 🔄 **智能转发**: 基于规则的消息路由和转发
- 🎯 **精确匹配**: 支持来源、目标、消息类型的多维度过滤
- 🔧 **消息转换**: 支持添加前缀、修改来源等消息转换
- ⚡ **高性能**: Go语言实现，支持高并发连接
- 📝 **配置驱动**: YAML配置文件，支持热重载
- 🔗 **多平台**: 支持游戏服务器、QQ、Discord等多平台互通

## 快速开始

### 1. 编译项目

```bash
cd websocket_broadcaster
go mod tidy
go build -o broadcaster main.go
```

### 2. 配置文件

编辑 `configs/config.yaml` 文件，配置服务器地址和广播规则：

```yaml
server:
  host: "0.0.0.0"
  port: "8765" 
  path: "/ws"

rules:
  - name: "转发聊天消息"
    from_sources: ["*"]
    to_targets: ["*"]
    message_types: ["chat"]
    enabled: true
```

### 3. 启动服务器

```bash
# 使用默认配置文件
./broadcaster

# 指定配置文件
./broadcaster custom_config.yaml
```

## 配置说明

### 服务器配置

```yaml
server:
  host: "0.0.0.0"      # 监听地址
  port: "8765"         # 监听端口  
  path: "/ws"          # WebSocket路径
```

### 群组配置（推荐）

对于多平台互通场景，推荐使用群组配置，更简洁易懂：

```yaml
groups:
  - name: "全平台互通"
    members:
      - "minecraft"
      - "qq_bot"
      - "discord_bot"
      - "telegram_bot"
    message_types: ["chat"]
    enabled: true
    transform:
      prefix_chat: ""
```

群组字段说明：
- `name`: 群组名称（用于日志）
- `members`: 群组成员列表（客户端标识）
- `message_types`: 转发的消息类型
- `enabled`: 是否启用此群组
- `transform`: 消息转换配置（可选）

### 传统规则配置

对于复杂的单向转发或特殊需求，仍可使用传统规则：

```yaml
rules:
  - name: "规则名称"
    from_sources: ["来源列表"]
    to_targets: ["目标列表"]
    message_types: ["消息类型"]
    enabled: true
    transform:
      prefix_chat: "[前缀] "
```

### 消息转换

```yaml
transform:
  change_from: "new_source"    # 修改来源标识
  prefix_chat: "[前缀] "       # 聊天消息前缀
  prefix_event: "[事件] "      # 事件消息前缀
```

## 使用场景

### 三个游戏服务器互通

使用群组配置，一次设置即可实现三个服务器的双向互通：

```yaml
groups:
  - name: "三服互通"
    members:
      - "survival_server"
      - "creative_server" 
      - "skyblock_server"
    message_types: ["chat", "event"]
    enabled: true
    transform:
      prefix_chat: "[跨服] "
```

这样配置后：
- `survival_server` 的消息会转发到 `creative_server` 和 `skyblock_server`
- `creative_server` 的消息会转发到 `survival_server` 和 `skyblock_server`
- `skyblock_server` 的消息会转发到 `survival_server` 和 `creative_server`

### 五个平台大互通

```yaml
groups:
  - name: "全平台大群聊"
    members:
      - "minecraft"
      - "qq_bot"
      - "discord_bot"
      - "telegram_bot"
      - "web_chat"
    message_types: ["chat"]
    enabled: true
```

### 传统规则配置对比

如果用传统规则实现三个服务器互通，需要6条规则：

```yaml
rules:
  # survival -> creative, skyblock
  - name: "生存到创造"
    from_sources: ["survival_server"]
    to_targets: ["creative_server"]
    message_types: ["chat"]
    enabled: true
  
  - name: "生存到空岛"
    from_sources: ["survival_server"] 
    to_targets: ["skyblock_server"]
    message_types: ["chat"]
    enabled: true
  
  # 还需要4条规则...
```

而群组配置只需要1条！

### 混合使用

可以同时使用群组和规则配置：

```yaml
# 群组：游戏服务器互通
groups:
  - name: "游戏互通"
    members: ["server1", "server2", "server3"]
    message_types: ["chat"]
    enabled: true

# 规则：特殊的单向转发
rules:
  - name: "监控转发"
    from_sources: ["*"]
    to_targets: ["monitor_system"]
    message_types: ["event"]
    enabled: true
```

## 消息格式

支持标准的WebSocket消息格式：

```json
{
  "from": "消息来源",
  "type": "消息类型",
  "body": {
    "sender": "发送者",
    "chatMessage": "聊天内容",
    "command": "命令",
    "eventDetail": "事件详情"
  },
  "totalId": "消息ID",
  "currentTime": "时间戳"
}
```

## 客户端连接

客户端可以通过WebSocket连接到广播器：

```javascript
const ws = new WebSocket('ws://localhost:8765/ws');

// 发送消息
ws.send(JSON.stringify({
  from: "web_client",
  type: "chat", 
  body: {
    sender: "用户名",
    chatMessage: "Hello World!",
    command: "",
    eventDetail: ""
  },
  totalId: "unique-id",
  currentTime: Date.now().toString()
}));
```

## 日志输出

广播器会输出详细的转发日志：

```
2024/01/01 12:00:00 WebSocket广播器启动: ws://0.0.0.0:8765/ws
2024/01/01 12:00:01 客户端连接: 192.168.1.100:54321
2024/01/01 12:00:02 收到消息 [minecraft -> chat]: Hello World!
2024/01/01 12:00:02 应用规则 [转发聊天消息]: minecraft -> [*]
```

## 高级功能

### 规则优先级

规则按配置文件中的顺序执行，可以通过调整顺序来控制优先级。

### 动态规则

支持通过修改配置文件来动态调整规则（需要重启服务）。

### 客户端标识

客户端可以通过 `X-Client-ID` 请求头来设置自定义标识。

## 故障排除

1. **连接失败**: 检查防火墙和端口占用
2. **消息不转发**: 检查规则配置和客户端标识
3. **性能问题**: 调整规则复杂度和客户端数量
