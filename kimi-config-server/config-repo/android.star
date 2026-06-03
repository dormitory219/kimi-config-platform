def build_config(ctx):
    config = {
        "system": {},
    }

    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"
    elif ctx.language == "en":
        config["system"]["greeting"] = "Hello"
    else:
        fail("unsupported language: " + ctx.language)

    return config
