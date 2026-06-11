---
id: IGNORED-example
title: This file is in an ignored directory
type: ignored
status: open
priority: low
owner: dev-agent
created_at: 2026-06-11 14:15
updated_at: 2026-06-11 14:15
---

# IGNORED-example

## 描述
此文件所处的文件夹 `ignored/` 未在 `ticket.yaml` 的 `sub_dirs` 中配置。
因此，当执行 `ticket list examples` 或 `ticket validate examples` 时，该文件夹下的所有工单都将被自动忽略，不会被读取或校验。
