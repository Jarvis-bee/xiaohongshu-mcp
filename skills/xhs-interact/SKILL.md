---
name: xhs-interact
description: |
  对小红书笔记进行互动：点赞/取消点赞、收藏/取消收藏、发表评论、回复评论。
  当用户想对小红书笔记进行互动时使用——赞一下、收藏一下、留个评论、回复某条评论、取消点赞、取消收藏等。
---

## 多账号约定

- 所有互动工具都支持可选 `account`。
- 用户指定账号或上游搜索/详情结果带有账号上下文时，必须传同一个 `account`。
- 未指定账号时使用 `default`。不要把 A 账号获取的 `feed_id` / `xsec_token` 交给 B 账号操作，除非用户明确要求。

## 输入判断

根据用户意图路由：
- 点赞/取消点赞 → 点赞流程
- 收藏/取消收藏 → 收藏流程
- 发表评论 → 评论流程
- 回复评论 → 回复流程

## 约束

- 评论和回复执行前展示账号别名、笔记信息和内容让用户确认（评论以目标账号身份公开发表，无法撤回）
- 点赞和收藏可直接执行（操作可逆，MCP 服务有幂等处理），但仍要使用正确 `account`
- 所有操作都需要 `feed_id` + `xsec_token`（来自搜索或详情结果，编造会导致报错）
- 操作前先用 `check_login_status` 检查目标 `account` 已登录

## 执行流程

### 点赞

调用 `like_feed`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `feed_id`（string，必填）
- `xsec_token`（string，必填）
- `unlike`（bool，可选）— true 取消点赞，默认 false 点赞

已点赞时再点赞会自动跳过，反之同理。

### 收藏

调用 `favorite_feed`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `feed_id`（string，必填）
- `xsec_token`（string，必填）
- `unfavorite`（bool，可选）— true 取消收藏，默认 false 收藏

已收藏时再收藏会自动跳过，反之同理。

### 发表评论

调用 `post_comment_to_feed`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `feed_id`（string，必填）
- `xsec_token`（string，必填）
- `content`（string，必填）— 评论内容

发送前展示目标 `account` 和评论内容让用户确认。

### 回复评论

调用 `reply_comment_in_feed`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `feed_id`（string，必填）
- `xsec_token`（string，必填）
- `comment_id`（string，可选）— 目标评论 ID
- `user_id`（string，可选）— 目标评论作者 ID
- `parent_comment_id`（string，可选）— 目标为子评论时传父评论 ID，帮助定位
- `content`（string，必填）— 回复内容

`comment_id` 和 `user_id` 至少提供一个。

发送前展示目标 `account` 和回复内容让用户确认。

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 引导使用 xhs-login 登录目标 `account` |
| 缺少 feed_id/xsec_token | 提示先搜索或浏览获取笔记信息，并保留同一 `account` |
| 笔记不可评论 | 告知用户该笔记已关闭评论 |
| 账号参数错误 | 提示账号别名只支持字母、数字、`.`、`_`、`-` |
