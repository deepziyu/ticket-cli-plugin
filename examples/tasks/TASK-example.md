---
id: TASK-example
title: Implement debounce on login button
type: tasks
status: open
priority: minor
owner: dev-agent
created_at: 2026-06-11 14:15
updated_at: 2026-06-11 14:15
module: ui
---

# TASK-example: Implement debounce on login button

## 描述
在登录按钮上添加前端防抖逻辑，防止多次重复提交。

## 验收标准
1. 点击“登录”按钮后，300ms 内再次点击无效。
2. 按钮在点击后应显示 Loading 状态。
