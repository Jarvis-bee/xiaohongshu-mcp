---
name: xhs-search
argument-hint: "[搜索关键词]"
description: |
  搜索小红书笔记，支持关键词搜索和多维度筛选（排序、内容类型、时间范围、位置等）。
  当用户想在小红书上搜索、查找内容时使用——包括搜笔记、找攻略、看看小红书上有没有某某内容、搜一下、查一查等场景。
---

## 多账号约定

- `search_feeds` 和登录检查都支持可选 `account`。
- 用户指定账号别名时，先用该 `account` 检查登录，并在搜索调用中传同一个 `account`。
- 未指定账号时使用 `default`。不要在指定账号未登录时自动回退到 `default`。

## 执行流程

### 1. 确认搜索条件

从用户输入中提取：
- `account`（可选）— MCP 账号别名，未指定为 `default`
- `keyword`（必填）— 搜索关键词
- `filters`（可选）— 筛选条件

搜索前先调用 `check_login_status`，携带同一 `account`。

### 2. 调用搜索

调用 `search_feeds`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `keyword`（string，必填）
- `filters`（object，可选）：
  - `sort_by`：综合 | 最新 | 最多点赞 | 最多评论 | 最多收藏
  - `note_type`：不限 | 视频 | 图文
  - `publish_time`：不限 | 一天内 | 一周内 | 半年内
  - `search_scope`：不限 | 已看过 | 未看过 | 已关注
  - `location`：不限 | 同城 | 附近

示例：

```json
{
  "account": "brand-a",
  "keyword": "熬夜恢复 养生",
  "filters": {
    "note_type": "图文",
    "sort_by": "最多收藏",
    "publish_time": "半年内"
  }
}
```

### 3. 展示结果

将搜索结果整理为列表展示，每条包含：
- 标题、作者
- 点赞数、评论数、收藏数
- `feed_id` 和 `xsec_token`（后续操作需要）
- 本次使用的 `account`，提醒后续详情/互动应沿用该账号

提示用户可以：
- 查看某条笔记详情（使用 xhs-explore，并传同一 `account`）
- 对笔记进行互动（使用 xhs-interact，并传同一 `account`）

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 先不要假装已搜索；明确说明目标 `account` 未登录，并引导使用 xhs-login/get_login_qrcode 登录后再搜 |
| 无搜索结果 | 建议调整关键词或筛选条件 |
| 账号参数错误 | 提示账号别名只支持字母、数字、`.`、`_`、`-` |

## Hermes 手动调用 MCP 注意事项

如果当前会话没有直接暴露小红书 MCP 工具，可以通过本机 HTTP MCP 服务调用：`http://127.0.0.1:18060/mcp`。

1. 先 `initialize`，请求头必须包含：
   - `Content-Type: application/json`
   - `Accept: application/json, text/event-stream`
2. 从响应头读取 `Mcp-Session-Id`，后续 `tools/list`、`tools/call` 都要带这个 header。
3. 搜索前先调用 `check_login_status`，`arguments` 带目标 `account`。如果返回 `❌ 未登录`，`search_feeds` 不能用，必须先走登录二维码。
4. `search_feeds` 的调用参数示例：

```json
{
  "name": "search_feeds",
  "arguments": {
    "account": "brand-a",
    "keyword": "熬夜恢复 养生",
    "filters": {
      "note_type": "图文",
      "sort_by": "最多收藏",
      "publish_time": "半年内"
    }
  }
}
```

不要用普通网页搜索冒充 MCP 搜索；用户明确说“用 MCP 搜”时，必须实际调用 MCP，至少完成目标 `account` 的登录状态检查。
