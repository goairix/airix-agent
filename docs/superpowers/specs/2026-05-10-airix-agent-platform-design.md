# Airix Agent 平台架构设计

**日期：** 2026-05-10  
**项目：** github.com/goairix/airix-agent  
**语言：** Go  
**AI 框架：** 字节跳动 Eino  
**架构模式：** DDD + 六边形架构（核心单体 + 沙盒独立进程）

---

## 一、整体系统架构

### 1.1 系统分层总览

```
┌─────────────────────────────────────────────────────────────┐
│                      业务系统（调用方）                        │
│              Hospital A / Hospital B / ...                   │
└──────────────┬──────────────────────────┬───────────────────┘
               │ REST / SSE               │ gRPC
               ▼                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    接口层 (interfaces/)                       │
│   HTTP Handler  │  gRPC Service  │  WebSocket/SSE Handler   │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    应用层 (application/)                      │
│  AgentApp │ KnowledgeApp │ ToolApp │ ModelApp │ WorkspaceApp │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    领域层 (domain/)                           │
│  agent │ knowledge │ tool │ model │ mcp │ skill │ workspace  │
│  sandbox │ session │ memory                                  │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  基础设施层 (infrastructure/)                  │
│  Eino Runtime │ VectorStore │ DB │ Cache │ MQ │ Storage      │
└──────────────────────────┬──────────────────────────────────┘
                           │ gRPC
┌──────────────────────────▼──────────────────────────────────┐
│                   沙盒执行器（独立进程）                        │
│         Code Runner │ Tool Isolator │ VM Manager            │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 工作空间（Workspace）隔离模型

工作空间是资源隔离的基本单位，**所有资源必须归属于某个工作空间，不存在全局共享资源**。

**权限模型：**
- **超级管理员**：唯一可以创建、删除、管理工作空间的角色，可以为普通管理员指定其管理的工作空间
- **普通管理员**：由超级管理员指定，可管理一个或多个工作空间内的资源，无法跨越被授权范围

```
SuperAdmin
├── 创建/删除 Workspace
├── 指定 Admin → Workspace A、Workspace B
└── 指定 Admin → Workspace C

Admin（张三）
├── Workspace A（院区一）
│   ├── Agents / KnowledgeBases / Tools
│   ├── ModelInstances / Skills / MCPServers
│   └── ThirdPartyServices
└── Workspace B（院区二）
    ├── Agents / KnowledgeBases / Tools
    └── ...

Admin（李四）
└── Workspace C（院区三）
    └── ...
```

### 1.3 核心领域边界

| 领域 | 职责 | 关键聚合根 |
|------|------|-----------|
| `workspace` | 工作空间生命周期、成员权限 | Workspace |
| `agent` | Agent 定义、配置、生命周期、版本发布 | Agent, AgentRelease |
| `session` | 对话会话、执行轨迹、Token 统计 | Session, Message |
| `memory` | 记忆存储与检索策略 | Memory |
| `knowledge` | 知识库、文档、向量索引 | KnowledgeBase, Document |
| `tool` | 工具注册、调用协议 | Tool |
| `model` | 模型厂商、模型实例管理 | ModelProvider, ModelInstance |
| `mcp` | MCP Server 注册与代理 | MCPServer |
| `skill` | Skill 注册、版本、调用 | Skill |
| `sandbox` | 沙盒会话、执行请求 | SandboxSession |

---

## 二、Agent 领域设计

### 2.1 Agent 类型与 Eino 映射

| Agent 类型        | 说明                          | Eino ADK 实现                                         |
|------------------|-----------------------------|----------------------------------------------------|
| `ReAct`          | ReAct 范式，工具调用循环          | `adk.NewChatModelAgent`                            |
| `TextGeneration` | 单次文本生成，无工具调用            | `adk.NewChatModelAgent`（MaxIterations=1，无 Tools） |
| `Supervisor`     | 中心化多 Agent 协调，主控分配子任务  | `supervisor.New`                                   |
| `PlanExecute`    | 规划-执行-反思循环，处理复杂多步任务  | `planexecute.New`（Planner + Executor + Replanner） |
| `DeepAgent`      | MainAgent + 专项子 Agent，上下文隔离 | `deep.New`                                         |
| `Super`          | 自适应综合体，自主选择执行策略        | `ChatModelAgent` + 元工具（见 2.5）                  |
| `Claw`           | 用户虚拟机控制（未来）             | `VMSession` + `adk.NewChatModelAgent`              |

### 2.2 Agent 聚合根结构

```
Agent（聚合根）
├── AgentID
├── WorkspaceID
├── Name / Description
├── AgentType（ReAct / TextGeneration / Supervisor / PlanExecute / DeepAgent / Super / Claw）
├── ModelConfig
│   ├── ModelInstanceID     ← 绑定模型管理模块的实例
│   └── Parameters（temperature, max_tokens 等）
├── SystemPrompt
├── ToolBindings[]          ← 绑定的工具列表
├── KnowledgeBindings[]     ← 绑定的知识库列表
├── SkillBindings[]         ← 绑定的 Skill 列表
├── MCPBindings[]           ← 绑定的 MCP Server 列表
├── MemoryConfig
│   ├── MemoryDriverType（default / claw）
│   ├── ContextMode（sliding_window / summarization）
│   ├── ContextWindowSize       ← sliding_window 模式：保留最近 N 轮
│   ├── SummarizationConfig     ← summarization 模式
│   │   ├── SummaryModelInstanceID  ← 用于生成摘要的模型实例
│   │   └── TriggerTokenThreshold   ← 触发压缩的 token 阈值
│   └── GlobalMemoryMode（full / search）
├── CollaborationConfig（可选，Supervisor / PlanExecute / DeepAgent / Super 时有值）
│   ├── SubAgentIDs[]       ← 子 Agent 列表（Supervisor / DeepAgent / Super 使用）
│   └── TransferPolicy      ← Transfer 触发条件（deterministic / llm_driven）
├── SandboxConfig
│   ├── Enabled
│   └── SandboxType（process / container / vm）
└── Status（draft / active / disabled）
```

### 2.3 Agent 运行时流程

```
调用方 → AgentApp.Run(sessionID, input)
    │
    ├─ 1. 加载 Agent 配置
    ├─ 2. 创建/恢复 Session（含 InterruptState 恢复判断）
    ├─ 3. 从 ModelManager 获取 ChatModel 实例
    ├─ 4. 从 MemoryManager 加载 GlobalMemory → 注入 SystemPrompt
    ├─ 5. 组装 AgentRunner（绑定工具/知识库/Skill/记忆工具）
    │   ├─ 工具数量 > 阈值 → 注入 ToolSearchMiddleware
    │   ├─ 上下文策略 = summarization → 注入 SummarizationMiddleware
    │   ├─ 启用工具结果压缩 → 注入 ToolReductionMiddleware
    │   ├─ AgentType = DeepAgent / Super → 注入 PlanTaskMiddleware + FileSystemMiddleware
    │   ├─ AgentType = Super → 额外注入 plan_execute / delegate 元工具
    │   └─ SkillBindings 非空 → 注入 SkillMiddleware
    ├─ 6. AgentRunner.Query / Resume（含 Interrupt & Resume 支持）
    │   ├─ 工具调用 → ToolRouter
    │   │   ├─ 普通工具 → 直接执行
    │   │   ├─ 沙盒工具 → SandboxClient.Execute()
    │   │   └─ memory_search → SessionMemoryStore.Search()
    │   ├─ 知识库检索 → KnowledgeRetriever（作为 Tool）
    │   ├─ MCP 调用 → MCPProxy
    │   ├─ Agent 协作 → Transfer / ToolCall（NewAgentTool）
    │   └─ 人工中断 → StatefulInterrupt → 持久化 InterruptState → 返回 interrupted
    ├─ 7. 流式输出 → SSE/WebSocket 推送给调用方
    ├─ 8. 持久化 Session（消息、轨迹、Token 消耗）
    └─ 9. 异步提取记忆 → SessionMemoryStore（领域事件驱动）
