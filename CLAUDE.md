# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Airix Agent Platform — 企业级 AI Agent 开发平台，支持多类型智能体编排、知识库管理、工具集成和多模型 LLM 交互。使用 Go 1.25 编写，模块路径 `github.com/dysodeng/app`。

## 常用命令

```bash
make init          # 首次初始化：安装工具、创建 .env、安装 git hooks
make run           # 启动开发服务器（hot reload，依赖 fresh）
make wire          # 重新生成依赖注入代码（Google Wire）
make proto         # protobuf 完整流程：lint → format → generate
make test          # 运行所有测试：go test -v ./...
make lint          # 运行 golangci-lint
```

运行单个测试：
```bash
go test -v -run TestFuncName ./path/to/package/...
```

修改 `internal/di/` 中的 Provider 或依赖接线后，必须执行 `make wire` 重新生成 `wire_gen.go`。**禁止直接手动编辑 `wire_gen.go`。**

## 架构

**DDD（领域驱动设计）+ 整洁架构**，多协议服务器并发运行（HTTP/gRPC/WebSocket/Event/Health）。

### 分层结构（`internal/`）

- **`interfaces/`** — 表现层：HTTP 控制器（Gin）、gRPC 处理器、WebSocket 处理器、DTO、中间件、路由、校验器
- **`application/`** — 应用层：用例编排、应用服务
- **`domain/`** — 领域层：实体、值对象、领域服务、仓储接口（port）
- **`infrastructure/`** — 基础设施层：持久化（GORM）、外部服务适配、事件消费、服务器实现
- **`di/`** — Google Wire 依赖注入接线
- **`infrastructure/pkg/`** — 共享工具包：logger (zap)、telemetry (OpenTelemetry)、redis、token、storage、crypto

**依赖方向：** `interfaces → application → domain ← infrastructure`，内层绝不引用外层。

### 服务器

`internal/infrastructure/server/` 包含独立的服务器实现：HTTP（8080）、gRPC（3000）、WebSocket、Event 消费者、Health（5000）。所有服务器通过 `di` 注入，在 `cmd/app/` 统一启动。

### 核心领域

`internal/domain/` 下按限界上下文组织：`agent`、`session`、`memory`、`knowledge`、`tool`、`model`、`mcp`、`skill`、`sandbox`、`user`、`file`、`permission`。

每个领域内部结构：
```
domain/<context>/
├── model/         # 实体和聚合根
├── valueobject/   # 值对象（含校验）
├── repository/    # 仓储接口（在 infrastructure 中实现）
├── service/       # 领域服务（跨聚合逻辑）
├── event/         # 领域事件
└── errors/        # 领域特定错误码
```

### 应用层模式

- `application/<domain>/dto/` — 命令（写操作入参）和响应 DTO
- `application/<domain>/decorator/` — 横切关注点（链路追踪、日志）装饰器
- `application/<domain>/event/handler/` — 领域事件处理器

## 配置

配置加载优先级：`.env` > `configs/config.yaml` > etcd（可选配置中心）。

环境变量命名规则：`SECTION_SUBSECTION_KEY`（如 `MAIN_DB_HOST`、`MONITOR_TRACER_OTLP_ENABLED`）。

配置结构定义在 `internal/infrastructure/config/config.go`。

## API 规范

- **统一响应格式**：`{ "code": 200, "data": {}, "message": "", "trace_id": "" }`
- **响应码**：200（成功）、0（业务错误）、400、401、403、404、500
- **分页结构**：`{ "data": { "record": [...], "total": 100 } }`
- **参数校验**：`binding:"required"` + `msg:"中文错误提示"` tag
- **链路追踪头**：OpenTelemetry 自动注入 `X-Trace-Id`、`X-Span-Id`

## 错误处理

分层错误传递：`infrastructure → domain → application → interfaces`。

- **基础设施层**：直接返回原始错误（GORM、Redis、gRPC），不做包装
- **领域层**：使用 `DomainError{Code, Message, Err}` 包装。错误定义在 `domain/<模块>/errors/`，支持 `errors.Is()` 按 Code 判等
- **应用层**：透传领域错误，通过 `logger.Error(ctx, err.Error())` 记录日志。参数校验失败直接 `errors.New("中文描述")` 返回
- **表现层**：通过 `err.Error()` 提取错误消息，返回给前端

