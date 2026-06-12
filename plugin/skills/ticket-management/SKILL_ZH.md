---
name: ticket-management
description: Manage, query, transition, and validate ticket-like Markdown files (e.g. bugs, tasks) using the 'ticket' CLI command.
---

# 🎫 工单与缺陷管理（Ticket Management）技能

[English Version](./SKILL.md)

本技能专门面向 AI Agent 及人类工程师，提供了一套通过 `ticket` 命令行工具来对本地目录下的 Markdown 工单/缺陷（Bugs）文件进行标准化检索、生命周期流转及格式校验的协议。

通过使用 `ticket` 命令，Agent 可以避免大范围遍历、解析原始 Markdown 文件的巨额 Token 开销，并降低由于直接文本修改带来的 Match Error。

---

## 🎯 触发条件

在会话中遇到以下任务时，Agent 应主动激活本技能：
1. **获取工单/BUG 列表**：需要了解当前目录下有哪些待处理、进行中或已解决的工单。
2. **更新工单状态**：对某个工单进行状态流转（例如从 `open` 流转到 `fixing` 或 `passed`），或添加处理结论。
3. **新建工单**：在测试用例跑通失败时，需要输出一个缺陷报告。
4. **校验工单规范**：对目录下的所有工单进行 YAML Frontmatter 元数据合规校验。

---

## 🛠️ CLI 命令行操作指南

在执行本技能时，Agent 应优先通过 Shell 运行 `ticket` 命令，而非直接编辑文件：

### 1. 检索工单列表 (`ticket list`)
*   **指令**：`ticket list [flags] [dir]`
*   **AI 推荐参数**：必须指定 `--format json` 标志以获取机器友好、体积紧凑的 JSON 数据。
    ```bash
    ticket list --format json --status open ./tests/02-cases/v0.1/bugs
    ```
*   **常用过滤 Flag**：
    *   `--status`：按状态过滤 (`open` / `fixing` / `resolved` / `passed` / `rejected`)
    *   `--priority`：按优先级过滤 (`critical` / `major` / `minor` / `low`)
    *   `--owner`：按负责人过滤
    *   `--type`：按类型过滤 (`bug` / `task` / `feature`)
    *   `--format`：格式化类型 (`table` 对人 / `json` 对 AI)

### 2. 更新工单属性与状态 (`ticket update`)
*   **指令**：`ticket update [flags] <file_path>` (注：为了提高易用性，flags 参数也可以置于文件路径之后，如 `ticket update ./bugs/BUG-01.md --status passed`)
*   **参数说明**：
    *   `--status <status>`：流转状态（流转为 `passed` 或 `rejected` 终态时，系统会自动将 `resolved_at` 设为当前时间；流转回 `open` 等状态时，会清除 `resolved_at`）。
    *   `--owner <owner>`：指派给指定负责人。
    *   `--priority <priority>`：调整优先级。
    *   `--conclusion "<text>"`：写入工单的处理结论/修复说明。
    *   `--field key=value`：写入或更新自定义的 YAML Frontmatter 元数据属性（例如：`--field module=sop --field trace_id=t-12345`）。
    *   `--format <text|json>`：输出格式。指定 `json` 时将直接返回更新后的完整工单 JSON 数据；指定 `text` (默认值) 时将输出美化的已修改字段差异列表。
*   **示例**：
    ```bash
    ticket update --status passed --conclusion "通过复用现有派发配置修复" ./bugs/BUG-S1-03-20260611.md
    ```

### 3. 创建工单 (`ticket create`)
*   **指令**：`ticket create [flags] <dir_path> <id>`
*   **示例**：
    ```bash
    ticket create --title "SOP 验收规则 L0 审核页面加载 500 报错" --type bug --status open --priority major --owner dev-agent ./bugs BUG-S1-05-20260611
    ```
*   **常用 Flag**：
    *   `--title`：标题
    *   `--type`：类型
    *   `--status`：初始状态
    *   `--priority`：优先级
    *   `--owner`：负责人
    *   `--body`：自定义 Markdown 正文
    *   `--field key=value`：自定义元数据扩展字段

### 4. 查看单个工单详情 (`ticket show`)
*   **指令**：`ticket show [flags] <file_path>`
*   **参数说明**：
    *   `--format json`：输出包含完整元数据和 body 文本的 JSON 对象。
    *   `--format text`：输出人类直观阅读的表单排版。