```

### 2.4 Session 与消息持久化

Session 完整持久化，支持审计、上下文组装和前端展示。Messages 和执行步骤由独立表存储，Session 聚合根不在内存中持有消息集合。

```
Session（聚合根）
├── SessionID
├── AgentID / WorkspaceID
├── UserID              ← 发起会话的用户，用于关联记忆
├── Status（running / interrupted / completed / failed）
├── InterruptState（可选，Status = interrupted 时有值）
│   ├── InterruptID         ← ADK StatefulInterrupt 的唯一标识，用于 ResumeWithParams
│   ├── CheckPointData      ← 序列化的 CheckPointStore 快照
│   └── PendingContext      ← 等待人工输入的上下文描述
└── TotalTokenUsage
```

消息存储见第三章。

### 2.5 Super Agent 设计

Super Agent 是自适应综合体，同一个 Agent 实例根据任务复杂度在运行时自主选择执行策略，无需外部预判。

**策略选择逻辑**

```
用户输入
  │
  ▼
SuperAgent（ChatModelAgent 作为决策核心）
  │
  ├─ 判断：简单任务 → 直接 ReAct（普通工具调用循环）
  │
  ├─ 判断：需要分步执行 → 调用 plan_execute 元工具
  │       → 内部启动 PlanExecute 子流程
  │         ├─ Planner：生成结构化任务步骤
  │         ├─ Executor：逐步执行，每步调用工具
  │         └─ Replanner：评估是否继续或终止
  │
  └─ 判断：需要专项能力或并行处理 → 调用 delegate 元工具
          → 将子任务分发给对应专项子 Agent（通过 NewAgentTool）
            ├─ SubAgent A（如 ResearchAgent）
            ├─ SubAgent B（如 CodeAgent）
            └─ SubAgent C（如 WorkflowAgent）
            → 汇总子 Agent 结果后继续决策
```

**运行时装配**

Super Agent 在 AgentRunner 组装阶段注入以下元工具和中间件：

```
元工具
├── plan_execute    ← 触发内部 PlanExecute 子流程
└── delegate        ← 将子任务通过 NewAgentTool 分发给指定子 Agent

