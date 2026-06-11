# Ticket CLI 示例与场景测试用例说明 (Examples & Test Cases)

本目录既是 `ticket-cli-plugin` 的用户配置参考指南，也是用于功能验证的场景测试用例目录（Scenario Tests）。它全面展示了通过 `ticket.yaml` 进行自定义分类、递归限制和额外必填字段校验的核心功能。

---

## 📂 目录结构说明

```text
examples/
├── ticket.yaml             # 自定义配置文件（定义递归目录与必填字段）
├── README.md               # 本文档
├── bugs/
│   └── BUG-example.md      # 标准 BUG 工单（包含 required 的 module 属性）
├── tasks/
│   └── TASK-example.md     # 标准 TASK 工单（包含 required 的 module 属性）
└── ignored/
    └── IGNORED-example.md  # 忽略目录下的工单（未在 ticket.yaml 中注册）
```

---

## ⚙️ 配置文件解析 (`ticket.yaml`)

```yaml
# 显式注册允许递归扫描的子目录（对应工单的 type）
sub_dirs:
  - bugs
  - tasks

# 定义额外字段的校验规则
extra_fields:
  - name: module
    required: true       # 标记为必填项。在 validate 时如果缺失或值为空将报错
  - name: severity
    required: false      # 标记为选填项。仅起展示或占位作用，不强制要求
```

---

## 🧪 场景测试用例 (Scenario Test Cases)

请使用 PowerShell 7 终端，进入 `examples` 目录执行以下测试用例：

### 用例 1：选择性扫描与分类查询 (Selective Recursion Scan)
*   **目的**：验证工具是否只扫描 `sub_dirs` 下注册的子目录，并忽略未注册的子目录。
*   **命令**：
    ```powershell
    ..\plugin\commands\ticket.exe list --format json .
    ```
*   **预期输出**：
    - 返回包含 `BUG-example` 和 `TASK-example` 两个工单信息的 JSON 数组。
    - `ignored/IGNORED-example.md` 工单**不应**出现在列表中。
    - 各工单的 `type` 分别被自动识别为所在的目录名（`bugs` 和 `tasks`）。

---

### 用例 2：完整目录合规校验 (Successful Validation)
*   **目的**：验证在所有工单满足规则时，校验是否能够顺利通过。
*   **命令**：
    ```powershell
    ..\plugin\commands\ticket.exe validate --format json .
    ```
*   **预期输出**：
    - 返回 `null`（或命令行以 `0` 状态码成功结束，没有错误输出）。

---

### 用例 3：额外字段必填项校验 (Required Field Check)
*   **目的**：验证当工单缺失 `ticket.yaml` 中配置的 `required: true` 字段时，校验能否正确拦截。
*   **步骤**：
    1. 打开 `tasks/TASK-example.md` 文件。
    2. 将 `module: ui` 这一行删去并保存。
    3. 运行校验命令：
       ```powershell
       ..\plugin\commands\ticket.exe validate --format json .
       ```
*   **预期输出**：
    - 返回校验失败的 JSON 信息，包含：
      `"ErrorMsg": "missing required extra field 'module' specified in ticket.yaml"`
    - 命令行退出状态码不为 0。
*   **恢复**：测试完成后，重新把 `module: ui` 加回 `tasks/TASK-example.md` 的 Frontmatter 头并保存。

---

### 用例 4：类型与目录一致性强校验 (Type Directory Match Check)
*   **目的**：验证工单的 `type` 字段是否与所在的物理目录名强制保持一致。
*   **步骤**：
    1. 打开 `bugs/BUG-example.md` 文件。
    2. 将 Frontmatter 里的 `type: bugs` 修改为 `type: tasks`（故意混淆）并保存。
    3. 运行校验命令：
       ```powershell
       ..\plugin\commands\ticket.exe validate --format json .
       ```
*   **预期输出**：
    - 返回错误信息：
      `"ErrorMsg": "invalid ticket type 'tasks', must match directory name 'bugs'"`
*   **恢复**：测试完成后，重新将 `type` 改回 `bugs` 并保存。

---

### 用例 5：创建工单时自动修正类型 (Auto-Correct Type on Create)
*   **目的**：验证在创建新工单时，工具能否自动检测目标目录名，并修正/填入正确的 `type` 属性。
*   **命令**：
    ```powershell
    # 这里故意传入参数 --type task，但目标目录为 bugs 
    ..\plugin\commands\ticket.exe create --title "Temp Bug" --type task bugs BUG-temp
    ```
*   **预期输出**：
    - 成功在 `bugs/` 下创建 `BUG-temp.md` 文件。
    - 打开 `bugs/BUG-temp.md`，其 YAML 头部中的 `type` 属性已被**自动修正/覆盖**为 `bugs`（而非命令行传入的 `task`）。
*   **清理**：验证完毕后删除 `bugs/BUG-temp.md`。

---

### 用例 6：更新工单与类型保持 (Auto-Correct Type on Update)
*   **目的**：验证在更新工单时，工具始终会自动对齐物理目录名。
*   **命令**：
    ```powershell
    ..\plugin\commands\ticket.exe update --status open --field module=auth bugs/BUG-example.md
    ```
*   **预期输出**：
    - 成功更新该工单。即使工单中有遗留的不规范的 `type` 字段，更新后也会被重新对齐为所在的物理子目录名（`bugs`）。
