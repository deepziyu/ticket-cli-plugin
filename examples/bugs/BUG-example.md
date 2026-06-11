---
id: BUG-example
title: Login button fails to respond on double click
type: bugs
status: open
priority: major
owner: dev-agent
created_at: 2026-06-11 14:15
updated_at: 2026-06-11 18:15
module: auth
severity: high
---
# BUG-example: Login button fails to respond on double click

## 描述
在用户快速双击登录按钮时，请求可能会失败或抛出 500 异常，需要对此进行防抖控制。

## 测试步骤与实际结果
1. 访问登录页面。
2. 连续快速点击“登录”按钮。
3. 观察到发送了两次相同的 HTTP 登录请求，后端处理时由于并发冲突报错。