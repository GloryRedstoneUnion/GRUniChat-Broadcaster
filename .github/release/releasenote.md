## 🚀 GRUniChat-Broadcaster v0.3.0 发布

### ✨ 新功能
- 🎯 **executeAt 命令路由**: 支持指定命令在特定服务器执行，实现精确的服务器控制
- 🛡️ **群组黑名单系统**: 添加组级别的消息过滤，支持源/目标服务器、内容关键词过滤
- 🔄 **智能消息路由**: 基于规则和群组的消息路由系统，支持多平台互通
- 📊 **增强配置系统**: 支持复杂的群组配置和黑名单规则
- 🚀 **版本检查功能**: 启动时自动检查最新版本，支持 `--no-check-update` 参数跳过
- 📝 **自动配置生成**: 首次运行时自动创建包含推荐设置的默认配置文件

### 🛠️ 改进
- ⚡ **性能优化**: 添加正则表达式缓存，提高模式匹配性能
- 🔧 **中间件架构**: 重构消息处理流水线，支持认证、验证、日志中间件
- 📋 **配置文件增强**: 支持更灵活的群组配置和黑名单规则设置
- 🌐 **多平台支持**: 新增 FreeBSD 平台支持 (amd64/arm64)
- 📈 **日志改进**: 增强调试模式和错误提示信息

### 🐛 修复
- 🔒 **命令路由安全**: 修复 executeAt 字段验证逻辑，确保目标服务器在线
- 📦 **配置加载**: 修复配置文件不存在时的处理逻辑
- 🔄 **热重载稳定性**: 改进配置热重载功能的稳定性

### 📦 下载

选择适合您系统的版本：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| **Windows** | x64 | `GRUniChat-Broadcaster-*-windows-amd64.exe` |
| **Windows** | ARM64 | `GRUniChat-Broadcaster-*-windows-arm64.exe` |
| **Linux** | x64 | `GRUniChat-Broadcaster-*-linux-amd64` |
| **Linux** | ARM64 | `GRUniChat-Broadcaster-*-linux-arm64` |
| **Linux** | x86 | `GRUniChat-Broadcaster-*-linux-386` |
| **Linux** | ARM | `GRUniChat-Broadcaster-*-linux-arm` |
| **macOS** | Intel | `GRUniChat-Broadcaster-*-darwin-amd64` |
| **macOS** | Apple Silicon | `GRUniChat-Broadcaster-*-darwin-arm64` |
| **FreeBSD** | x64 | `GRUniChat-Broadcaster-*-freebsd-amd64` |
| **FreeBSD** | ARM64 | `GRUniChat-Broadcaster-*-freebsd-arm64` |

### 🔧 使用方法

#### 基础使用
```bash
# 启动广播器（自动生成配置文件）
./GRUniChat-Broadcaster

# 使用自定义配置文件
./GRUniChat-Broadcaster -config custom.yaml

# 启用调试模式
./GRUniChat-Broadcaster -debug

# 跳过版本检查
./GRUniChat-Broadcaster -no-check-update
```

#### 配置示例
```yaml
# 基础配置
server:
  host: "0.0.0.0"
  port: "8765"
  path: "/ws"

# 群组配置
groups:
  - name: "全平台互通"
    members: ["survival", "creative", "qq_bot"]
    message_types: ["chat", "event"]
    enabled: true
    blacklist:
      - name: "过滤管理命令"
        content: ["^/op", "^/gamemode"]
        enabled: true
```

#### executeAt 命令使用
```bash
# 通过WebSocket发送指定服务器命令
{
  "type": "command",
  "body": {
    "content": "give @a diamond 64", 
    "executeAt": "survival"
  }
}
```

### ✨ 主要功能

#### 🎯 executeAt 命令路由
```json
{
  "type": "command",
  "body": {
    "content": "weather clear",
    "executeAt": "survival"
  },
  "source": "admin_console"
}
```
- 指定命令在特定服务器执行
- 防止命令重复执行
- 支持离线服务器检测

