# Config Platform Implementation Plan

> **Goal:** Build a Git + Starlark based dynamic config platform with React frontend and Go backend.

**Architecture:** Go HTTP server with embedded Starlark engine serves both admin UI API and client config API. Config scripts stored in Git repo. React SPA with Monaco Editor for script editing and preview panel.

**Tech Stack:** Go 1.24 + starlark-go + go-git, React 19 + Vite + Monaco Editor

---

## Phase 1: Go Backend

### Task 1: Initialize Go Project

**Files:**
- Create: `kimi-config-server/go.mod`
- Create: `kimi-config-server/cmd/server/main.go`
- Create: `kimi-config-server/Makefile`

- [ ] **Step 1: Create go.mod**

```bash
cd /Users/yuqiang/Desktop/Work/kimi-app-scan-qrcode
mkdir -p kimi-config-server
cd kimi-config-server
go mod init github.com/moonshot-ai/kimi-config-server
```

- [ ] **Step 2: Install dependencies**

```bash
cd kimi-config-server
go get go.starlark.net/starlark github.com/go-git/go-git/v5 github.com/gin-gonic/gin
go mod tidy
```

- [ ] **Step 3: Create main.go skeleton**

```go
package main

import "fmt"

func main() {
    fmt.Println("kimi-config-server")
}
```

- [ ] **Step 4: Verify build**

```bash
cd kimi-config-server
go build ./cmd/server
```

---

### Task 2: Starlark Execution Engine

**Files:**
- Create: `kimi-config-server/internal/starlark/engine.go`
- Create: `kimi-config-server/internal/starlark/engine_test.go`

- [ ] **Step 1: Write engine.go**

Implements:
- `Execute(script string, ctx EvalContext) (map[string]interface{}, error)`
- Thread with timeout (1s) and memory limit (10MB)
- Injects `ctx` as Starlark dict with version, language, region, platform
- Calls `build_config(ctx)` and converts return value to Go map

- [ ] **Step 2: Write test**

Test with sample Starlark script that returns config based on ctx.version.

- [ ] **Step 3: Run test**

```bash
cd kimi-config-server
go test ./internal/starlark/...
```

---

### Task 3: Git Repository Manager

**Files:**
- Create: `kimi-config-server/internal/git/repo.go`

- [ ] **Step 1: Implement Repo struct**

Methods:
- `Open(path string)` - opens Git repo
- `ReadFile(path string) ([]byte, error)` - read file from worktree
- `WriteFile(path string, content []byte) error` - write file
- `Commit(message, author string) error` - git add + commit
- `History(path string, n int) ([]Commit, error)` - git log
- `Pull() error` - git pull origin main

- [ ] **Step 2: Test with temp repo**

```bash
go test ./internal/git/...
```

---

### Task 4: Admin API Handlers

**Files:**
- Create: `kimi-config-server/internal/api/admin.go`
- Create: `kimi-config-server/internal/api/preview.go`

- [ ] **Step 1: Implement admin handlers**

Routes:
- `GET /api/platforms` -> list .star files in repo
- `GET /api/scripts/:platform` -> read file content
- `POST /api/scripts/:platform` -> write file (draft)
- `POST /api/scripts/:platform/publish` -> write file + git commit

- [ ] **Step 2: Implement preview handler**

- `POST /api/preview`
- Body: `{ "script": "...", "ctx": {"version":"2.5.5","language":"zh","region":"domestic","platform":"ios"} }`
- Executes script with Starlark engine, returns JSON config

---

### Task 5: Client Config API

**Files:**
- Create: `kimi-config-server/internal/api/client.go`

- [ ] **Step 1: Implement config handler**

- `GET /v1/config?platform=ios&version=2.5.5&lang=zh&region=domestic`
- Reads platform.star from repo
- Executes with ctx from query params
- Returns JSON config

---

### Task 6: Server Setup and CORS

**Files:**
- Modify: `kimi-config-server/cmd/server/main.go`

- [ ] **Step 1: Wire everything together**

Gin router with CORS for local development. Mount all API routes.

- [ ] **Step 2: Add init logic**