中间件（在普通 AgentType 基础上强制开启）
├── PlanTaskMiddleware      ← 跟踪子任务执行状态
├── ToolSearchMiddleware    ← 工具集动态管理（子 Agent 多时必需）
└── SummarizationMiddleware ← 长流程上下文自动压缩
```

**配置结构**

```
CollaborationConfig（Super 时）
├── SubAgentIDs[]       ← 可调度的专项子 Agent 列表
├── TransferPolicy      ← llm_driven（由 Agent 自主决定何时 delegate）
└── MaxDelegationDepth  ← 防止无限递归的最大分发层数（默认 2）
```

**与其他类型的关系**

| 对比 | Super | DeepAgent |
|------|-------|-----------|
| 子 Agent 调度时机 | LLM 自主决定 | MainAgent ReAct 循环中按需调用 |
| 是否支持 PlanExecute | 是（内置元工具） | 否（需单独配置） |
| 适用场景 | 任务类型不确定、高度复杂 | 任务类型固定、子 Agent 职责明确 |

### 2.6 Agent 版本发布

Agent 支持版本管理，`AgentRelease` 作为独立聚合根，与 Agent 通过 `AgentID` 松耦合关联。Agent 聚合根新增 `ActiveReleaseID` 字段（草稿态时为空）。

#### AgentRelease 聚合根

```
AgentRelease（聚合根）
├── ReleaseID              ← 时间戳格式，如 20260510-143022
├── AgentID                ← 关联的 Agent
├── WorkspaceID
├── ReleasedAt             ← 发布时间
├── ReleasedBy             ← 发布人
├── ChangeLog              ← 变更说明（可选）
├── Status（active / inactive）  ← 同一 Agent 只有一个 active
│
└── Snapshot               ← 深度快照，发布时固化
    ├── AgentConfig        ← Agent 所有配置字段的完整拷贝
    │   ├── Name / Description / AgentType
    │   ├── SystemPrompt
    │   ├── ModelConfig
    │   ├── MemoryConfig
    │   ├── CollaborationConfig
    │   └── SandboxConfig
    ├── ToolSnapshots[]    ← 快照时刻工具的完整配置
    ├── KnowledgeSnapshots[] ← 知识库配置快照
    ├── SkillSnapshots[]   ← Skill 完整内容快照
    └── MCPSnapshots[]     ← MCP Server 配置快照
```

#### 发布与回滚流程

```
发布流程
  Agent（draft）→ AgentApp.Publish()
    ├── 深度读取当前 Agent + 所有绑定资源配置
    ├── 创建 AgentRelease（ReleaseID = 时间戳）
    ├── 将旧 active Release 置为 inactive
    ├── 新 Release.Status = active
    └── 更新 Agent.ActiveReleaseID = 新 ReleaseID

回滚流程
  AgentApp.Rollback(agentID, releaseID)
    ├── 读取目标 AgentRelease.Snapshot
    ├── 基于 Snapshot 覆写 Agent 当前草稿配置
    └── Agent.ActiveReleaseID 不变（仍指向当前 active 版本）
    ※ 回滚不自动发布，需用户确认后再次调用 Publish()
```

#### 调用接口

业务系统发起会话时，支持两种方式指定版本：

```
# 默认走当前 active 版本
POST /sessions  { agent_id: "xxx" }

# 指定特定发布版本
POST /sessions  { release_id: "20260510-143022" }
```

运行时解析优先级：
- `release_id` 有值 → 加载对应 `AgentRelease.Snapshot` 运行
- 否则 → 加载 `Agent.ActiveReleaseID` 对应 Snapshot 运行
- 若 `ActiveReleaseID` 为空（纯草稿）→ 报错，不允许调用

#### 仓储接口

```go
// domain/agent/repository
type AgentReleaseRepository interface {
    Save(ctx context.Context, release AgentRelease) error
    GetByID(ctx context.Context, releaseID string) (AgentRelease, error)
    GetActive(ctx context.Context, agentID string) (AgentRelease, error)
    ListByAgent(ctx context.Context, agentID string, pagination Pagination) ([]AgentRelease, error)
}
```

---

## 三、上下文与记忆管理

### 3.1 消息存储模型

会话消息按三层结构持久化：**Session → Message → MessageItem**。

#### Message（一轮对话）

一"轮"定义为：用户发送 query，经过若干次 LLM 调用、工具调用、知识库检索，直到 LLM 输出最终回复，整个过程为一轮。

```
Message
├── id（UUID）
├── session_id
├── workspace_id（冗余，便于查询）
├── agent_id（冗余）
├── sort_order          ← 填充时间戳，(session_id, sort_order) 联合索引
├── query               ← 用户原始输入
├── agent_input         ← Agent 注入的变量参数（jsonb，后续提示词管理设计）
├── status              ← running / completed / failed / interrupted
├── total_tokens        ← 整轮 token 合计
├── input_tokens
├── output_tokens
├── cached_tokens
├── execution_time_ms   ← 整轮执行耗时
├── first_token_latency_ms
├── created_at
└── completed_at
```

#### MessageItem（轮次内步骤）

一个 Message 对应多个 MessageItem，记录轮次内每个步骤（LLM 回复、工具调用、错误等）。

```
MessageItem
├── id（UUID）
├── message_id
├── session_id（冗余）
├── sort_order          ← 应用层自增（从 0 开始），无索引
│                         取出整个 Message 的 items 后在内存中 sort.Sort 排序
├── item_type           ← thinking / assistant / tool_call / error
├── is_final            ← 是否为最终回复（item_type=assistant 时有意义）
├── content（jsonb）    ← 所有类型的内容统一存于此字段
├── input_tokens        ← 本次 LLM 调用输入 token（thinking/assistant 时有值）
├── output_tokens       ← 本次 LLM 调用输出 token
├── latency_ms          ← 本步骤耗时
└── created_at
```

**content 各类型结构：**

```json
// thinking（模型深度思考内容）
{ "text": "..." }

