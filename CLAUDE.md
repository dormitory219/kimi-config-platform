# Kimi Config Platform

动态配置管理平台，基于 Git + Starlark 脚本引擎。

**线上地址**: https://kimi-app-config.vercel.app

## 项目结构

```
kimi-app-config/
├── api/
│   └── index.go            # Vercel Serverless Function 入口
├── githubpkg/              # GitHub API 客户端（替代本地 git）
├── starlarkpkg/            # Starlark 执行引擎
├── kimi-config-server/     # 本地开发用的 Go 后端（Gin + go-git）
│   ├── cmd/server/
│   ├── internal/
│   └── config-repo/        # 配置脚本 Git 仓库
│       ├── ios.star
│       └── android.star
├── kimi-config-web/        # React 前端
│   ├── src/
│   │   ├── components/
│   │   ├── api.ts
│   │   └── App.tsx
│   └── package.json
├── vercel.json             # Vercel 部署配置
└── go.mod                  # Go 模块定义
```

## 本地开发

### 方式一：本地启动（Gin + go-git）

```bash
# 启动后端
cd kimi-config-server
go run ./cmd/server
# 默认端口 8080

# 启动前端（新终端）
cd kimi-config-web
npm run dev
# 默认端口 5173，自动代理 /api 到 localhost:8080
```

打开 http://localhost:5173

### 方式二：Vercel 本地预览

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

## 使用

### 编写配置脚本

在 Monaco Editor 中编写 Starlark（Python 语法子集）脚本：

```starlark
def build_config(ctx):
    config = {
        "system": {},
        "taskbar": {"items": []},
    }

    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"
    elif ctx.language == "en":
        config["system"]["greeting"] = "Hello"
    else:
        fail("unsupported language: " + ctx.language)

    if ctx.version >= "2.5.5":
        config["system"]["kimiplusPlaza"] = {"enabled": True}

    return config
```

`ctx` 包含：
- `ctx.platform` — `"ios"` / `"android"`
- `ctx.version` — 版本号字符串（如 `"2.5.5"`）
- `ctx.language` — `"zh"` / `"en"` / `"ja"`
- `ctx.region` — `"domestic"` / `"overseas"`

### 预览配置

在右侧 Preview 面板输入 version / language / region，点击 **Run Preview**，实时查看生成的 JSON 配置。

### 保存与发布

- **Save** — 保存脚本（直接生成 GitHub commit）
- **Publish** — 带消息提交 GitHub commit + 记录历史
- **History** — 查看 Git commit 历史（Tab 切换）

线上和本地的保存行为略有不同：
- **本地（Gin）**: Save 只写工作区文件，Publish 才生成 Git commit
- **线上（Vercel）**: Save 和 Publish 都直接通过 GitHub API 创建 commit（因为 Serverless 无状态，无法保存工作区）

### 客户端获取配置

**线上**:
```bash
curl 'https://kimi-app-config.vercel.app/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

**本地**:
```bash
curl 'http://localhost:8080/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

## 部署架构（线上）

全栈部署在 **Vercel**（免费计划，无需信用卡）：

| 组件 | 部署方式 | 说明 |
|------|----------|------|
| 前端 | Vercel 静态站点 | `kimi-config-web/` 构建为静态文件 |
| 后端 | Vercel Serverless Functions (Go) | `api/index.go` 处理所有 API 请求 |
| 数据存储 | GitHub API | 配置文件通过 GitHub REST API 读写，天然 Git 版本管理 |

### 环境变量（Vercel）

| 变量 | 说明 |
|------|------|
| `GITHUB_TOKEN` | GitHub Personal Access Token (`repo` 权限) |
| `GITHUB_REPO` | 仓库名，如 `dormitory219/kimi-config-platform` |
| `GITHUB_CONFIG_PATH` | 配置文件目录，如 `kimi-config-server/config-repo` |

### 线上使用

**管理后台**: https://kimi-app-config.vercel.app

**客户端获取配置**:
```bash
curl 'https://kimi-app-config.vercel.app/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

**API 端点**:
- `GET  /api/platforms`                列出所有平台
- `GET  /api/scripts/:platform`        读取脚本内容
- `GET  /api/scripts/:platform/history` Git commit 历史
- `POST /api/scripts/:platform`        保存脚本（直接 commit）
- `POST /api/scripts/:platform/publish` 发布脚本（带消息 commit）
- `POST /api/preview`                  预览配置（执行脚本）
- `GET  /v1/config`                    客户端获取配置

## 技术栈

- **线上后端**: Go 1.23 + Vercel Serverless Functions + GitHub API + go-starlark
- **本地后端**: Go 1.23 + Gin + go-starlark + go-git
- **前端**: React 19 + TypeScript + Vite + Monaco Editor
- **配置语言**: Starlark（Python 子集，沙箱安全）

## 数据存储

配置脚本存储在 `config-repo/` 的 Git 仓库中：
- 每次 Publish = 一次 Git commit
- 天然支持版本历史、回滚、diff
- 无需数据库

## 安全限制

Starlark 脚本运行环境：
- 禁止文件 I/O（除 `load`）
- 禁止网络访问
- 禁止 `import`
- 执行超时 1 秒

## 开发

```bash
# 后端测试（本地）
cd kimi-config-server
go test ./internal/starlark/... -v

# 前端开发
cd kimi-config-web
npm run dev

# Vercel 线上部署
vercel deploy --prod
```