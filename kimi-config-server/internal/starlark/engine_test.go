package starlark

import (
	"testing"
)

func TestExecute(t *testing.T) {
	script := `
def build_config(ctx):
    config = {
        "system": {
            "greeting": "Hello",
            "enabled": True,
        },
        "version": "1.0",
    }

    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"

    if ctx.version >= "2.5.5":
        config["system"]["newFeature"] = True

    return config
`

	// Test with English, old version
	result, err := Execute(script, EvalContext{
		Platform: "ios",
		Version:  "2.5.0",
		Language: "en",
		Region:   "domestic",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	sys, ok := result["system"].(map[string]interface{})
	if !ok {
		t.Fatal("expected system to be a map")
	}

	if sys["greeting"] != "Hello" {
		t.Errorf("expected greeting=Hello, got %v", sys["greeting"])
	}

	if _, hasNewFeature := sys["newFeature"]; hasNewFeature {
		t.Error("expected no newFeature for version 2.5.0")
	}

	// Test with Chinese, new version
	result2, err := Execute(script, EvalContext{
		Platform: "ios",
		Version:  "2.5.5",
		Language: "zh",
		Region:   "domestic",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	sys2, ok := result2["system"].(map[string]interface{})
	if !ok {
		t.Fatal("expected system to be a map")
	}

	if sys2["greeting"] != "你好" {
		t.Errorf("expected greeting=你好, got %v", sys2["greeting"])
	}

	if sys2["newFeature"] != true {
		t.Errorf("expected newFeature=True, got %v", sys2["newFeature"])
	}
}

func TestExecuteMissingFunction(t *testing.T) {
	script := `
x = 1
`
	_, err := Execute(script, EvalContext{})
	if err == nil {
		t.Fatal("expected error for missing build_config")
	}
}