// assistant（LLM 回复，含中间回复和最终回复）
{ "text": "..." }

// tool_call（工具调用，知识库检索同此结构）
{
  "tool_name": "search_kb",
  "tool_call_id": "call_abc123",
  "arguments": { ... },
  "result": { ... },
  "error": "..."
}

// error
{ "code": "xxx", "message": "..." }
```

#### LLM Context 组装规则

按 `message.sort_order` + `message_item.sort_order` 顺序，将 MessageItem 映射为 LLM messages 格式：

- `thinking` → `role: assistant`（thinking block）
- `assistant` → `role: assistant`
- `tool_call` → 拆为两条：`role: assistant`（tool_calls）+ `role: tool`（tool result）
- `error` → 不注入 LLM context，仅用于展示和审计

### 3.2 上下文（Context Window）

上下文是当前会话内传给 LLM 的消息窗口，支持两种策略：

**Sliding Window（默认）**

- 取当前 Session 最近 N 轮 Message 的所有 MessageItem，按顺序组装 LLM messages
- N 由 Agent 配置的 `ContextWindowSize` 决定
- 简单高效，适合对话轮次较短的场景

**Summarization（自动压缩）**

- 使用 ADK `SummarizationMiddleware`，在 token 数超过 `TriggerTokenThreshold` 时自动压缩历史对话
- 压缩时生成摘要注入上下文，保持连续性；可选保留原始用户消息
- 适合长对话、多工具调用的 ReAct / DeepAgent 场景
- 配置 `SummaryModelInstanceID` 指定摘要专用模型，与推理模型解耦

### 3.3 记忆（Memory）

记忆分两层，与上下文完全独立：

**全局记忆（Global Memory）**

- 归属于 `WorkspaceID + UserID`，跨会话持久
- 存储用户偏好、重大事件、长期背景信息
- 注入 `system prompt`，量小时全量注入，量大时语义检索后注入
- 由 Agent 在会话结束后根据策略自动提取并写入，也支持手动管理

**会话记忆（Session Memory）**

- 归属于某个 Session，按日期存储
- 存储历史会话的关键信息摘要（非原始消息）
- 不主动注入，作为 `memory_search` 工具暴露给 Agent
- Agent 在 ReAct 循环中按需调用，自主决定何时回忆历史

### 3.4 Memory 实体结构

```
Memory（实体）
├── MemoryID
├── WorkspaceID
├── UserID
├── MemoryType（session / global）
├── SessionID（session 类型时有值）
├── Content             ← 结构化摘要或自然语言片段
├── Tags[]              ← 便于检索的标签
├── Importance          ← 重要性评分，影响检索排序
└── CreatedAt / Date    ← 按日期组织
```

### 3.5 运行时 Prompt 组装

```
system:    [SystemPrompt] + [GlobalMemory]
messages:  [最近 N 轮 Context Window]
tools:     [...其他工具, memory_search]
input:     [当前用户输入]
```

### 3.6 记忆抽象层

记忆模块做可插拔抽象，不同 Agent 类型使用不同驱动：

```go
// domain/memory/port
type SessionMemoryStore interface {
    Save(ctx context.Context, entry MemoryEntry) error
    Search(ctx context.Context, query string, opts SearchOptions) ([]MemoryEntry, error)
    ListByDate(ctx context.Context, userID string, date time.Time) ([]MemoryEntry, error)
}

type GlobalMemoryStore interface {
    Upsert(ctx context.Context, entry GlobalMemoryEntry) error
    LoadAll(ctx context.Context, userID string) ([]GlobalMemoryEntry, error)
    Search(ctx context.Context, userID string, query string) ([]GlobalMemoryEntry, error)
}

type MemoryExtractor interface {
    Extract(ctx context.Context, session Session) ([]MemoryEntry, error)
}
```

| 驱动 | 适用场景 |
|------|---------|
| `DefaultMemoryDriver` | 标准 Agent，按日期存 DB + 向量检索；SessionMemoryStore 的存取后端复用 `filesystem.Backend` 接口抽象 |
| `ClawMemoryDriver` | Claw 模式，直接接入 ADK `FileSystem Middleware`，由中间件自动注入文件读写工具，实现结构化文件层级上下文记忆 |

Agent 配置时指定 `MemoryDriverType`，运行时由工厂方法实例化对应驱动，上层不感知具体实现。

`MemoryExtractor` 在 `AgentSessionCompleted` 事件的消费方中异步执行，对应 ADK `AfterAgent` 钩子时机，不阻塞主流程。

---

## 四、模型管理

### 4.1 领域设计

模型管理分两个聚合根：**ModelProvider（厂商）** 和 **ModelInstance（实例）**。

```
ModelProvider（聚合根）
├── ProviderID
├── Name（OpenAI / Anthropic / Aliyun / Ollama / ...）
├── Protocol（openai-compatible / anthropic / custom）
├── BaseURL
├── AuthType（api-key / oauth / none）
└── SupportedCapabilities[]（chat / embedding / rerank / tts / stt）

