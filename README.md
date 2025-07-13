# WebSocket 广播器

一个用Go实现的WebSocket消息广播中继服务器，支持灵活的规则配置、消息转换和智能命令路由。

## ✨ 核心特性

- 🔄 **智能转发**: 基于规则和群组的消息路由系统
- 🎯 **精确命令路由**: 支持 `executeAt` 字段的单服务器命令执行
- 🛡️ **群组黑名单**: 组级别的消息过滤和内容屏蔽
- 🔧 **消息转换**: 支持添加前缀、修改来源等消息转换
- ⚡ **高性能**: Go语言实现，支持高并发连接
- 📝 **配置驱动**: YAML配置文件，支持热重载
- 🔗 **多平台**: 支持游戏服务器、QQ、Discord等多平台互通

## 🚀 最新功能

### executeAt 命令路由
支持指定命令在特定服务器执行，实现精确的服务器控制：

```json
{
  "from": "admin_panel",
  "type": "command",
  "body": {
    "sender": "Admin",
    "command": "give @a diamond 64",
    "executeAt": "survival"
  }
}
```

### 群组黑名单
支持组级别的消息过滤，防止不当内容传播：

```yaml
groups:
  - name: "全服互通"
    members: ["survival", "creative", "lobby"]
    blacklist:
      - name: "防止创造到生存"
        from: ["creative"]
        to: ["survival"]
        enabled: true
      - name: "过滤危险命令"
        content: ["^/stop", "^/restart"]
        enabled: true
```

## 项目结构

```
GRUniChat-Broadcaster/
├── main.go                    # 主程序入口
├── go.mod                     # Go模块定义
├── go.sum                     # 依赖校验文件
├── config.yaml                # 默认配置文件
├── configs/                   # 配置文件目录
│   ├── config.yaml           # 默认配置文件
│   └── config.example.yaml   # 配置示例文件
├── internal/                  # 内部包（私有）
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── message/              # 消息处理
│   │   └── message.go
│   └── connection/           # 连接管理
│       └── connection.go
├── pkg/                      # 公共包
│   ├── broadcaster/          # 广播器核心
│   ├── database/             # 数据库支持
│   ├── logger/               # 日志系统
│   ├── middleware/           # 中间件
│   ├── redis/                # Redis支持
│   ├── router/               # 路由器
│   └── utils/                # 工具函数
├── test_command_format.py     # Python测试脚本
├── EXECUTE_AT_GUIDE.md       # executeAt字段使用指南
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

#### 🎯 pkg/broadcaster/
消息广播核心：
- 广播器实现和连接管理
- executeAt命令路由逻辑
- 群组黑名单过滤系统

#### � pkg/router/
消息路由系统：
- 规则匹配和目标计算
- 群组路由逻辑
- 路由信息管理

#### 🔧 pkg/middleware/
中间件系统：
- 认证、验证、日志中间件
- 中间件链管理
- 消息处理流水线

#### 💾 pkg/database/
数据存储支持：
- 内存、Redis、MySQL、PostgreSQL支持
- 消息持久化和查询
- 连接池管理

#### 📝 pkg/logger/
日志系统：
- 结构化日志输出
- 多级别日志支持
- 调试模式

#### 🛠️ pkg/utils/
通用工具函数：
- 模式匹配函数
- 字符串数组操作
- 消息类型验证

## 功能特性

- 🔄 **智能转发**: 基于规则的消息路由和转发
- 🎯 **精确匹配**: 支持来源、目标、消息类型的多维度过滤
- 🔧 **消息转换**: 支持添加前缀、修改来源等消息转换
- ⚡ **高性能**: Go语言实现，支持高并发连接
- 📝 **配置驱动**: YAML配置文件，支持热重载
- 🔗 **多平台**: 支持游戏服务器、QQ、Discord等多平台互通

## 🚀 快速开始

### 环境要求
- Go 1.19+
- 支持跨平台（Windows、Linux、macOS）

### 1. 编译项目

```bash
cd GRUniChat-Broadcaster
go mod tidy
go build -o broadcaster main.go
```

### 2. 配置文件

编辑 `config.yaml` 文件：

```yaml
debug: true
host: "localhost"
port: 9001
database:
  type: "memory"  # memory/redis/mysql/postgresql
auth:
  enabled: true
  token: "your-secure-token"
group_config:
  blacklist:
    default:
      - "forbidden_group_1"
      - "forbidden_group_2"
```

### 3. 启动服务器

```bash
# 使用默认配置文件
./broadcaster

# Windows下
.\broadcaster.exe
```

## 🔧 配置说明

### 基础配置

```yaml
debug: true           # 调试模式
host: "localhost"     # 服务器地址
port: 9001           # 监听端口

# 数据库配置
database:
  type: "memory"      # 数据库类型: memory/redis/mysql/postgresql
  
# 认证配置
auth:
  enabled: true       # 启用认证
  token: "your-token" # 认证令牌

# 群组黑名单配置
group_config:
  blacklist:
    default:          # 默认黑名单
      - "forbidden_group_1"
      - "forbidden_group_2"
