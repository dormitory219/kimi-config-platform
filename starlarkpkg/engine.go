package starlarkpkg

import (
	"context"
	"fmt"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// EvalContext holds the context values passed to Starlark build_config function
type EvalContext struct {
	Platform string `json:"platform"`
	Version  string `json:"version"`
	Language string `json:"language"`
	Region   string `json:"region"`
}

// Execute runs a Starlark script and returns the config map
func Execute(script string, ctx EvalContext) (map[string]interface{}, error) {
	// Create thread with timeout
	thread := &starlark.Thread{
		Name: "config-eval",
		Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			return nil, fmt.Errorf("load not supported in MVP")
		},
	}

	// Set execution deadline
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	thread.SetLocal("deadline", ctxTimeout)

	// Predeclared builtins - restrict to safe ones
	predeclared := starlark.StringDict{
		"True":  starlark.True,
		"False": starlark.False,
		"None":  starlark.None,
	}

	// Execute the script to get globals
	globals, err := starlark.ExecFile(thread, "config.star", script, predeclared)
	if err != nil {
		return nil, fmt.Errorf("exec error: %w", err)
	}

	// Find build_config function
	buildConfig, ok := globals["build_config"]
	if !ok {
		return nil, fmt.Errorf("build_config function not found")
	}

	fn, ok := buildConfig.(*starlark.Function)
	if !ok {
		return nil, fmt.Errorf("build_config is not a function")
	}

	// Build ctx as a struct so ctx.language works (dot notation)
	ctxStruct := starlarkstruct.FromStringDict(
		starlark.String("ctx"),
		starlark.StringDict{
			"platform": starlark.String(ctx.Platform),
			"version":  starlark.String(ctx.Version),
			"language": starlark.String(ctx.Language),
			"region":   starlark.String(ctx.Region),
		},
	)

	// Call build_config(ctx)
	result, err := starlark.Call(thread, fn, starlark.Tuple{ctxStruct}, nil)
	if err != nil {
		return nil, fmt.Errorf("build_config error: %w", err)
	}

	// Convert result to Go map
	config, err := starlarkToGo(result)
	if err != nil {
		return nil, fmt.Errorf("convert error: %w", err)
	}

	m, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("build_config must return a dict")
	}

	return m, nil
}

// starlarkToGo converts a Starlark value to a Go interface{}
func starlarkToGo(v starlark.Value) (interface{}, error) {
	switch val := v.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(val), nil
	case starlark.Int:
		i, ok := val.Int64()
		if !ok {
			// Try as string if too big
			return val.BigInt().String(), nil
		}
		return i, nil
	case starlark.Float:
		return float64(val), nil
	case starlark.String:
		return string(val), nil
	case *starlark.Dict:
		m := make(map[string]interface{})
		for _, item := range val.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("dict key must be string, got %T", item[0])
			}
			value, err := starlarkToGo(item[1])
			if err != nil {
				return nil, err
			}
			m[string(key)] = value
		}
		return m, nil
	case *starlark.List:
		arr := make([]interface{}, 0, val.Len())
		iter := val.Iterate()
		defer iter.Done()
		var x starlark.Value
		for iter.Next(&x) {
			item, err := starlarkToGo(x)
			if err != nil {
				return nil, err
			}
			arr = append(arr, item)
		}
		return arr, nil
	default:
		return nil, fmt.Errorf("unsupported starlark type: %T", v)
	}
}