ModelInstance（聚合根）
├── InstanceID
├── WorkspaceID
├── ProviderID
├── ModelName（gpt-4o / claude-3-5-sonnet / qwen-max / ...）
├── Capability（chat / embedding / rerank）
├── APIKey（加密存储）
├── Parameters（默认 temperature、max_tokens 等）
├── RateLimit（RPM / TPM）
└── Status（active / disabled）
```

### 4.2 协议适配层

Eino 已抽象了 `ChatModel` / `Embedder` 接口，适配器层按协议实现：

```
ModelAdapterFactory
├── OpenAICompatibleAdapter   ← 覆盖 OpenAI、Aliyun、Ollama、DeepSeek 等
├── AnthropicAdapter
├── GoogleAdapter
└── CustomAdapter             ← 自定义 HTTP 协议，配置请求/响应映射
```

Agent 运行时调用 `ModelManager.GetChatModel(instanceID)` 拿到 Eino `ChatModel`，不感知厂商细节。

### 4.3 Embedding 与 Rerank

模型管理同时管理 Embedding 模型和 Rerank 模型，供知识库检索使用：

- `ModelManager.GetEmbedder(instanceID)` → Eino `Embedder`
- `ModelManager.GetReranker(instanceID)` → Eino `Reranker`

知识库配置时绑定具体的 Embedding 实例 ID，解耦模型选择与知识库逻辑。

---

## 五、知识库

### 5.1 领域设计

```
KnowledgeBase（聚合根）
├── KBID
├── WorkspaceID
├── Name / Description
├── EmbeddingModelInstanceID  ← 绑定 Embedding 模型
├── RerankModelInstanceID     ← 可选，绑定 Rerank 模型
├── VectorStoreConfig
│   ├── DriverType（milvus / qdrant / pgvector / weaviate）
│   └── CollectionName
├── ChunkConfig
│   ├── ChunkSize / Overlap
│   └── SplitStrategy（sentence / paragraph / semantic）
└── Status

Document（实体）
├── DocumentID / KBID
├── Title / SourceType（file / url / manual）
├── ParseStatus（pending / processing / done / failed）
├── ChunkCount
└── Metadata

Chunk（实体）
├── ChunkID
├── DocumentID / KBID
├── Content                 ← 原始文本
├── Metadata
│   ├── PageNumber
│   ├── Position
│   └── CustomTags          ← 手动添加分片时可打标签
└── CreatedAt
```

向量数据存在 VectorStore，DB 只存 Chunk 元数据，通过 ChunkID 关联。

### 5.2 增量更新方式

一个知识库包含多个 Document，支持三种增量更新：

1. **上传新文件** → 解析（PDF/Word/MD）→ 分块 → Embed → 写入 VectorStore（追加）
2. **直接添加分片** → 跳过解析和分块，直接提供文本 → Embed → 写入 VectorStore
3. **更新已有文件** → 删除旧分片 → 重新处理 → 写入（替换）

### 5.3 向量存储可插拔设计

```go
// domain/knowledge/port
type VectorStore interface {
    Upsert(ctx context.Context, collection string, docs []VectorDoc) error
    Search(ctx context.Context, collection string, query Vector, opts SearchOptions) ([]SearchResult, error)
    Delete(ctx context.Context, collection string, ids []string) error
}
```

驱动实现放在 `infrastructure/adapter/vectorstore/`：

```
vectorstore/
├── milvus/
├── qdrant/
├── pgvector/
└── weaviate/
```

### 5.4 RAG 检索流程

知识库检索以 Tool 形式暴露给 Agent（`knowledge_search`），Agent 在 ReAct 循环中按需调用，检索结果作为工具返回值进入 messages：

```
Agent 调用 knowledge_search(query)
  → Embedder.Embed(query)
  → VectorStore.Search(collection, vector, topK)
  → Reranker.Rerank(results)（可选）
  → 返回 TopN 片段文本 → 进入 messages（role: tool）
```

文档处理管道（异步，领域事件驱动）：

```
DocumentUploaded 事件
  → 解析器（PDF/Word/MD）
  → 分块器
  → Embedder.BatchEmbed()
  → VectorStore.Upsert()
  → 更新 Document.ParseStatus = done