On startup:
- Check if config repo exists at `./config-repo`, if not init it
- Create sample `ios.star` and `android.star` if they don't exist

- [ ] **Step 3: Test server**

```bash
go run ./cmd/server
# In another terminal:
curl http://localhost:8080/api/platforms
```

---

## Phase 2: React Frontend

### Task 7: Initialize React Project

**Files:**
- Create: `kimi-config-web/package.json`
- Create: `kimi-config-web/vite.config.ts`
- Create: `kimi-config-web/index.html`
- Create: `kimi-config-web/src/main.tsx`
- Create: `kimi-config-web/src/App.tsx`

- [ ] **Step 1: Create project with Vite**

```bash
cd /Users/yuqiang/Desktop/Work/kimi-app-scan-qrcode
npm create vite@latest kimi-config-web -- --template react-ts
cd kimi-config-web
npm install
```

- [ ] **Step 2: Install dependencies**

```bash
cd kimi-config-web
npm install @monaco-editor/react monaco-editor axios
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

---

### Task 8: API Client

**Files:**
- Create: `kimi-config-web/src/api.ts`

- [ ] **Step 1: Implement API functions**

```typescript
export interface ScriptContext {
  version: string;
  language: string;
  region: string;
  platform: string;
}

export const api = {
  getPlatforms: () => axios.get('/api/platforms'),
  getScript: (platform: string) => axios.get(`/api/scripts/${platform}`),
  saveScript: (platform: string, content: string) => axios.post(`/api/scripts/${platform}`, { content }),
  publishScript: (platform: string) => axios.post(`/api/scripts/${platform}/publish`),
  preview: (script: string, ctx: ScriptContext) => axios.post('/api/preview', { script, ctx }),
};
```

---

### Task 9: Editor Component

**Files:**
- Create: `kimi-config-web/src/components/Editor.tsx`

- [ ] **Step 1: Monaco Editor wrapper**

Props: `value`, `onChange`, `platform`
- Monaco Editor with Python syntax highlighting (Starlark is Python-like)
- Basic editor config (font size, minimap, etc.)

---

### Task 10: Preview Component

**Files:**
- Create: `kimi-config-web/src/components/Preview.tsx`

- [ ] **Step 1: Preview panel**

- Form inputs for version, language, region
- "Run Preview" button
- JSON output display (syntax highlighted)
- Error display if script fails

---

### Task 11: Sidebar Component

**Files:**
- Create: `kimi-config-web/src/components/Sidebar.tsx`

- [ ] **Step 1: Platform selector**

- List platforms from API
- Click to switch editor content
- Show published status

---

### Task 12: App Layout Integration

**Files:**
- Modify: `kimi-config-web/src/App.tsx`

- [ ] **Step 1: Three-panel layout**

```
+----------+-------------------+----------+
| Sidebar  | Editor            | Preview  |
|          |                   |          |
+----------+-------------------+----------+
```

- State: current platform, script content, preview result
- Auto-preview on ctx change (debounced)
- Save/Publish buttons in header

---

### Task 13: Dev Proxy and Test

**Files:**
- Modify: `kimi-config-web/vite.config.ts`

- [ ] **Step 1: Add proxy**

```typescript
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/v1': 'http://localhost:8080',
  }
}
```

- [ ] **Step 2: Run both dev servers**

Terminal 1: `cd kimi-config-server && go run ./cmd/server`
Terminal 2: `cd kimi-config-web && npm run dev`

- [ ] **Step 3: End-to-end test**

1. Open http://localhost:5173
2. Edit ios.star script
3. Change preview ctx, verify output
4. Click Save, verify file written
5. Click Publish, verify git commit

---

## Phase 3: Polish

### Task 14: Sample Config Scripts

**Files:**
- Create: `kimi-config-server/config-repo/ios.star`
- Create: `kimi-config-server/config-repo/android.star`

- [ ] **Step 1: Write sample scripts**

Sample iOS config with system, taskbar, model, upgrade sections.

### Task 15: README

- [ ] **Step 1: Write README with setup and usage instructions**
