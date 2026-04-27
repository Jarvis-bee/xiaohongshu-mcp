---
name: xhs-login
description: |
  管理小红书登录状态：检查是否已登录、二维码扫码登录、重置登录切换账号。
  当用户提到登录、扫码、账号、切换账号、退出登录、登录状态检查，或其他 skill 报告"未登录"需要先登录时使用。
---

## 多账号约定

- 当前 MCP 支持多账号，登录相关工具都支持可选 `account` 参数。
- 用户指定账号别名时，所有调用都传入该 `account`；未指定时使用 `default`。
- `account` 只支持字母、数字、`.`、`_`、`-`，不要把小红书昵称当成别名，除非用户明确这样配置。
- 每个账号的 cookies 独立保存到 `COOKIES_DIR/<account>/cookies.json`；未设置 `COOKIES_DIR` 时默认在 `~/.xiaohongshu-mcp/accounts/<account>/cookies.json`。
- “切换账号”优先理解为切换 `account` 参数；只有用户明确要重置某个账号登录态时，才对该账号调用 `delete_cookies`。

## 执行流程

### 1. 确定目标账号

从用户输入或上下文提取目标 `account`：

- 明确说“用 brand-a / 检查 brand-a / 登录 brand-a” → `account=brand-a`。
- 未指定 → 使用 `default`，调用时可省略 `account` 或传 `{"account": "default"}`。
- 账号别名不符合规则时，提示用户改成字母、数字、`.`、`_`、`-` 组成的别名。

### 2. 检查登录状态

调用 `check_login_status`：

```json
{
  "account": "brand-a"
}
```

- 已登录 → 告知用户当前 MCP 账号别名和返回的用户名。
- 未登录 → 进入步骤 3。

### 3. 扫码登录

调用 `get_login_qrcode`，必须使用同一个 `account`：

```json
{
  "account": "brand-a"
}
```

MCP 工具返回两部分内容：
- 文本：超时提示（含截止时间和账号别名）
- 图片：PNG 格式二维码（MCP image content type，Base64 编码）

**展示二维码**：MCP 返回的图片会通过客户端渲染给用户。如果客户端无法直接展示图片（如纯文本终端），则将 Base64 数据保存为临时 PNG 文件，文件名中带上账号别名，告知用户文件路径让其手动打开：

```bash
# fallback: 保存二维码到临时文件
echo "<base64_data>" | base64 -d > /tmp/xhs-qrcode-brand-a.png
open /tmp/xhs-qrcode-brand-a.png      # macOS
xdg-open /tmp/xhs-qrcode-brand-a.png  # Linux
```

提示用户：
- 打开小红书 App 扫描二维码
- 二维码有效期有限，过期需对同一 `account` 重新获取

扫码完成后，再调用 `check_login_status` 并携带同一 `account` 确认登录成功。

### 4. 重新登录 / 重置某个账号

当用户要求重新登录某个账号、清除登录态或二维码一直异常时：

1. 明确目标 `account`，并向用户确认要清除的是该账号别名的登录态。
2. 调用 `delete_cookies`（⚠️ 需用户确认）：

```json
{
  "account": "brand-a"
}
```

3. 调用 `get_login_qrcode`，继续使用同一个 `account`。
4. 引导用户扫码。

## 约束

- `delete_cookies` 只清除目标 `account` 的 cookies，但仍然是破坏性操作，执行前必须确认。
- 登录需要用户手动用手机 App 扫码，无法自动完成。
- 小红书网页端同一真实账号通常不建议在多个网页端同时登录；多账号应使用不同真实账号或明确知道会互相踢下线的风险。

## 失败处理

| 场景 | 处理 |
|---|---|
| MCP 工具不可用 | 引导用户使用 `/setup-xhs-mcp` 完成部署和连接配置 |
| 二维码超时 | 使用同一 `account` 重新调用 `get_login_qrcode` |
| 账号参数错误 | 提示账号别名只支持字母、数字、`.`、`_`、`-` |
| 指定账号未登录 | 不要回退到 `default`；继续引导登录该 `account` |

## Hermes 当前会话里的 HTTP fallback

有时 Hermes 当前工具列表没有直接暴露 `check_login_status` / `get_login_qrcode` 这些 MCP 子工具，但本地 `xiaohongshu-mcp` 服务已经在 `http://localhost:18060/mcp` 运行。这时可以按 MCP HTTP 协议手动调用。

关键点：
- 先 `initialize`，读取响应头里的 `Mcp-Session-Id`
- 后续请求都带 `Mcp-Session-Id` 请求头
- 再发 `notifications/initialized`
- 最后调用 `tools/call`
- `arguments` 中按需传 `account`；未指定时可使用 `default`

示例：检查指定账号登录状态并获取二维码保存成本地图片：

```bash
ACCOUNT=brand-a python3 - <<'PY'
import json, urllib.request, base64, os, re
base = 'http://localhost:18060/mcp'
account = os.environ.get('ACCOUNT', 'default')
if not re.fullmatch(r'[A-Za-z0-9._-]+', account):
    raise SystemExit('invalid ACCOUNT')
common = {'Content-Type': 'application/json', 'Accept': 'application/json, text/event-stream'}

def mcp_call(payload, sid=None, timeout=60):
    headers = common.copy()
    if sid:
        headers['Mcp-Session-Id'] = sid
    req = urllib.request.Request(base, data=json.dumps(payload).encode(), headers=headers)
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return r.headers, r.read().decode()

h, _ = mcp_call({
  'jsonrpc': '2.0', 'id': 1, 'method': 'initialize',
  'params': {'protocolVersion': '2025-03-26', 'capabilities': {}, 'clientInfo': {'name': 'hermes-debug', 'version': '1.0'}}
})
sid = h.get('Mcp-Session-Id')
mcp_call({'jsonrpc': '2.0', 'method': 'notifications/initialized', 'params': {}}, sid)
args = {'account': account}

# check_login_status
_, status = mcp_call({'jsonrpc': '2.0', 'id': 2, 'method': 'tools/call', 'params': {'name': 'check_login_status', 'arguments': args}}, sid)
print(status)

# get_login_qrcode
_, body = mcp_call({'jsonrpc': '2.0', 'id': 3, 'method': 'tools/call', 'params': {'name': 'get_login_qrcode', 'arguments': args}}, sid)
obj = json.loads(body)
content = obj['result']['content']
text = next(x['text'] for x in content if x['type'] == 'text')
img_b64 = next(x['data'] for x in content if x['type'] == 'image')
out = f'/home/bee/.hermes/tmp/xhs-qrcode-{account}.png'
with open(out, 'wb') as f:
    f.write(base64.b64decode(img_b64))
print(text)
print(out)
PY
```

交付二维码时用 `MEDIA:/home/bee/.hermes/tmp/xhs-qrcode-<account>.png`。
