---
name: xhs-search
argument-hint: "[搜索关键词]"
description: |
  搜索小红书笔记，支持关键词搜索和多维度筛选（排序、内容类型、时间范围、位置等）。
  当用户想在小红书上搜索、查找内容时使用——包括搜笔记、找攻略、看看小红书上有没有某某内容、搜一下、查一查等场景。
---

## 执行流程

### 1. 确认搜索条件

从用户输入中提取：
- `keyword`（必填）— 搜索关键词
- `filters`（可选）— 筛选条件

### 2. 调用搜索

调用 `search_feeds`：
- `keyword`（string，必填）
- `filters`（object，可选）：
  - `sort_by`：综合 | 最新 | 最多点赞 | 最多评论 | 最多收藏
  - `note_type`：不限 | 视频 | 图文
  - `publish_time`：不限 | 一天内 | 一周内 | 半年内
  - `search_scope`：不限 | 已看过 | 未看过 | 已关注
  - `location`：不限 | 同城 | 附近

### 3. 展示结果

将搜索结果整理为列表展示，每条包含：
- 标题、作者
- 点赞数、评论数、收藏数
- `feed_id` 和 `xsec_token`（后续操作需要）

提示用户可以：
- 查看某条笔记详情（使用 xhs-explore）
- 对笔记进行互动（使用 xhs-interact）

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 先不要假装已搜索；明确说明 MCP 当前未登录，并引导使用 xhs-login/get_login_qrcode 登录后再搜 |
| 无搜索结果 | 建议调整关键词或筛选条件 |

## Hermes 手动调用 MCP 注意事项

如果当前会话没有直接暴露小红书 MCP 工具，可以通过本机 HTTP MCP 服务调用：`http://127.0.0.1:18060/mcp`。

1. 先 `initialize`，请求头必须包含：
   - `Content-Type: application/json`
   - `Accept: application/json, text/event-stream`
2. 从响应头读取 `Mcp-Session-Id`，后续 `tools/list`、`tools/call` 都要带这个 header。
3. 搜索前先调用 `check_login_status`。如果返回 `❌ 未登录`，`search_feeds` 不能用，必须先走登录二维码。
4. `search_feeds` 的调用参数示例：

```json
{
  "name": "search_feeds",
  "arguments": {
    "keyword": "熬夜恢复 养生",
    "filters": {
      "note_type": "图文",
      "sort_by": "最多收藏",
      "publish_time": "半年内"
    }
  }
}
```

不要用普通网页搜索冒充 MCP 搜索；用户明确说“用 MCP 搜”时，必须实际调用 MCP，至少完成登录状态检查。
