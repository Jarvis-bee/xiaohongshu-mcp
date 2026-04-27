---
name: xhs-login
description: |
  管理小红书登录状态：检查是否已登录、二维码扫码登录、重置登录切换账号。
  当用户提到登录、扫码、账号、切换账号、退出登录、登录状态检查，或其他 skill 报告"未登录"需要先登录时使用。
---

## 执行流程

### 1. 检查登录状态

调用 `check_login_status`（无参数），返回是否已登录及用户名。

- 已登录 → 告知用户当前登录账号
- 未登录 → 进入步骤 2

### 2. 扫码登录

调用 `get_login_qrcode`（无参数）。MCP 工具返回两部分内容：
- 文本：超时提示（含截止时间）
- 图片：PNG 格式二维码（MCP image content type，Base64 编码）

**展示二维码**：MCP 返回的图片会通过客户端渲染给用户。如果客户端无法直接展示图片（如纯文本终端），则将 Base64 数据保存为临时 PNG 文件，告知用户文件路径让其手动打开：
```bash
# fallback: 保存二维码到临时文件
echo "<base64_data>" | base64 -d > /tmp/xhs-qrcode.png
open /tmp/xhs-qrcode.png   # macOS
xdg-open /tmp/xhs-qrcode.png  # Linux
```

提示用户：
- 打开小红书 App 扫描二维码
- 二维码有效期有限，过期需重新获取

扫码完成后，调用 `check_login_status` 确认登录成功。

### 3. 重新登录 / 切换账号

当用户要求重新登录或切换账号时：

1. 调用 `delete_cookies`（⚠️ 需用户确认）— 清除当前登录状态
2. 调用 `get_login_qrcode` — 获取新二维码
3. 引导用户扫码

## 约束

- `delete_cookies` 会清除登录状态，执行前必须确认
- 登录需要用户手动用手机 App 扫码，无法自动完成

## 失败处理

| 场景 | 处理 |
|---|---|
| MCP 工具不可用 | 引导用户使用 `/setup-xhs-mcp` 完成部署和连接配置 |
| 二维码超时 | 重新调用 `get_login_qrcode` |

## Hermes 当前会话里的 HTTP fallback

有时 Hermes 当前工具列表没有直接暴露 `check_login_status` / `get_login_qrcode` 这些 MCP 子工具，但本地 `xiaohongshu-mcp` 服务已经在 `http://localhost:18060/mcp` 运行。这时可以按 MCP HTTP 协议手动调用。

关键点：
- 先 `initialize`，读取响应头里的 `Mcp-Session-Id`
- 后续请求都带 `Mcp-Session-Id` 请求头
- 再发 `notifications/initialized`
- 最后调用 `tools/call`

示例：检查登录状态并获取二维码保存成本地图片：

```bash
python3 - <<'PY'
import json, urllib.request, base64
base = 'http://localhost:18060/mcp'
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

# check_login_status
_, status = mcp_call({'jsonrpc': '2.0', 'id': 2, 'method': 'tools/call', 'params': {'name': 'check_login_status', 'arguments': {}}}, sid)
print(status)

# get_login_qrcode
_, body = mcp_call({'jsonrpc': '2.0', 'id': 3, 'method': 'tools/call', 'params': {'name': 'get_login_qrcode', 'arguments': {}}}, sid)
obj = json.loads(body)
content = obj['result']['content']
text = next(x['text'] for x in content if x['type'] == 'text')
img_b64 = next(x['data'] for x in content if x['type'] == 'image')
out = '/home/bee/.hermes/tmp/xhs-qrcode.png'
with open(out, 'wb') as f:
    f.write(base64.b64decode(img_b64))
print(text)
print(out)
PY
```

交付二维码时用 `MEDIA:/home/bee/.hermes/tmp/xhs-qrcode.png`。
