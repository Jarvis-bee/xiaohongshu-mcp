---
name: post-to-xhs
argument-hint: "[标题或内容描述]"
description: |
  发布内容到小红书，支持图文笔记和视频笔记。自动判断发布类型，校验标题和素材，用户确认后发布。
  当用户想在小红书发布内容时使用——包括发笔记、发图文、发视频、上传图片、写一篇小红书、把内容发到红书上、种草笔记、好物分享等，即使用户只说"帮我发一下"但上下文明确是小红书也应触发。
---

## 多账号约定

- `publish_content`、`publish_with_video`、`check_login_status` 都支持可选 `account`。
- 用户指定账号别名时，发布前登录检查和最终发布必须传同一个 `account`；未指定时使用 `default`。
- `account` 只支持字母、数字、`.`、`_`、`-`，不要把小红书昵称当成别名，除非用户明确这样配置。
- 发布预览中必须展示目标 `account`，避免发到错误账号。

## 输入判断

根据用户提供的素材判断发布类型：
- 提供了视频文件 → 视频笔记
- 提供了图片 → 图文笔记
- 仅提供文本 → 提示用户至少提供图片或视频

## 约束

- 标题最多 20 个中文字或英文单词（小红书平台限制，超长会被截断）
- 图文笔记至少 1 张图片（小红书不允许纯文本笔记）
- 视频笔记仅支持本地视频文件绝对路径（MCP 服务需要读取本地文件）
- 图片和视频不能混用，只能二选一（小红书平台限制）
- 正文中不要包含 # 标签（标签通过 `tags` 参数单独传递，MCP 服务会自动处理格式）
- 发布前展示完整内容让用户确认（发布后无法撤回）

## 执行流程

### 1. 收集发布信息

确保以下内容齐全：
- `account`（可选）— MCP 账号别名，未指定为 `default`
- `title`（必填）— 标题
- `content`（必填）— 正文
- 图片列表或视频路径（必填其一）
- `tags`（可选）— 话题标签
- `schedule_at`（可选）— 定时发布，ISO8601 格式
- `is_original`（可选，仅图文）— 声明原创
- `visibility`（可选）— 公开可见 | 仅自己可见 | 仅互关好友可见

信息不完整时，向用户询问缺少的部分。发布前先调用 `check_login_status` 检查目标 `account` 已登录，未登录则引导使用 xhs-login 登录该账号。

### 2. 内容校验

- 检查 `account` 是否只包含字母、数字、`.`、`_`、`-`
- 检查标题长度（≤20 中文字）
- 检查图片/视频文件路径是否为绝对路径
- 如用户提供 URL 内容，先用 WebFetch 提取文本和图片

### 3. 确认发布

向用户展示完整的发布内容预览：
- 目标账号别名 `account`
- 标题、正文、标签
- 图片列表或视频路径
- 定时时间、可见范围（如有）

等待用户确认后才执行发布。

### 4. 发布

**图文笔记** — 调用 `publish_content`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `title`（string，必填）
- `content`（string，必填）
- `images`（string[]，必填）— 图片路径或 URL
- `tags`（string[]，可选）
- `schedule_at`（string，可选）
- `is_original`（bool，可选）
- `visibility`（string，可选）

**视频笔记** — 调用 `publish_with_video`：
- `account`（string，可选）— 账号别名，留空使用 `default`
- `title`（string，必填）
- `content`（string，必填）
- `video`（string，必填）— 本地视频绝对路径
- `tags`（string[]，可选）
- `schedule_at`（string，可选）
- `visibility`（string，可选）

### 5. 报告结果

发布成功后，告知用户目标 `account`、笔记 ID 和发布状态。

## Docker MCP 发布图片路径坑

当 `xiaohongshu-mcp` 通过 Docker Compose 运行时，容器默认只挂载：

- `./data:/app/data`（多账号 cookies 通常落到 `/app/data/accounts/<account>/cookies.json` 或 compose 配置的 `COOKIES_DIR/<account>/cookies.json`）
- `./images:/app/images`

所以发布图文时，传给 `publish_content.images` 的本机绝对路径如果位于 `/home/bee/.hermes/tmp/...`、`/tmp/...` 等目录，容器内会看不到，日志会出现：

```text
图片文件不存在: /home/bee/.hermes/tmp/...
```

正确做法：

1. 先把要发布的图片复制到宿主机仓库的 Docker 挂载目录，例如：
   `/home/bee/apps/xiaohongshu-mcp/docker/images/<job-name>/xxx.png`
2. 调用 `publish_content` 时使用容器内路径：
   `/app/images/<job-name>/xxx.png`
3. 如果发布接口长时间无响应，先看 `docker logs --tail 120 xiaohongshu-mcp`，不要只看客户端超时。

### 用户要求不用容器时：本机源码直启

如果用户明确说“不用容器启动，直接启动”，或者发布图片主要位于宿主机绝对路径（例如 `/home/bee/.hermes/tmp/...`）且不想复制到 `/app/images`，可以改用本机源码直启。这样 MCP 服务读取的是宿主机文件系统，`publish_content.images` 可以直接传本机绝对路径。

实操步骤：

```bash
cd /home/bee/apps/xiaohongshu-mcp
# 先确认本机 Chrome 可用
which google-chrome || which chromium || which chromium-browser
# 若 18060 已被 Docker 占用，先停掉容器
cd /home/bee/apps/xiaohongshu-mcp/docker && docker compose stop
# 回仓库根目录直启；非无头模式更利于登录/发布排障
cd /home/bee/apps/xiaohongshu-mcp
ROD_BROWSER_BIN=/usr/bin/google-chrome COOKIES_DIR=/home/bee/apps/xiaohongshu-mcp/docker/data/accounts go run . -headless=false
```