```

---

## 六、工具管理

### 6.1 工具分类

| 类型 | 说明 | 执行位置 |
|------|------|---------|
| `BuiltinTool` | 平台内置，Go 代码实现，随服务部署 | 主进程 |
| `APITool` | 第三方 API 工具，配置化定义（OpenAPI / 自定义） | 主进程 |
| `SandboxTool` | 用户自定义代码工具 | 沙盒执行器 |

### 6.2 工具聚合根

```
Tool（聚合根）
├── ToolID
├── WorkspaceID
├── Name / Description
├── ToolType（builtin / api / sandbox）
├── InputSchema         ← JSON Schema，LLM 用来生成调用参数
├── OutputSchema
├── Config
│   ├── BuiltinConfig
│   │   └── HandlerName     ← 注册的 Go handler 名称
│   ├── APIConfig
│   │   ├── BaseURL / Method / Path
│   │   ├── AuthType（none / api-key / oauth2 / bearer）
│   │   ├── Headers / QueryParams
│   │   └── RequestMapping / ResponseMapping
│   └── SandboxConfig
│       ├── Runtime（python / nodejs / bash）
│       ├── Code
│       └── Dependencies[]
└── Status
```

### 6.3 工具执行路由

```go
// domain/tool/port
type ToolExecutor interface {
    Execute(ctx context.Context, toolID string, input map[string]any) (any, error)
}
```

运行时根据 `ToolType` 路由到不同执行器：

```
ToolRouter
├── BuiltinToolExecutor   → 直接调用注册的 Go handler
├── APIToolExecutor       → 构造 HTTP 请求，调用第三方 API
└── SandboxToolExecutor   → 通过 gRPC 发送到沙盒执行器
```

Eino 的工具调用通过 `ToolsNode` 接入，`ToolRouter` 实现 Eino 的 `Tool` 接口，Agent 运行时统一调用，不感知工具类型。

### 6.4 大规模工具集管理

当 Agent 绑定工具数量较多（工具数 > 阈值，或 MCPServer 动态发现工具较多）时，全量传给模型会导致 token 超限。通过 ADK 中间件按需管理：

**ToolSearchMiddleware（动态工具选择）**

- 注入元工具 `tool_search`，初始只对模型暴露工具名和描述
- Agent 在 ReAct 循环中通过正则搜索按需激活工具，模型只处理过滤后的工具列表

**ToolReductionMiddleware（工具结果压缩）**

- 工具返回内容超过 `MaxLengthForTrunc` 时，自动截断并将完整内容存入 Backend，返回位置提示
- 历史工具结果累计 token 超过 `MaxTokensForClear` 时，将旧结果卸载到文件系统，模型通过 `read_file` 按需读取

两者可与 `SummarizationMiddleware` 组合使用，在 AgentRunner 组装阶段按配置条件注入。

---

## 七、MCP 集成

### 7.1 MCPServer 聚合根

```
MCPServer（聚合根）
├── ServerID
├── WorkspaceID
├── Name / Description
├── TransportType（stdio / sse / streamable-http）
├── Config
│   ├── StdioConfig
│   │   ├── Command / Args[]
│   │   └── Env{}
│   ├── SSEConfig
│   │   └── URL / Headers
│   └── StreamableHTTPConfig
│       └── URL / Headers
├── AuthConfig
└── Status（connected / disconnected / error）
```

### 7.2 MCP 代理层

MCP Server 注册后，平台维护长连接（或按需连接），对 Agent 暴露统一工具调用接口：

```
MCPProxy
├── 启动时连接所有 active MCPServer
├── 动态发现 MCP Server 暴露的 Tools / Resources / Prompts
├── 将 MCP Tools 包装成 Eino Tool 接口
└── Agent 绑定 MCPServer 后，其工具自动合并进 Agent 工具列表
```

MCP 工具对 Agent 透明，与 BuiltinTool / APITool 统一走 `ToolRouter`，底层路由到 `MCPToolExecutor`。

---

## 八、第三方服务集成

第三方服务是平台层面的服务集成（短信、邮件、支付、TTS、OCR 等），与 APITool 的核心区别：

- **APITool**：由 LLM 决定何时调用，面向 Agent 工具调用场景
- **ThirdPartyService**：平台基础能力，由应用层或 BuiltinTool 直接调用，不依赖 LLM 驱动

### 8.1 ThirdPartyService 聚合根

```
ThirdPartyService（聚合根）
├── ServiceID
├── WorkspaceID
├── Name / Description
├── ServiceType（sms / email / tts / payment / ocr / custom）
├── ProviderName（aliyun / tencent / custom / ...）
├── Config（加密存储，含 API Key、endpoint 等）
└── Status（active / disabled）
```

### 8.2 服务适配层

```go
// domain/thirdparty/port
type ServiceProvider interface {
    Call(ctx context.Context, method string, params map[string]any) (map[string]any, error)
}
```

驱动实现放在 `infrastructure/adapter/thirdparty/`，按服务类型和厂商组织：

```
thirdparty/
├── tts/
│   ├── aliyun/
│   └── tencent/
├── sms/
│   ├── aliyun/
│   └── tencent/
└── email/
    └── smtp/
```

### 8.3 与工具层的关系

第三方服务可以被 BuiltinTool 封装后暴露给 Agent：

```
BuiltinTool（send_sms）
  → ThirdPartyServiceManager.Get(serviceID)
  → ServiceProvider.Call("send", params)
```

这样 Agent 通过工具调用间接使用第三方服务，服务本身不感知 Agent 的存在。

---

## 九、Skill 管理

### 9.1 Skill 聚合根

Skill 是可复用的能力单元，可以是 Prompt 模板、子 Agent、或工具组合。

```
Skill（聚合根）
├── SkillID
├── WorkspaceID
├── Name / Description      ← 渐进式披露的摘要，Agent 初始只看到这两个字段
├── Version
├── SkillType（prompt / sub-agent / tool-chain）
├── Config
│   ├── PromptConfig
│   │   └── Template / Variables[]
│   ├── SubAgentConfig
│   │   └── AgentID         ← 指向另一个 Agent
│   └── ToolChainConfig
│       └── ToolIDs[]       ← 顺序执行的工具链
├── InputSchema / OutputSchema
└── Status
```

### 9.2 Skill 运行时机制

ADK `SkillMiddleware` 采用**渐进式披露（Progressive Disclosure）**机制：

```
Agent 初始状态
  → 只看到 Skill 的 Name 和 Description（摘要）
  → 通过 skill_search / get_skill 元工具按需获取完整 SKILL.md 内容
  → 获取完整指令后执行对应逻辑

