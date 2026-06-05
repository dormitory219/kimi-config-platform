def build_config(ctx):
    # 创建空结构，不填任何默认值
    config = {
        "system": {},
        "taskbar": {"items": []},
        "model": {},
    }

    # === 语言配置 ===
    if ctx.language == "zh":
        config["taskbar"]["items"].append({"type": "DEEP_RESEARCH", "title": "深度研究"})
        config["taskbar"]["items"].append({"type": "PHOTO_SOLVER", "title": "拍照解题"})
    elif ctx.language == "en":
        config["system"]["greeting"] = "Hello"
        config["taskbar"]["items"].append({"type": "DEEP_RESEARCH", "title": "Deep Research"})
        config["taskbar"]["items"].append({"type": "PHOTO_SOLVER", "title": "Photo Solver"})
    elif ctx.language == "ja":
        config["system"]["greeting"] = "こんにちは"
        config["taskbar"]["items"].append({"type": "DEEP_RESEARCH", "title": "深度研究"})
        config["taskbar"]["items"].append({"type": "PHOTO_SOLVER", "title": "拍照解题"})
    else:
        fail("unsupported language: " + ctx.language)


    # === 地区配置 ===
    if ctx.region == "domestic":
        config["system"]["cacheLimitMB"] = 100
        config["system"]["webviewRedirectAllowSchemes"] = ["kimi", "https"]
    elif ctx.region == "overseas":
        config["system"]["cacheLimitMB"] = 100
        config["system"]["webviewRedirectAllowSchemes"] = ["kimi", "https"]
        config["system"]["domains"] = {"apiDomain": "api.kimi.moonshot.cn"}
    else:
        fail("unsupported region: " + ctx.region)

    # === 版本配置 ===
    if ctx.version >= "2.5.5":
        config["system"]["kimiplusPlaza"] = {"enabled": True}
        config["taskbar"]["items"].append({"type": "KIMI_PLUS", "title": "Kimi+"})

    if ctx.version >= "2.6.0":
        config["model"]["defaultModel"] = "kimi-k2.5"
    elif ctx.version >= "2.5.5":
        config["model"]["defaultModel"] = "kimi-k2"
    else:
        fail("unsupported version: " + ctx.version)

    return config
