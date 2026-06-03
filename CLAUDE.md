# Kimi Config Platform

动态配置管理平台，基于 Git + Starlark 脚本引擎。

## 项目结构

```
kimi-app-config/
├── kimi-config-server/     # Go 后端服务
│   ├── cmd/server/         # 服务入口
│   ├── internal/
│   │   ├── api/            # HTTP API handlers
│   │   ├── starlark/       # Starlark 执行引擎
│   │   └── git/            # Git 仓库操作
│   └── config-repo/        # 配置脚本 Git 仓库
│       ├── ios.star
│       └── android.star
└── kimi-config-web/        # React 前端
    ├── src/
    │   ├── components/     # Editor / Preview / History / Sidebar
    │   ├── api.ts          # API 客户端
    │   └── App.tsx         # 主布局
    └── package.json
```

## 启动

### 1. 启动后端（Go）

```bash
cd kimi-config-server
go run ./cmd/server
# 默认端口 8080
```

后端 API：
- `GET  /api/platforms`               列出所有平台
- `GET  /api/scripts/:platform`       读取脚本内容
- `GET  /api/scripts/:platform/history` Git commit 历史
- `POST /api/scripts/:platform`       保存脚本（工作区）
- `POST /api/scripts/:platform/publish` 发布脚本（git commit）
- `POST /api/preview`                 预览配置（执行脚本）
- `GET  /v1/config`                   客户端获取配置

### 2. 启动前端（React）

```bash
cd kimi-config-web
npm run dev
# 默认端口 5173，自动代理 /api 到 localhost:8080
```

打开 http://localhost:5173

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

- **Save** — 保存到工作区文件（不生成 Git commit）
- **Publish** — Git commit + 记录历史
- **History** — 查看 Git commit 历史（Tab 切换）

### 客户端获取配置

```bash
curl 'http://localhost:8080/v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic'
```

## 技术栈

- **后端**: Go 1.23 + Gin + go-starlark + go-git
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
# 后端测试
cd kimi-config-server
go test ./internal/starlark/... -v

# 前端开发
cd kimi-config-web
npm run dev
```