Skill 存储格式（运行时动态生成，兼容 ADK 目录结构）
  skill_{skillID}/
  └── SKILL.md    ← 包含 Name、Description、完整使用说明、参数格式
```

平台层 Skill 聚合根在运行时由 `SkillMiddleware` 的 Backend 适配器动态渲染为 ADK 所需的 SKILL.md 格式，上层不感知存储细节。

### 9.3 Skill 抽象层

```go
// domain/skill/port
type SkillExecutor interface {
    Execute(ctx context.Context, skillID string, input map[string]any) (any, error)
}

type SkillRegistry interface {
    Register(ctx context.Context, skill Skill) error
    Get(ctx context.Context, skillID string) (Skill, error)
    List(ctx context.Context, workspaceID string) ([]Skill, error)
}
```

| 驱动 | 适用场景 |
|------|---------|
| `DefaultSkillDriver` | 标准 Agent，Prompt / SubAgent / ToolChain；由平台 SkillMiddleware Backend 适配器渲染为 ADK 格式 |
| `ClawSkillDriver` | Claw 模式，直接基于文件系统发现和加载 Skill，与 ADK FileSystem Middleware 协同 |

### 9.4 Skill 调用链

Skill 的 Auth、限流、日志等横切关注点在**应用层**实现，不在 ADK Middleware 层：

```
Agent 触发 skill_search / get_skill（ADK SkillMiddleware 注入的元工具）
  → SkillMiddleware.Backend 查询 SkillRegistry
  → 返回完整 Skill 内容给 Agent
  → Agent 执行 Skill 指令
      ├── 应用层 SkillApp.Invoke() 校验权限 / 限流 / 记录 ExecutionTrace
      └── SkillExecutor.Execute()（具体驱动实现）
```

### 9.5 Skill 调用方式

- **Agent 绑定 Skill**：通过 ADK `SkillMiddleware` 渐进式披露，Agent 在 ReAct 循环中按需发现并调用
- **直接调用**：应用层提供 `SkillApp.Invoke()` 接口，业务系统可直接调用 Skill 而不经过完整 Agent

---

## 十、沙盒执行器

### 10.1 整体设计

沙盒执行器是独立进程，通过 gRPC 与主服务通信，负责三类隔离执行：

```
主服务
  └─ SandboxClient（gRPC）
        │
        ▼
  沙盒执行器（独立进程）
  ├── CodeRunner        ← 用户自定义代码执行
  ├── ToolIsolator      ← 工具调用隔离
  └── VMManager         ← Claw 模式用户虚拟机管理
```

### 10.2 沙盒抽象层

```go
// domain/sandbox/port
type SandboxSession interface {
    Execute(ctx context.Context, req ExecuteRequest) (<-chan ExecuteEvent, error)
    Terminate(ctx context.Context) error
}

type SandboxFactory interface {
    Create(ctx context.Context, config SandboxConfig) (SandboxSession, error)
}
```

| 实现 | 生命周期 | 交互模式 |
|------|---------|---------|
| `ProcessSession` | 短，按次执行 | 输入代码 → 输出结果 |
| `ContainerSession` | 中，按任务 | 输入代码 → 输出结果 |
| `VMSession` | 长，跨会话持久 | 双向控制流（Action → Observation） |

### 10.3 三种隔离级别

**Process 隔离**（默认）
- OS 进程 + seccomp/namespace 限制系统调用
- 适合代码工具执行，轻量快速
- 无网络访问，文件系统沙盒化

**Container 隔离**
- Docker/containerd 容器
- 适合需要依赖安装的复杂工具
- 可配置受限网络访问

**VM 隔离**（Claw 模式）
- Firecracker microVM 或 QEMU
- 用户独立虚拟机，完整操作系统环境
- Agent 通过 VM 内的 Agent 进程控制桌面/终端
- 支持持久化（VM 快照）

### 10.4 VMSession 扩展接口

Claw 模式下 Agent 与 VM 的交互是 Action/Observation 循环：

```go
type VMSession interface {
    SandboxSession
    SendAction(ctx context.Context, action VMAction) (Observation, error)
    GetObservation(ctx context.Context, obsType ObservationType) (Observation, error)
    Snapshot(ctx context.Context, name string) error
    Restore(ctx context.Context, name string) error
}

type VMAction struct {
    ActionType string         // shell / click / type / scroll / ...
    Payload    map[string]any
}

type Observation struct {
    ObsType string            // terminal / screenshot / file
    Content []byte
}
```

Agent 运行时通过 `VMSession.SendAction()` 驱动 VM，通过 `GetObservation()` 感知环境状态，整个循环由 Eino AgentLoop 编排。

### 10.5 gRPC 接口定义

```protobuf
service SandboxService {
    rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);
    rpc Execute(ExecuteRequest) returns (stream ExecuteResponse);
    rpc TerminateSession(TerminateSessionRequest) returns (TerminateSessionResponse);
    rpc GetSession(GetSessionRequest) returns (GetSessionResponse);
}