*   **示例**：
    ```bash
    ticket show --format json ./bugs/BUG-S1-03-20260611.md
    ```

### 5. 校验目录规范 (`ticket validate`)
*   **指令**：`ticket validate [flags] [dir_path]`
*   **参数说明**：
    *   `--format <text|json>`：输出格式。若无校验错误，指定 `json` 时返回 `[]` 数组。
*   **示例**：
    ```bash
    ticket validate --format json ./bugs
    ```

### 6. 批量迁移旧工单 (`ticket migrate`)
*   **指令**：`ticket migrate [flags] [file_or_dir_path]`
*   **功能说明**：扫描目录下或针对单个文件进行迁移。解析文件名中的 legacy 状态后缀（如 `-pass`、`-passed`、`-fail` 等）写入 Frontmatter `status` 字段中，并自动重命名文件，补齐缺失的元数据（如 `title`、`type`、时间字段等）。
    *   **幂等性与最小写入**：如果工单已标准化且无任何内容或状态变化，它在二次运行时将被完全跳过（Skipped），不触发任何磁盘写回，亦不会改变文件的修改时间，避免污染 Git 工作区。
*   **参数说明**：
    *   `--dry-run`：仅预览迁移计划而不修改磁盘。
    *   `--only-invalid`：仅迁移在校验中为无效（即不符合规范）的工单文件。
    *   `--format <text|json>`：输出格式。指定 `json` 时以结构化数组形式返回所有的迁移明细（包含 `file`、`action`、`reason`、`changed_fields` 等属性）。
*   **示例**：
    ```bash
    # 迁移目录
    ticket migrate ./bugs
    # 仅预览迁移
    ticket migrate --dry-run ./bugs
    ```

### 7. 启动可视化看板面板 (`ticket dashboard`)
*   **指令**：`ticket dashboard [flags] [dir_path]`
*   **功能说明**：在本地启动一个轻量级、高颜值的可视化 Kanban 看板 Web 应用（默认运行在 `http://localhost:8080`），并自动在浏览器中打开。看板采用与 Linear/TAPD 相似的高级暗黑视觉设计，且支持直接通过 HTML5 拖拽（Drag-and-Drop）卡片来实时流转工单状态、新增/删除工单、以及进行 Markdown 描述的在线可视化编写与实时预览。
*   **参数说明**：
    *   `--port` 或 `-p`：指定本地 Web 服务器的监听端口（默认值为 `8080`）。
*   **示例**：
    ```bash
    ticket dashboard --port 18080 ./bugs
    ```

---

## ⚙️ 目录自定义配置 (`ticket.yaml`)

工单根目录或子目录下可以放置 `ticket.yaml` 配置文件，用于定义子目录分类和自定义必填字段校验。

### 1. 配置格式示例
```yaml
# ticket.yaml
sub_dirs:
  - bugs
  - tasks

extra_fields:
  - name: module
    required: true
  - name: severity
```

### 2. 核心校验规则
1. **显式子目录扫描**：若配置了 `sub_dirs`，则只扫描/递归扫描该列表下的子目录。默认不配置 `sub_dirs` 时，将**不进行递归读取**（仅扫描当前目录）。若子目录内有进一步的 `ticket.yaml`，则按其内部的 `sub_dirs` 继续递归。
2. **分类与 Type 强绑定**：工单 Frontmatter 中的 `type` 字段必须强制与工单文件所在的目录名（转换为小写）完全一致。在 `ticket create` 和 `ticket update` 时，CLI 会**自动修正并覆盖**该值为当前目录名。
3. **必填额外字段校验**：在 `ticket validate` 时，会对 `ticket.yaml` 中标记为 `required: true` 的额外字段进行非空校验。子目录若未配置自己的 `ticket.yaml`，将默认继承父目录的 `extra_fields` 校验规则。

---

## 🚫 限制与注意事项

1. **不要手动修改 Frontmatter 格式**：除非 `ticket update` 报错或环境受阻，Agent 必须通过运行 CLI 命令来流转工单状态。这可以避免拼写错误或破坏 YAML 的格式。
2. **保持文件名 Immutable（不变）**：工单创建后其文件名应当保持不变，状态的变化完全反映在 Frontmatter 顶部的 `status` 字段中。
3. **类型与目录名一致**：工单的 `type` 字段在创建与更新时会自动对齐其所在的物理文件夹名（小写）。请确保工单存放在对应的物理文件夹中。
