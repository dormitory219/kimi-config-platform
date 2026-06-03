# iOS Git Convention

> 供开发者和 agent 快速查阅的结构化版本。内容整理自 `iOS Git Convention.pdf`；如有歧义，以团队约定和原 PDF 为准。

## 1. 分支模型

- 以 `main` 为主干分支开展日常工作。
- 常规迭代需求从 `main` 拉出 `feature/*` 分支开发，完成后合回 `main`。
- 对所有正在开展需求都通用的改动，可以从 `main` 拉分支处理，再合回 `main`。
- 开发中的功能分支在合并前，应持续通过 `rebase main` 获取最新代码。
- 发版时从 `main` 合入 `release/*` 分支，在 `release/*` 上进行 QA、bugfix 和发版。
- 发布完成后，在 `release/*` 上打 tag 归档，再合回 `main`。
- 紧急线上修复从最近线上版本对应的 `release/*` 拉出 `hotfix` 分支，修复并发布后合回 `release/*`，再合回 `main`。

## 2. Commit Message

- 原则上遵循 Conventional Commits。
- 推荐格式：`type(scope): message`
- `type` 常用：
  - `feat`: 新功能
  - `fix`: 缺陷修复
  - `chore`: 杂项、资源或配置更新
  - `refactor`: 重构
  - `style`: 代码格式调整
  - `test`: 测试相关
- `scope` 必填，并尽量与团队在同类代码中已有的写法保持一致。
- `message` 使用英文，客观描述改动内容，不写心路历程。
- 尽量保持简洁统一：
  - 首字母不要随意大写
  - 末尾不要加句号
- 一个 commit 只做一件事。
- 如果只是反复修改同一逻辑，尽量在未 push 前使用 `git commit --amend`、`fixup` 或 `squash` 整理历史。

## 3. Merge Request

- 功能分支初具规模时就可以先发 Draft MR，方便团队尽早了解和 review。
- MR 由作者负责推进和最终 merge，不默认 reviewer approve 后就会代为 merge。
- MR 至少经过一位 reviewer review 后才能 merge。
- 谁着急谁主动 push 进度、催 review，不默认“提了 MR 就会有人来 review”。
- MR 作者需要认真处理 review comment，给出修改或解释。

## 4. 推荐实践

- 日常同步主干优先使用 `rebase`。
- 例如：
  - `git pull --rebase origin main`
  - 本地 `main` 最新时直接 `git rebase main`
- 如果 rebase 负担已经很重，通常说明分支分叉时间过长，可以改为先 `merge main`。
- 一旦某个分支已经 merge 过，就不要继续对这段共享历史做 rebase，应改用 merge。
- 提交 MR 前整理自己的 commit 历史：
  - 可以用 `git rebase -i` 做 `fixup` / `squash`
  - 不建议把特别大的需求全部 squash 成一个 commit
- 本地尽量使用 fast-forward merge，GitLab 上的 MR 保持 merge commit。