message ExecuteRequest {
    string session_id = 1;
    string code = 2;
    map<string, string> env = 3;
    int32 timeout_seconds = 4;
}

message ExecuteResponse {
    string stdout = 1;
    string stderr = 2;
    int32 exit_code = 3;
    bool finished = 4;
}
```

### 10.6 主服务集成点

```
domain/sandbox/port
└── SandboxClient interface
    ├── CreateSession()
    ├── Execute()
    └── Terminate()

infrastructure/adapter/sandbox/
└── GRPCSandboxClient  ← 实现 SandboxClient 接口
```

---

## 十一、可观测性与整体数据流

### 11.1 Trace（链路追踪）

每次 Agent 运行创建一条完整 Trace，基于现有 OTel 基础设施扩展：

```
AgentRun (root span)
├── ContextLoad              ← 加载上下文/记忆
├── AgentRunnerExecute       ← AgentRunner 执行（含 Interrupt & Resume）
│   ├── LLMCall              ← 模型调用（含 model、tokens、latency）
│   ├── ToolCall             ← 工具调用（含 toolID、input、output）
│   ├── KnowledgeSearch      ← 知识库检索
│   ├── SkillInvoke          ← Skill 调用
│   └── AgentTransfer        ← Agent 协作转移（Transfer / ToolCall 模式）
├── SandboxExecute           ← 沙盒执行（如有）
└── SessionPersist           ← 会话持久化
```

### 11.2 Metrics（指标）

```
airix_agent_run_total{workspace, agent_id, status}
airix_agent_run_duration_seconds{workspace, agent_id}
airix_llm_tokens_total{workspace, model_instance, type}   // input/output/cached
airix_tool_call_total{workspace, tool_id, status}
airix_knowledge_search_duration_seconds{workspace, kb_id}
airix_sandbox_execute_duration_seconds{workspace, runtime}
```

### 11.3 Token 消耗统计

```
TokenUsageRecord
├── WorkspaceID / AgentID / SessionID
├── ModelInstanceID
├── InputTokens
├── OutputTokens
├── CachedTokens        ← 部分模型支持 prompt cache
├── TotalTokens         ← InputTokens + OutputTokens + CachedTokens
└── CreatedAt
```

支持按工作空间、Agent、模型实例多维度查询。结构化日志统一携带 `workspace_id`、`agent_id`、`session_id`，便于按会话过滤。

### 11.4 整体数据流

```
业务系统
  │ REST/gRPC/WebSocket
  ▼
接口层
  │ DTO → Command/Query
  ▼
应用层（AgentApp.Run / AgentApp.Resume）
  ├─ WorkspaceService    → 校验工作空间权限
  ├─ SessionService      → 创建/恢复 Session（含 InterruptState 判断）
  ├─ ModelManager        → 获取 ChatModel 实例
  ├─ MemoryManager       → 加载 GlobalMemory → 注入 SystemPrompt
  └─ ToolRouter          → 注册所有工具（Builtin/API/Sandbox/MCP/Skill/Memory）
  │
  ▼
AgentRunner（含 ChatModelAgentMiddleware 链）
  ├─ ToolSearchMiddleware    → 动态工具选择（工具数量多时）
  ├─ SummarizationMiddleware → 上下文自动压缩（长对话时）
  ├─ ToolReductionMiddleware → 工具结果压缩（结果过长时）
  ├─ SkillMiddleware         → Skill 渐进式披露
  ├─ FileSystemMiddleware    → 文件系统访问（Claw / DeepAgent）
  ├─ PlanTaskMiddleware      → 任务管理（DeepAgent）
  │
  ├─ LLM Call            → ModelAdapter → 厂商 API（流式）
  └─ Tool Call           → ToolRouter
      ├─ BuiltinTool     → Go handler
      ├─ APITool         → HTTP 请求
      ├─ SandboxTool     → gRPC → 沙盒执行器
      ├─ MCPTool         → MCPProxy → MCP Server
      ├─ AgentTool       → Transfer / NewAgentTool（多 Agent 协作）
      └─ memory_search   → SessionMemoryStore.Search()
  │
  ├─ Interrupt           → 持久化 InterruptState → Status = interrupted → 返回调用方
  │
  ▼
流式输出 → SSE/WebSocket → 业务系统
  │
  ▼
Session 持久化
  ├─ Messages / ExecutionTrace → DB
  ├─ TokenUsageRecord → DB
  └─ AgentSessionCompleted 事件 → MemoryExtractor（异步，对应 AfterAgent 钩子时机）
```

### 11.5 领域事件

| 事件 | 触发时机 | 消费方 |
|------|---------|-------|
| `AgentSessionCompleted` | 会话正常结束 | MemoryExtractor（异步提取记忆）、Token 统计 |
| `AgentSessionInterrupted` | 会话被 HITL 中断 | 通知业务系统等待人工输入 |
| `DocumentUploaded` | 文档上传 | 文档解析管道 |
| `DocumentParsed` | 解析完成 | Embed + 写入 VectorStore |
| `SandboxSessionExpired` | 沙盒超时 | 资源回收 |
| `AgentPublished` | Agent 发布新版本 | 通知关联业务系统版本变更、审计日志 |
| `AgentRollbacked` | 回滚操作创建新草稿 | 审计日志 |
| `ModelInstanceDisabled` | 模型实例禁用 | 通知关联 Agent |