#### 🛡️ 群组黑名单系统
```yaml
groups:
  - name: "服务器互通"
    members: ["survival", "creative", "lobby"]
    blacklist:
      - name: "阻止创造到生存"
        from: ["creative"]
        to: ["survival"]
        enabled: true
      - name: "过滤管理命令"
        content: ["^/op", "^/stop", "^/restart"]
        enabled: true
```
- 支持源/目标服务器过滤
- 内容关键词和正则表达式匹配
- 通配符模式支持

#### 🌐 跨平台消息广播
- 游戏服务器间消息转发
- QQ群与游戏服务器互通  
- Discord、Telegram等多平台支持
- 高并发WebSocket连接处理

#### 🔧 智能配置管理
- YAML配置文件，支持热重载
- 自动生成默认配置
- 交互式配置确认
- 配置文件验证和错误提示

### 🆕 与 v0.2.0 相比的主要变化

#### 新增功能
- ✅ **executeAt 命令路由**: 全新的命令路由系统，支持精确控制命令执行位置
- ✅ **群组黑名单**: 组级别的消息过滤系统，支持多种过滤条件
- ✅ **智能路由**: 基于规则和群组的消息路由，替代简单的广播模式
- ✅ **正则表达式缓存**: 提高模式匹配性能
- ✅ **中间件架构**: 可扩展的消息处理流水线

#### 架构改进
- 🔄 **消息处理重构**: 从简单广播升级为智能路由系统
- 🔄 **配置系统增强**: 支持更复杂的群组和规则配置
- 🔄 **错误处理改进**: 更好的错误提示和异常处理

#### 兼容性
- ✅ **向后兼容**: 兼容 v0.2.0 的基础配置格式
- ✅ **平滑升级**: 自动迁移旧配置到新格式
- ✅ **API 稳定**: WebSocket 消息协议保持兼容

---

**🔗 链接**
- [项目主页](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR)
- [问题反馈](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/issues)
- [使用文档](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/blob/main/README.md)

### 🔧 使用方法

#### 基础使用
```bash
# 启动广播器（自动生成配置文件）
./GRUniChat-Broadcaster

# 使用自定义配置文件
./GRUniChat-Broadcaster -config custom.yaml

# 启用调试模式
./GRUniChat-Broadcaster -debug

# 跳过版本检查
./GRUniChat-Broadcaster -no-check-update
```

#### 配置示例
```yaml
# 基础配置
server:
  host: "0.0.0.0"
  port: "8765"
  path: "/ws"

# 群组配置
groups:
  - name: "全平台互通"
    members: ["survival", "creative", "qq_bot"]
    message_types: ["chat", "event"]
    enabled: true
    blacklist:
      - name: "过滤管理命令"
        content: ["^/op", "^/gamemode"]
        enabled: true
```

#### executeAt 命令使用
```bash
# 通过WebSocket发送指定服务器命令
{
  "type": "command",
  "body": {
    "content": "give @a diamond 64", 
    "executeAt": "survival"
  }
}
```

### ✨ 主要功能

#### � executeAt 命令路由
```json
{
  "type": "command",
  "body": {
    "content": "weather clear",
    "executeAt": "survival"
  },
  "source": "admin_console"
}
```
- 指定命令在特定服务器执行
- 防止命令重复执行
- 支持离线服务器检测

#### 🛡️ 群组黑名单系统
```yaml
groups:
  - name: "服务器互通"
    members: ["survival", "creative", "lobby"]
    blacklist:
      - name: "阻止创造到生存"
        from: ["creative"]
        to: ["survival"]
        enabled: true
      - name: "过滤管理命令"
        content: ["^/op", "^/stop", "^/restart"]
        enabled: true
```
- 支持源/目标服务器过滤
- 内容关键词和正则表达式匹配
- 通配符模式支持

#### 🌐 跨平台消息广播
- 游戏服务器间消息转发
- QQ群与游戏服务器互通  
- Discord、Telegram等多平台支持
- 高并发WebSocket连接处理

#### � 智能配置管理
- YAML配置文件，支持热重载
- 自动生成默认配置
- 交互式配置确认
- 配置文件验证和错误提示

---

**🔗 链接**
- [项目主页](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR)
- [问题反馈](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/issues)
- [使用文档](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/blob/main/README.md)
