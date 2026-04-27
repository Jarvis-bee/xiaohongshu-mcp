---
name: xhs-profile
description: |
  查看小红书用户主页：基本信息、粉丝/关注/获赞数据、发布的笔记列表。
  当用户想查看某个博主、作者、用户的主页信息和作品时使用。
---

## 多账号约定

- `my_profile` 和 `user_profile` 都支持可选 `account`。
- 用户指定账号或上游搜索/详情结果带有账号上下文时，主页查询要传同一个 `account`。
- 未指定账号时使用 `default`。

## 执行流程

### 1. 确定查询对象

- 用户想看当前登录账号主页 → 调用 `my_profile`。
- 用户想看某个博主/作者主页 → 调用 `user_profile`。

查询前先用 `check_login_status` 检查目标 `account` 已登录。

### 2. 获取当前账号主页

调用 `my_profile`：
- `account`（string，可选）— 账号别名，留空使用 `default`

### 3. 获取指定用户信息

调用 `user_profile`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `user_id`（string，必填）— 用户 ID（来自笔记详情或搜索结果）
- `xsec_token`（string，必填）

### 4. 展示结果

- 基本信息：昵称、头像、简介、性别、地区
- 数据：粉丝数、关注数、获赞与收藏数
- 最近发布的笔记列表（含 feed_id 和 xsec_token）
- 本次使用的 `account`，提示后续详情/互动沿用该账号

提示用户可以查看某条笔记详情或进行互动。

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 引导使用 xhs-login 登录目标 `account` |
| 用户不存在 | 告知用户该主页无法访问 |
| 账号参数错误 | 提示账号别名只支持字母、数字、`.`、`_`、`-` |
