# GRUniChat-Broadcaster

一个用Go实现的高性能WebSocket消息广播中继服务器，专为多平台实时消息同步而设计。

## ✨ 核心特性

- 🔄 **智能消息路由**: 支持基于群组规则的消息转发和过滤
- 🎯 **精确命令路由**: 支持 `executeAt` 字段的定向命令执行
- 🛡️ **群组黑名单**: 灵活的消息过滤和内容屏蔽机制
- 🔧 **消息转换**: 兼容多种消息格式（新版本和旧版本）
- ⚡ **高性能**: Go语言实现，支持高并发WebSocket连接
- � **配置热重载**: 运行时动态重载配置，无需重启服务
- � **结构化日志**: 详细的调试日志和运行状态监控
- 🎨 **优雅启动/关闭**: 完整的服务生命周期管理

## 🚀 最新功能

### 标准消息格式（新版本）
支持简洁的标准消息格式：

```json
{
  "type": "chat|command|event",
  "body": {
    "content": "消息内容",
    "executeAt": "目标服务器ID"  // 仅command类型使用
  },
  "source": "消息来源",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### executeAt 命令路由
支持指定命令在特定服务器执行，实现精确的服务器控制：

```json
{
  "type": "command",
  "body": {
    "content": "weather clear",
    "executeAt": "survival"
  },
  "source": "admin_console",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 配置热重载
运行时动态重载配置文件，支持交互式确认：
- 自动检测配置文件变化
- 交互式确认更新
- 保持现有连接不断开

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
- 打印启动横幅和版本信息
- 命令行参数解析（config, debug, hot-reload, interactive, no-check-update）
- 版本更新检查（GitHub API）
- 配置文件加载和验证
- 热重载管理器初始化
- WebSocket服务器启动
- 优雅关闭处理

### � internal/ 内部包

#### �📨 internal/message/
消息相关的结构和方法：
- `Message` 和 `Body` 结构体定义
- 消息内容获取、克隆、验证等方法
- 时间戳更新功能

#### ⚙️ internal/config/
配置文件处理：
- 配置结构体定义（服务器、群组、黑名单等）
- YAML配置文件加载和验证
- 配置热重载管理器
- 交互式配置更新确认

#### 🔗 internal/connection/
WebSocket连接管理：
- 连接管理器实现
- WebSocket升级和处理
- 消息广播和路由逻辑
- executeAt 命令路由
- 群组黑名单过滤
- 连接统计和状态管理

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
go build -o broadcaster.exe .

# Linux/macOS
go build -o broadcaster .
```

### 2. 配置文件

编辑 `config.yaml` 文件：

```yaml
server:
  host: "localhost"
  port: 8765
  path: "/ws"

groups:
  - name: "游戏服务器互通"
    members: ["survival", "creative", "lobby"]
    blacklist:
      - name: "防止测试消息"
        from: ["test_*"]
        to: ["survival"]
        enabled: true
```

### 3. 启动服务器

```bash
# 基本启动
./broadcaster

# 指定配置文件
./broadcaster -config custom.yaml

# 启用调试模式
./broadcaster -debug

# 禁用热重载
./broadcaster -hot-reload=false

# 禁用版本检查
./broadcaster -no-check-update

# Windows下
.\broadcaster.exe
```

### 4. 命令行参数

- `-config`: 配置文件路径（默认: config.yaml）
- `-debug`: 启用调试模式（默认: false）
- `-hot-reload`: 启用配置热重载（默认: true）
- `-interactive`: 启用交互式热重载确认（默认: true）
- `-no-check-update`: 跳过版本检查（默认: false）

## 🔧 配置说明

### 基础配置

```yaml
server:
  host: "localhost"      # 服务器监听地址
  port: 8765            # 监听端口
  path: "/ws"           # WebSocket路径

# 群组配置
groups:
  - name: "服务器互通"
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

### 消息格式支持

#### 新版本格式（推荐）
```json
{
  "type": "chat|command|event",
  "body": {
    "content": "消息内容",
    "executeAt": "目标服务器ID"  // 可选，仅command类型
  },
  "source": "消息来源",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### 旧版本格式（兼容）
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
  "currentTime": "时间戳(毫秒)"
}
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

### 群组黑名单规则

配置群组级别的消息过滤：

```yaml
groups:
  - name: "全服互通"
    members: ["survival", "creative", "lobby", "qq_bot"]
    blacklist:
      - name: "阻止创造到生存"
        from: ["creative"]
        to: ["survival"]
        enabled: true
      - name: "过滤管理命令"
        content: ["^/op", "^/stop", "^/restart"]
        enabled: true
      - name: "禁止测试群组"
        from: ["test_*"]
        to: ["*"]
        enabled: true
```

**规则字段说明**：
- `name`: 规则名称（描述性）
- `from`: 来源过滤（支持通配符 `*`）
- `to`: 目标过滤（支持通配符 `*`）
- `content`: 内容过滤（支持正则表达式）
- `enabled`: 是否启用此规则

## 📋 使用场景

### 多平台消息互通
- 游戏服务器间实时聊天转发
- QQ群与Minecraft服务器联通
- Discord与游戏平台集成
- 跨平台管理和监控

### 精确命令管理
- 通过 `executeAt` 字段指定命令执行目标
- 避免命令在所有服务器重复执行
- 支持远程单服务器操作
- 管理面板精确控制
```

## 📨 消息格式详解

### 标准消息结构（新版本）

```json
{
  "type": "chat|command|event",
  "body": {
    "content": "消息内容",
    "executeAt": "目标服务器ID"  // 仅command类型使用
  },
  "source": "消息来源",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 兼容消息结构（旧版本）

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
  "currentTime": "时间戳(毫秒)"
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

## 🔌 客户端连接示例

客户端可以通过WebSocket连接到广播器：

```javascript
const ws = new WebSocket('ws://localhost:8765/ws');

ws.onopen = function() {
    console.log('已连接到GRUniChat-Broadcaster');
};

// 发送聊天消息
ws.send(JSON.stringify({
    type: "chat",
    body: {
        content: "[玩家] Hello World!"
    },
    source: "survival",
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

// 发送事件消息
ws.send(JSON.stringify({
    type: "event",
    body: {
        content: "服务器已启动"
    },
    source: "survival",
    timestamp: new Date().toISOString()
}));
```

## 📝 日志系统

广播器提供详细的结构化日志输出：

```
>>> 服务器启动成功: ws://localhost:8765/ws
>>> 配置热重载: 已启用
>>> 调试模式: 已启用
>>> 按 Ctrl+C 停止服务器

[INFO] 2024/01/15 12:00:01 新客户端连接: survival (192.168.1.100)
[INFO] 2024/01/15 12:00:02 收到消息 [survival] chat: Hello World!
[INFO] 2024/01/15 12:00:02 消息已广播到 2 个客户端
[WARN] 2024/01/15 12:00:03 executeAt目标服务器离线: creative
[INFO] 2024/01/15 12:00:04 配置文件已更新，正在重新加载...
```

### 日志级别
- `INFO`: 一般信息（连接、消息传递）
- `WARN`: 警告信息（服务器离线、过滤消息）
- `ERROR`: 错误信息（连接失败、配置错误）
- `DEBUG`: 调试信息（详细的消息内容和处理过程）

## 🔧 高级功能

### 自动版本检查
启动时自动检查GitHub最新版本：
```
>>> 检查更新中... 发现新版本!
>>> 当前版本: v0.2.0
>>> 最新版本: v0.3.0
>>> 下载地址: https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/releases/latest
```

### executeAt 命令路由

精确控制命令在哪个服务器执行：

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

**特性**：
- 如果指定了 `executeAt`，命令只会发送到指定服务器
- 如果目标服务器离线，会返回错误响应
- 如果未指定 `executeAt`，命令按群组规则广播

### 配置热重载

支持运行时重新加载配置：
- 自动检测配置文件变化
- 交互式确认配置更新
- 保持现有连接不断开
- 实时应用新的路由规则

### 优雅启动关闭

完整的服务生命周期管理：
- 启动横幅显示版本信息
- 信号处理（Ctrl+C）
- 30秒超时的优雅关闭
- 清理所有资源和连接

## 🛠️ 故障排除

### 常见问题

1. **连接失败**: 
   - 检查防火墙和端口占用（默认8765）
   - 确认服务器地址和端口配置
   - 查看服务器启动日志

2. **消息不转发**: 
   - 检查客户端连接状态（查看日志）
   - 确认群组配置和成员设置
   - 检查黑名单规则是否阻止了消息

3. **executeAt 命令失败**:
   - 确认目标服务器在线且已连接
   - 检查服务器名称拼写（区分大小写）
   - 查看错误响应消息

4. **配置热重载不生效**:
   - 确认启用了热重载功能（-hot-reload=true）
   - 检查配置文件语法是否正确
   - 查看交互式确认提示

5. **版本检查失败**:
   - 检查网络连接
   - 使用 -no-check-update 跳过版本检查

### 调试技巧

```bash
# 启用调试模式查看详细日志
./broadcaster -debug

# 禁用版本检查加快启动
./broadcaster -no-check-update

# 禁用热重载（生产环境）
./broadcaster -hot-reload=false -interactive=false
```

### 配置验证

程序启动时会自动验证配置文件：
- 检查YAML语法
- 验证群组成员设置
- 检查黑名单规则格式

## 📖 相关文档

- [WebSocket协议文档](./WEBSOCKET_PROTOCOL.md) - 详细的消息格式和协议说明
- [配置文件示例](./configs/config.example.yaml) - 完整的配置示例

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进项目！

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 许可证

本项目基于 MIT License 开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🌟 致谢

- Glory Redstone Union 团队
- 所有贡献者和用户的支持