```

### executeAt 命令路由

支持通过 `executeAt` 字段指定命令执行的目标服务器：

```json
{
  "type": "command",
  "body": {
    "content": "list",
    "executeAt": "survival"
  },
  "source": "web_admin",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

详细使用方法请参考 [EXECUTE_AT_GUIDE.md](./EXECUTE_AT_GUIDE.md)

```yaml
blacklist:
  - name: "规则名称"
    description: "规则描述"
    from: ["source1", "source2*"]     # 来源过滤（支持通配符）
    to: ["target1", "target2*"]       # 目标过滤（支持通配符）
### 群组黑名单规则

可以配置群组级别的黑名单规则来过滤特定消息：

```yaml
group_config:
  blacklist:
    survival_server:      # 服务器特定黑名单
      - "banned_group_1"
      - "banned_group_2"
    creative_server:
      - "test_group"
    default:             # 默认黑名单（所有服务器）
      - "global_banned"
```

## 📋 使用场景

### 多服务器互通
- 游戏服务器间消息转发
- QQ群与游戏服务器互通
- Discord与Minecraft联通
- 多平台统一管理

### 命令路由管理
- 通过 `executeAt` 字段指定命令执行目标
- 支持单服务器精确命令投递
- 避免命令重复执行问题
```

## 📨 消息格式

### 标准消息结构

支持标准的WebSocket消息格式：

```json
{
  "from": "消息来源",
  "type": "消息类型",
  "body": {
    "sender": "发送者",
    "chatMessage": "聊天内容",
    "command": "命令",
    "executeAt": "目标服务器",
    "eventDetail": "事件详情"
  },
  "totalId": "消息ID",
  "currentTime": "时间戳"
}
```

  "type": "chat|command|event",
  "body": {
    "content": "消息内容",
    "executeAt": "目标服务器ID"  // 仅command类型使用
  },
  "source": "消息来源",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 消息类型说明

- **chat**: 聊天消息
- **command**: 命令消息，支持 `executeAt` 字段指定执行目标
- **event**: 事件消息

### executeAt 字段

当消息类型为 `command` 时，可以使用 `executeAt` 字段指定命令执行的目标服务器：

```json
{
  "type": "command",
  "body": {
    "content": "list",
    "executeAt": "survival"
  },
  "source": "admin_console",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**行为说明**：
- 如果指定了 `executeAt` 字段，命令只会发送到指定的服务器
- 如果目标服务器离线，会返回错误信息
- 如果未指定 `executeAt`，命令会按正常路由规则广播

## 🔌 客户端连接

客户端可以通过WebSocket连接到广播器：

```javascript
const ws = new WebSocket('ws://localhost:8765/ws');

// 发送聊天消息
ws.send(JSON.stringify({
  type: "chat",
  body: {
    content: "Hello World!"
  },
  source: "web_client",
  timestamp: new Date().toISOString()
}));

// 发送命令（指定服务器）
ws.send(JSON.stringify({
  type: "command", 
  body: {
    content: "weather clear",
    executeAt: "survival"
  },
  source: "admin_console",
  timestamp: new Date().toISOString()
}));
```

## 📝 日志系统

广播器提供详细的日志输出，支持调试模式：

```
[INFO] 2024/01/15 12:00:00 WebSocket服务器启动: ws://localhost:9001/ws
[INFO] 2024/01/15 12:00:01 客户端连接: 192.168.1.100:54321 (ID: survival)
[INFO] 2024/01/15 12:00:02 收到消息 [survival -> chat]: Hello World!
[INFO] 2024/01/15 12:00:02 消息已广播到 2 个客户端
[WARN] 2024/01/15 12:00:03 executeAt目标服务器离线: creative
```

**过滤策略**：
- **来源过滤**: 指定来源服务器的消息
- **目标过滤**: 阻止发送到特定目标
- **内容过滤**: 关键词和正则表达式过滤
- **组合过滤**: 多条件同时满足才触发

### 热重载配置

广播器支持配置文件热重载：

```bash
# 修改配置文件后，程序会自动检测并重新加载
# 支持交互式确认，确保配置正确
```

### 消息路由优先级

1. **executeAt 优先**: 命令消息的 `executeAt` 字段具有最高优先级
2. **黑名单过滤**: 在路由规则之后应用黑名单
3. **群组路由**: 按群组配置进行消息分发
## 🔧 高级功能

### executeAt 命令路由

使用 `executeAt` 字段可以精确控制命令在哪个服务器执行：

```json
{
  "type": "command",
  "body": {
    "content": "give @a diamond 10",
    "executeAt": "survival"
  },
  "source": "admin_console",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**应用场景**：
- 管理面板控制特定服务器  
- 上游模块解析命令格式并路由
- 自动化脚本的精确操作

### 群组黑名单系统

配置群组级别的黑名单过滤：

```yaml
group_config:
  blacklist:
    survival:           # 生存服务器黑名单
      - "test_group"
      - "banned_group"
    creative:           # 创造服务器黑名单  
      - "creative_only"
    default:            # 全局黑名单
      - "global_banned"
```

### 热重载配置

支持运行时重新加载配置文件：
- 自动检测配置文件变化
- 交互式确认配置更新
- 保持现有连接不断开

## 🛠️ 故障排除

### 常见问题

1. **连接失败**: 
   - 检查防火墙和端口占用
   - 确认服务器地址和端口配置

2. **消息不转发**: 
   - 检查客户端连接状态
   - 查看日志确认消息路由

3. **executeAt 命令失败**:
   - 确认目标服务器在线且已连接
   - 检查服务器名称拼写

4. **黑名单不生效**:
   - 检查黑名单配置语法
   - 确认群组名称匹配

### 调试技巧

```bash
# 启用调试模式
./broadcaster

# 配置文件中启用调试
debug: true
```

## 📖 相关文档

- [executeAt 字段详细说明](./EXECUTE_AT_GUIDE.md)
- [WebSocket协议文档](./WEBSOCKET_PROTOCOL.md)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进项目！

## 📄 许可证

MIT License