规则：不跨层包装；日志只在应用层记录一次；错误信息使用中文（直接返回给前端）。

## 数据库规范

**双模型严格分离（禁止混用）：**
- 领域模型（`domain/<模块>/model/`）— 纯业务对象，无 GORM tag
- 数据实体（`internal/infrastructure/persistence/entity/<模块>/`）— GORM 模型，含 tag

数据实体嵌入 `model.DistributedPrimaryKeyID`（UUID 主键）+ `model.Time`（created_at/updated_at）+ `model.SoftDelete`（软删除），必须实现 `TableName()` 方法。

**模型转换**：在仓储实现内部定义私有方法，`fromModel()`（数据实体→领域模型）和 `toModel()`（领域模型→数据实体）。

**数据库迁移**：gormigrate，位于 `internal/infrastructure/migration/<模块>.go`，ID 格式 `<表名>_<时间戳>`（如 `user_202510061500`）。启动时自动执行。

GORM 数据实体字段 tag 必须包含 `type`、`not null`、`comment`。

## Wire 依赖注入规范

模块 Set 位于 `internal/di/modules/<模块>.go`，注册顺序：仓储实现 → 领域服务 → 应用服务 → 控制器。

所有业务模块的 Set 统一在 `internal/di/module.go` 的 `ModulesSet` 中汇总。

新增依赖步骤：
1. 编写 `NewXxxService` 构造函数（入参为接口类型，返回值为接口类型）
2. 注册到对应模块的 `wire.NewSet`
3. 新模块需在 `internal/di/module.go` 的 `ModulesSet` 中引入
4. 执行 `make wire`，提交 `wire_gen.go`

构造函数返回具体类型时，使用 `wire.Bind(new(接口), new(*实现))` 绑定。

## 代码风格

- 注释和领域术语使用**中文**
- import 分组顺序：标准库 → 第三方库 → 内部包（`github.com/dysodeng/app`），由 goimports 自动处理
- 接口定义与实现分离：接口（大写导出）定义在同文件顶部，实现结构体（小写未导出）紧随其后。如 `AgentService` 接口 + `agentService` 结构体
- 类型名、函数名禁止以包名开头（避免 `agent.AgentService`、`config.ConfigItem` 这类重复）
- 领域枚举类型使用 `type XxxType uint8`，配合 `const iota`、`String()`、`Validate()` 方法
- 领域模型实体提供 `Validate()` 方法做自身业务规则校验
- 每个服务结构体包含 `baseTraceSpanName` 字段，格式 `<层>.<模块>.service.<ServiceName>`，每个方法入口创建 span
- 构造函数命名 `NewXxxService`，入参为接口类型，返回值为接口类型
- 定义类型时，必要的注释要加上

## DDD 约束

**领域层（domain）**：
- 纯业务逻辑，不依赖任何框架（无 GORM、Gin、Wire 等 import）
- 仓储接口定义在 `domain/<模块>/repository/`，实现在 `infrastructure/persistence/repository/<模块>/`
- 领域错误使用 `DomainError{Code, Message}`，定义在 `domain/<模块>/errors/`
- 领域事件继承 `BaseDomainEvent`，定义在 `domain/<模块>/event/`

**应用层（application）**：
- 接收原始类型参数（`string`、`uint8`），内部转换为领域类型
- 不直接操作数据库，通过领域服务完成业务
- 事务通过 `TransactionManager` 管理，领域层不感知事务

**表现层（interfaces）**：
- 控制器只做参数解析和调用应用服务，不调用领域服务或仓储
- 校验失败通过 validator 翻译错误信息

## Git 规范

提交格式：`<type>(<scope>): <subject>` — type 可选：feat/fix/hotfix/docs/style/refactor/perf/test/build/ci/chore/revert。subject 6~50 字符。

提交消息中**不允许**添加任何署名信息（如 `Co-Authored-By`、`Signed-off-by` 等）。

Pre-commit hook 执行链：goimports → gofmt → go vet → shadow → golangci-lint。任一检查不通过则拒绝提交。

## 与用户交互使用中文