验证：

```bash
curl -sS -i http://localhost:18060/mcp | head -n 20
# 返回 405 Method Not Allowed / GET requires an active session 说明 MCP HTTP 服务已运行
```

注意：
- 本机直启默认 cookies 目录是 `~/.xiaohongshu-mcp/accounts`；如果想沿用 Docker 登录态，显式设置 `COOKIES_DIR=/home/bee/apps/xiaohongshu-mcp/docker/data/accounts`。
- `go run . -headless=false` 是长驻服务，应用 background process 跑，不要前台阻塞后续操作。
- 若 `curl` 可达但日志为空，不代表没启动；Gin/MCP 对 GET `/mcp` 返回 405 是正常健康信号。

## 上游修复与容器更新验证

如果遇到 `发布失败: 没有找到发布 TAB - 上传图文`：

1. 确认仓库是否是 `Jarvis-bee/xiaohongshu-mcp` fork，upstream 应为 `https://github.com/xpzouying/xiaohongshu-mcp.git`。
2. 拉取并合并 upstream/main，确认包含提交：
   `c63748f fix: 修复发布图文时找不到上传TAB的问题 (#666)`。
3. Docker Compose 使用的是镜像 `xpzouying/xiaohongshu-mcp`，合并本地代码不会自动影响容器；需要执行：

```bash
cd /home/bee/apps/xiaohongshu-mcp/docker
docker compose pull
docker compose up -d --force-recreate
```

4. 用 `docker inspect xiaohongshu-mcp --format '{{.Image}} {{.Created}}'` 验证容器镜像/创建时间已变化。
5. 强制重建容器后登录态可能失效，先对目标 `account` 调用 `check_login_status`，必要时重新获取二维码登录。

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 引导使用 xhs-login 登录目标 `account` |
| 标题超长 | 提示用户缩短标题 |
| 图片路径无效 | Docker 环境优先检查图片是否在 `/app/images` 挂载路径内 |
| 视频使用了相对路径 | 提示改为绝对路径，并确认容器可见 |
| 发布失败 | 先查看 `docker logs --tail 120 xiaohongshu-mcp`，确认日志里的 account/cookies 路径和具体错误 |

### Hermes 当前会话 HTTP fallback

当 Hermes 当前工具列表没有直接暴露 `publish_content`，但 `xiaohongshu-mcp` 已在 `http://localhost:18060/mcp` 运行时，可以手动按 MCP HTTP 协议调用：先 `initialize` 取响应头 `Mcp-Session-Id`，带 session 发 `notifications/initialized`，再 `tools/call`。

关键坑：请求头必须包含 `Accept: application/json, text/event-stream`，否则 initialize 会 400：`Accept must contain both 'application/json' and 'text/event-stream'`。

最小 Python 流程骨架：

```python
import json, urllib.request
BASE = 'http://localhost:18060/mcp'
HEADERS = {'Content-Type': 'application/json', 'Accept': 'application/json, text/event-stream'}

init_payload = {
  'jsonrpc': '2.0', 'id': 1, 'method': 'initialize',
  'params': {
    'protocolVersion': '2025-03-26',
    'capabilities': {},
    'clientInfo': {'name': 'hermes', 'version': '1.0'},
  },
}
resp = urllib.request.urlopen(urllib.request.Request(BASE, data=json.dumps(init_payload).encode(), headers=HEADERS))
session_id = resp.headers.get('Mcp-Session-Id')
headers = {**HEADERS, 'Mcp-Session-Id': session_id}

urllib.request.urlopen(urllib.request.Request(
  BASE,
  data=json.dumps({'jsonrpc': '2.0', 'method': 'notifications/initialized', 'params': {}}).encode(),
  headers=headers,
)).read()

payload = {
  'jsonrpc': '2.0', 'id': 8, 'method': 'tools/call', 'params': {
    'name': 'publish_content',
    'arguments': {
      'account': 'brand-a',
      'title': '真正让女生变好看的，不只是护肤',
      'content': '正文内容，不要包含#标签',
      'images': ['/absolute/path/1.png', '/absolute/path/2.png'],
      'visibility': '仅自己可见',
      'tags': ['女生变美', '气质提升'],
    },
  },
}
print(urllib.request.urlopen(urllib.request.Request(BASE, data=json.dumps(payload).encode(), headers=headers), timeout=240).read().decode())
```

已知故障：如果返回 `发布失败: 没有找到发布 TAB - 上传图文`，通常不是图片或正文问题，而是 MCP 浏览器自动化没有点到发布页的“上传图文”入口。日志里常伴随大量 `发布 TAB 被遮挡，尝试移除遮挡`。这时优先排查/更新 xiaohongshu-mcp 上游选择器和发布流程，不要反复重做内容素材。

在 bee 的本地仓库 `/home/bee/apps/xiaohongshu-mcp`，fork 来源是 `xpzouying/xiaohongshu-mcp`。遇到这个发布 TAB 问题时，先确认 upstream 是否已合并：

```bash
cd /home/bee/apps/xiaohongshu-mcp
git remote -v
# 如没有 upstream：
git remote add upstream https://github.com/xpzouying/xiaohongshu-mcp.git
git fetch upstream --prune
# 保护本地账号配置等未提交改动
git status --short
git stash push -m 'temp-before-upstream-merge' <需要保护的文件>
git merge upstream/main
# 若提示 git 身份缺失，bee 偏好：
git config --global user.name 'Jarvis'
git config --global user.email 'bee.helper.ai@outlook.com'
# 合并后恢复 stash
git stash pop
```

已验证上游提交 `c63748f fix: 修复发布图文时找不到上传TAB的问题 (#666)` 对这个报错相关。合并后需要重启 `xiaohongshu-mcp` 容器再重试发布。
