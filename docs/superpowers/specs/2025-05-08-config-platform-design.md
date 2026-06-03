# Kimi Config Platform 设计文档

> 基于 Git + Starlark 的动态配置系统

## 目标

重构现有动态配置平台，解决分版本/语言/地区配置时的"维度组合爆炸"问题。开发者通过 Python-like 的 Starlark 语法编写配置规则，后端按客户端请求的维度实时生成配置。

## 架构

```
React 前端 (Monaco Editor) <-> Go HTTP API (Starlark 引擎) <-> Git 仓库 (*.star)
                                        |
                                        v
                                    App 客户端
```

## 技术栈

- **前端**: React + Monaco Editor + TypeScript
- **后端**: Go + google/starlark-go + git2go/go-git
- **配置语言**: Starlark (Python 子集)
- **存储**: Git 仓库（文件系统）

## Starlark DSL

```starlark
def build_config(ctx):
    # ctx.version, ctx.language, ctx.region, ctx.platform
    config = {
        "system": {"greeting": "Hello"},
        "taskbar": {...},
    }

    if ctx.version >= "2.5.5":
        config["system"]["kimiplusPlaza"] = {"enabled": True}

    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"

    return config
```

## API

### 管理后台
- `GET /api/platforms` - 列出平台
- `GET /api/scripts/:platform` - 读取脚本
- `POST /api/scripts/:platform` - 保存脚本
- `POST /api/scripts/:platform/publish` - 发布（git commit）
- `POST /api/preview` - 预览配置

### 客户端
- `GET /v1/config?platform&version&lang&region` - 获取配置

## Git 仓库结构

```
config-repo/
├── ios.star
├── android.star
└── lib/
    └── common.star
```

## 安全

- Starlark 沙箱：禁止文件 I/O、网络、import
- 执行超时 1s，内存限制 10MB
