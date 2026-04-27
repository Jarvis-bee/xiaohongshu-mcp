---
name: xhs-explore
description: |
  浏览小红书推荐流、查看笔记详情和评论。
  当用户想看推荐内容、刷首页、查看某条笔记的详情/评论、或已有 feed_id 想获取完整内容时使用。
---

## 多账号约定

- `list_feeds`、`get_feed_detail` 都支持可选 `account`。
- 用户指定账号别名时，所有浏览和详情调用都传同一个 `account`；未指定时使用 `default`。
- 详情、互动、作者主页应尽量沿用获取 `feed_id` / `xsec_token` 时使用的同一账号。

## 输入判断

- 用户想浏览推荐 → 步骤 1
- 用户提供了 feed_id → 步骤 2
- 用户指定账号或上下文已有账号 → 记录为本次 `account`

## 执行流程

### 1. 获取推荐流

先调用 `check_login_status` 携带目标 `account`。已登录后调用 `list_feeds`：

```json
{
  "account": "brand-a"
}
```

未指定账号时可省略 `account` 使用 `default`。

展示每条笔记的标题、作者、互动数据，附带 `feed_id`、`xsec_token` 和本次使用的 `account`。

### 2. 查看笔记详情

调用 `get_feed_detail`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `feed_id`（string，必填）
- `xsec_token`（string，必填）
- `load_all_comments`（bool，可选，默认 false，仅返回前 10 条评论）
- `limit`（int，可选，load_all_comments=true 时生效，默认 20）
- `click_more_replies`（bool，可选，是否展开二级回复）
- `reply_limit`（int，可选，跳过回复数超过此值的评论，默认 10）
- `scroll_speed`（string，可选：slow | normal | fast）

展示：笔记内容、图片、作者信息、互动数据、评论列表。

提示用户可以：
- 点赞/收藏（使用 xhs-interact，并传同一 `account`）
- 发表评论（使用 xhs-interact，并传同一 `account`）
- 查看作者主页（使用 xhs-profile，并传同一 `account`）

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 引导使用 xhs-login 登录目标 `account` |
| 笔记已删除或不可见 | 告知用户该笔记无法访问 |
| 账号参数错误 | 提示账号别名只支持字母、数字、`.`、`_`、`-` |
