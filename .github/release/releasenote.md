## 🚀 新版本发布

### ✨ 新功能
- 添加版本检查功能，支持 `--no-check-update` 参数跳过检查
- 新增 FreeBSD 平台支持 (amd64/arm64)
- 改进启动横幅显示
- 添加自动配置生成功能

### 🛠️ 改进
- 移除所有 emoji 字符以提高环境兼容性
- 优化配置文件加载流程
- 改进错误提示信息
- 增强热重载功能的稳定性

### 🐛 修复
- 修复配置文件不存在时的处理逻辑
- 修复版本比较算法的边界情况
- 修复某些环境下的字符编码问题

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

```bash
# 查看帮助
./GRUniChat-Broadcaster -h

# 使用自定义配置文件
./GRUniChat-Broadcaster -config custom.yaml

# 启用调试模式
./GRUniChat-Broadcaster -debug

# 跳过版本检查
./GRUniChat-Broadcaster -no-check-update

# 禁用热重载
./GRUniChat-Broadcaster -hot-reload=false

# 禁用交互式确认
./GRUniChat-Broadcaster -interactive=false
```

### ✨ 主要功能

- 🌐 **跨平台 WebSocket 消息广播** - 支持多种操作系统和架构
- 🔄 **配置文件热重载** - 无需重启即可更新配置
- 📊 **实时连接统计** - 提供详细的连接状态和消息状态查询
- 🛡️ **自动配置生成** - 首次运行时自动创建默认配置文件
- 🔍 **版本更新检查** - 自动检查并提示最新版本
- 🎯 **API 端点** - 提供统计信息和消息状态查询接口

---

**🔗 链接**
- [项目主页](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR)
- [问题反馈](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/issues)
- [使用文档](https://github.com/GloryRedstoneUnion/GRUniChat-MCDR/blob/main/README.md)
