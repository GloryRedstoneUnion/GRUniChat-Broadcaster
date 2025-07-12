# 配置文件说明

## 群组配置 vs 传统规则配置

WebSocket 广播器支持两种配置方式：**群组配置**（推荐）和**传统规则配置**。

### 群组配置（Groups）- 推荐使用

群组配置是一种简化的配置方式，特别适合多平台互通场景。

**优势：**
- 配置简洁：一个群组配置即可实现多个成员间的双向互通
- 易于理解：直观地定义哪些平台需要互相通信
- 减少冗余：无需为每个方向单独配置规则

**示例：**
```yaml
groups:
  - name: "全平台互通"
    members:
      - "minecraft"
      - "qq_bot" 
      - "discord_bot"
    message_types: ["chat"]
    enabled: true
```

这一个群组配置相当于以下6条传统规则：
- minecraft → qq_bot
- minecraft → discord_bot
- qq_bot → minecraft
- qq_bot → discord_bot
- discord_bot → minecraft
- discord_bot → qq_bot

### 传统规则配置（Rules）

传统规则配置提供更精细的控制，适合复杂的单向转发需求。

**使用场景：**
- 单向转发（如监控系统）
- 复杂的消息过滤和转换
- 特殊的路由需求

**示例：**
```yaml
rules:
  - name: "监控转发"
    from_sources: ["*"]
    to_targets: ["monitor_system"]
    message_types: ["event"]
    enabled: true
```

### 配置优先级

路由器按以下顺序处理配置：

1. **群组配置优先**：首先检查发送方是否属于任何群组
2. **传统规则兜底**：如果没有匹配的群组，则使用传统规则

### 混合使用

可以同时使用群组配置和传统规则配置：

```yaml
# 群组：常规的多平台互通
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

### 选择建议

- **简单场景**：使用群组配置
- **复杂场景**：群组配置 + 传统规则配置
- **特殊需求**：传统规则配置

## 配置文件列表

| 文件名 | 说明 |
|--------|------|
| `config.yaml` | 默认配置文件（群组配置为主） |
| `config.groups.yaml` | 纯群组配置示例 |
| `config.example.yaml` | 详细配置示例 |
| `config.simple.yaml` | 简化配置示例 |
