# Kimi Config Platform - 基于 Git + Starlark 的动态配置系统

## 背景

### 现有配置平台的问题

Kimi iOS App 启动时会通过 `AppConfigService` 拉取服务端配置，接口是 `Kimi_Gateway_Config_V1_GetConfigRequest`。配置在 Web 管理后台（[outsight.dev.kimi.team/ui/config/app-config](https://outsight.dev.kimi.team/ui/config/app-config)）按**平台 × 最低版本 × 语言 × 地区**维度存储。

**核心痛点：维度组合爆炸。**

例如想让 `kimiplusPlaza` 在 **2.5.5 及以上版本** 生效：
- 2.5.5 × 中文 × 国内
- 2.5.5 × 中文 × 海外
- 2.5.5 × 英文 × 国内
- 2.5.5 × 英文 × 海外
- 2.6.6 还要再配 4 套...

10 个版本 × 3 种语言 × 2 个地区 = **60 条配置记录**，而且每条都是**覆盖式完整配置**，不是增量/差量式。

### 目标

让开发者用 Python-like 语法编写配置规则，后端按客户端请求的维度实时执行脚本生成配置，彻底解决"改一处配 N 处"的问题。

---

## 技术方案选型

### 方案对比

| 方案 | 优点 | 缺点 |
|------|------|------|
| A. 文件系统 + 内存缓存 | 极简，无依赖 | 无历史版本 |
| **B. Git 后端 + 文件系统** | **天然版本管理、可追溯、可回滚** | 多一个 Git 依赖 |
| C. 对象存储 (S3) | 分布式部署 | MVP 阶段过度设计 |

**选择方案 B：Git 存储脚本，每次发布 = 一次 Git commit。**

### 配置脚本语言选择

| 方案 | 优点 | 缺点 |
|------|------|------|
| A. 系统 Python | 最灵活 | 安全风险（文件/网络访问）|
| B. 嵌入式 Python | 性能好 | 复杂度高 |
| C. Python 微服务 | 独立服务 | 架构复杂 |
| **D. Starlark** | **Python 语法、沙箱安全、确定性执行** | 功能子集 |

**选择 Starlark** —— Google 出品的 Python 子集，Bazel 大量使用。开发者写起来和 Python 几乎一样，但执行环境完全受限（禁止文件 I/O、网络访问），Go 有官方实现 `google/starlark-go`。

### 技术栈

| 环境 | 技术栈 |
|------|--------|
| **线上** | Go 1.23 + Vercel Serverless Functions + starlark-go + GitHub API |
| **本地** | Go 1.23 + Gin + starlark-go + go-git |
| **前端** | React 19 + TypeScript + Vite + Monaco Editor |
| **配置语言** | Starlark (Python 子集) |
| **存储** | Git 仓库（GitHub API 线上 / 文件系统本地）|

---

## 架构设计

### 部署架构（线上）

全栈部署在 **Vercel**（免费计划，无需信用卡）：

```
┌─────────────────────────────────────────────────────────────────┐
│                         Vercel                                  │
│  ┌──────────────┐      ┌─────────────────────┐                 │
│  │  前端 (React) │──────│  Serverless Functions│                │
│  │  静态站点     │      │  (Go + starlark-go) │                │
│  └──────────────┘      └──────────┬──────────┘                 │
│                                   │                             │
│                                   ▼                             │
│                          ┌─────────────────┐                    │
│                          │   GitHub API    │                    │
│                          │  (读写 .star 脚本)│                    │
│                          └─────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
                                    ▲
                                    │
                              App 客户端
                          (iOS / Android)
```

| 组件 | 部署方式 | 说明 |
|------|----------|------|
| 前端 | Vercel 静态站点 | `kimi-config-web/` 构建为静态文件 |
| 后端 | Vercel Serverless Functions (Go) | `api/index.go` 处理所有 API 请求 |
| 数据存储 | GitHub API | 配置文件通过 GitHub REST API 读写，天然 Git 版本管理 |

**环境变量（Vercel）**：
- `GITHUB_TOKEN` — GitHub Personal Access Token（`repo` 权限）
- `GITHUB_REPO` — 仓库名，如 `dormitory219/kimi-config-platform`
- `GITHUB_CONFIG_PATH` — 配置文件目录，如 `kimi-config-server/config-repo`

### 本地开发架构

```
React 前端 (Monaco Editor)  <--->  Go HTTP API (Gin)  <--->  Git 仓库 (*.star)
                                          │                     (本地文件系统)
                                          v
                                    App 客户端 (iOS/Android)
```

### 数据流

1. 开发者在 React 前端选择平台，页面默认加载 latest 配置版本
2. latest 配置版本只读；如需修改，点击 **New Draft**
3. 后端复制 latest 内容，返回下一个配置版本号（如 `v2 -> v3 draft`）和脚本内容
4. 开发者编辑 draft，右侧可 Preview、History、Diff
5. 点击 **Save** → 保存 draft；该配置版本成为新的 latest
6. 点击 **Publish** → 通过 GitHub API 创建 commit（线上）或 `git commit`（本地）
7. App 启动时请求 `GetConfig(version, lang, region)` → 后端读取对应平台脚本并执行 Starlark → 返回 JSON 配置

这里有两个不同的版本概念：

- **配置版本**：管理后台里的 `v1` / `v2` / `v3`，用于配置脚本演进
- **客户端 App 版本**：Preview 和 `/v1/config` 请求中的 `version=2.5.5`，会传入 Starlark 的 `ctx.version`

### Git 仓库结构

```
config-repo/
├── ios.star          # iOS legacy v1 配置脚本
├── android.star      # Android legacy v1 配置脚本
├── ios/
│   ├── v2.star       # iOS v2 配置脚本
│   └── v3.star       # iOS v3 配置脚本
├── android/
│   └── v2.star       # Android v2 配置脚本
├── lib/              # 共享库（可被主脚本 load）
│   └── common.star
└── README.md
```

为了兼容早期实现，根目录的 `<platform>.star` 被识别为 `v1`。保存新版本时写入 `<platform>/vN.star`，例如 `ios/v2.star`。

### Starlark DSL 设计

```starlark
def build_config(ctx):
    # ctx 包含：platform, version, language, region
    # 这里的 ctx.version 是客户端 App 版本，如 "2.5.5"
    config = {
        "system": {},
        "taskbar": {"items": []},
        "model": {},
    }

    # 语言配置
    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"
        config["taskbar"]["items"].append({"type": "DEEP_RESEARCH", "title": "深度研究"})
    elif ctx.language == "en":
        config["system"]["greeting"] = "Hello"
        config["taskbar"]["items"].append({"type": "DEEP_RESEARCH", "title": "Deep Research"})
    else:
        fail("unsupported language: " + ctx.language)

    # 版本配置
    if ctx.version >= "2.5.5":
        config["system"]["kimiplusPlaza"] = {"enabled": True}

    return config
```

**关键设计：没有默认配置。**

每个字段都通过条件分支显式设置，未知语言/版本直接 `fail` 报错，不会在运行时静默返回错误配置。

### 后端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/platforms` | 列出所有平台脚本 |
| GET | `/api/scripts/:platform/versions` | 列出配置版本、latest 和 nextVersion |
| POST | `/api/scripts/:platform/drafts` | 从 latest 复制内容，创建下一版本草稿响应 |
| GET | `/api/scripts/:platform?version=vN` | 读取指定配置版本脚本内容 |
| GET | `/api/scripts/:platform/history?version=vN` | 指定配置版本的 Git commit 历史 |
| POST | `/api/scripts/:platform?version=vN` | 保存指定配置版本脚本 |
| POST | `/api/scripts/:platform/publish?version=vN` | 发布指定配置版本（git commit）|
| POST | `/api/preview` | 预览配置（执行脚本）|
| GET | `/v1/config?platform&version&lang&region` | 客户端获取配置 |

版本读取规则：

- 读取指定配置版本时，优先读取 `<platform>/vN.star`
- 如果读取 `v1` 且目录文件不存在，则 fallback 到 legacy `<platform>.star`
- 保存 `vN` 时必须写入目标版本文件；不会 fallback 写回 legacy 脚本
- `/v1/config` 的 `version` 参数是客户端 App 版本；当该值形如 `vN` 且存在 `<platform>/vN.star` 时，可用于调试读取指定配置版本，否则兼容读取 `<platform>.star`

### 前端页面设计

```
┌─────────────────────────────────────────────────────────────┐
│  Kimi Config Platform       [New Draft] [Save] [Publish]    │
├─────────┬──────────────────────────┬────────────────────────┤
│         │ ios/v3.star   Version ▼   │ PREVIEW │ HISTORY │DIFF│
│ Platforms│   Monaco Editor          │                        │
│  ios     │   (Starlark 语法高亮)     │  Platform: ios         │
│  android │                          │  Version: 2.5.5        │
│          │   def build_config(ctx): │  Language: zh          │
│          │       ...                │  Region: domestic      │
│          │                          │  [Run Preview]         │
│          │                          │                        │
│          │                          │  {                     │
│          │                          │    "system": {         │
│          │                          │      "greeting": "你好" │
│          │                          │    }                   │
│          │                          │  }                     │
└─────────┴──────────────────────────┴────────────────────────┘
```

前端交互规则：

- latest 配置版本显示 `Read-only`，不可直接编辑，Save 禁用
- latest 可 New Draft / Publish
- 旧版本只读，只能查看 Preview / History / Diff
- 只有 latest 可以点击 **New Draft**
- draft 可编辑，可 Save / Publish；draft 不允许再次 New Draft
- Save draft 后，draft 成为新的 latest
- Diff 按配置版本链比较：`v1` 无 diff，`v2` 对比 `v1`，`v3` 对比 `v2`，draft `vN` 对比其基线 `vN-1`

---

## 核心实现

### 1. Starlark 执行引擎

`go.starlark.net/starlark` 执行用户脚本，关键是把 Go 的 `EvalContext` 转成 Starlark 的 **struct**（支持点符号访问）：

```go
ctxStruct := starlarkstruct.FromStringDict(
    starlark.String("ctx"),
    starlark.StringDict{
        "platform": starlark.String(ctx.Platform),
        "version":  starlark.String(ctx.Version),
        "language": starlark.String(ctx.Language),
        "region":   starlark.String(ctx.Region),
    },
)

// 调用 build_config(ctx)
result, err := starlark.Call(thread, fn, starlark.Tuple{ctxStruct}, nil)
```

用 `starlarkstruct` 而不是 `dict`，这样脚本里才能写 `ctx.language`（点符号），而不是 `ctx["language"]`。

### 2. 结果转 JSON

递归遍历 Starlark 返回值（dict/list/string/bool/int），转成 Go 的 `map[string]interface{}`，然后 JSON 序列化返回。

### 3. Git 仓库管理

```go
// 启动时自动 init Git 仓库
git.PlainInit(path, false)

// 发布 = git add + git commit
wt.Add(".")
wt.Commit(message, &git.CommitOptions{Author: ...})
```

---

## Demo 演示

### 场景 1：中文 + 2.5.5 版本

**线上**:
```bash
curl 'https://kimi-app-config.vercel.app/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

**本地**:
```bash
curl 'http://localhost:8080/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

返回：

```json
{
  "system": {
    "greeting": "你好",
    "cacheLimitMB": 100,
    "kimiplusPlaza": {"enabled": true},
    "webviewRedirectAllowSchemes": ["kimi", "https"]
  },
  "taskbar": {
    "items": [
      {"title": "深度研究", "type": "DEEP_RESEARCH"},
      {"title": "拍照解题", "type": "PHOTO_SOLVER"},
      {"title": "Kimi+", "type": "KIMI_PLUS"}
    ]
  },
  "model": {
    "defaultModel": "kimi-k2"
  }
}
```

### 场景 2：英文 + 2.5.5 版本

```bash
curl 'http://localhost:8080/v1/config?platform=ios&version=2.5.5&lang=en&region=domestic'
```

返回：

```json
{
  "system": {
    "greeting": "Hello",
    "cacheLimitMB": 100,
    "kimiplusPlaza": {"enabled": true},
    "webviewRedirectAllowSchemes": ["kimi", "https"]
  },
  "taskbar": {
    "items": [
      {"title": "Deep Research", "type": "DEEP_RESEARCH"},
      {"title": "Photo Solver", "type": "PHOTO_SOLVER"},
      {"title": "Kimi+", "type": "KIMI_PLUS"}
    ]
  },
  "model": {
    "defaultModel": "kimi-k2"
  }
}
```

### 场景 3：不支持的语言直接报错

```bash
curl 'http://localhost:8080/v1/config?platform=ios&version=2.5.5&lang=fr&region=domestic'
```

返回：

```json
{"error": "build_config error: fail: unsupported language: fr"}
```

### 场景 4：管理后台实时预览

在 Web 界面中：
1. 默认打开 iOS latest 配置版本，编辑器为只读
2. 点击 **New Draft**，系统从 latest 复制内容并创建下一配置版本草稿
3. 在 draft 中编辑 Starlark 脚本
4. 右侧 Preview 面板输入客户端 App `version=2.5.5, language=zh`
5. 点击 **Run Preview**，实时看到 JSON 输出
6. 切换到 **Diff** Tab，查看 draft 和上一配置版本的差异
7. 点击 **Save**，draft 成为新的 latest
8. 点击 **Publish**，Git commit 记录生成
9. 切换到 **History** Tab 查看 commit 历史

---

## 新旧架构对比

| 维度 | 旧架构 | 新架构 |
|------|--------|--------|
| 配置方式 | 每套维度组合配一份完整配置 | 一份 Starlark 脚本，条件分支覆盖 |
| 修改成本 | 改 N 处（维度组合数） | 改 1 处 |
| 版本管理 | 无 | Git 天然支持 |
| 回滚 | 手动 | `git revert` |
| 预览能力 | 无 | 实时执行预览 |
| 配置演进 | 覆盖已有记录 | latest 只读，New Draft 派生下一配置版本 |
| Diff | 无 | 按配置版本链和上一版比较 |
| 未知条件处理 | 静默返回默认配置 | `fail` 显式报错 |
| 安全 | 无限制 | Starlark 沙箱（禁止 I/O、网络）|

---

## 访问方式

### 线上环境

**管理后台**: https://kimi-app-config.vercel.app

**客户端获取配置**:
```bash
curl 'https://kimi-app-config.vercel.app/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

### 本地开发

**方式一：独立启动（Gin + go-git）**

```bash
cd /Users/yuqiang/Desktop/Work/kimi-app-config

# 终端 1：启动后端
cd kimi-config-server && go run ./cmd/server
# 端口 8080

# 终端 2：启动前端
cd kimi-config-web && npm run dev
# 端口 5173，自动代理 /api 到 8080
```

打开 http://localhost:5173

**方式二：Vercel 本地预览**

```bash
# 安装 Vercel CLI
npm install -g vercel

# 登录
vercel login

# 拉取环境变量
vercel env pull

# 本地预览（同时启动前端 + Serverless Functions）
vercel dev
```

---

## 待完善

1. **语义化版本比较**：当前 `ctx.version >= 
