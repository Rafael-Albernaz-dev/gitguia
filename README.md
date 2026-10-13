# gitguia

[![Go Version](https://img.shields.io/badge/go-1.24%2B-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![CI](https://github.com/Rafael-Albernaz-dev/gitguia/actions/workflows/ci.yml/badge.svg)](https://github.com/Rafael-Albernaz-dev/gitguia/actions)

**An opinionated, local-first CLI and TUI companion in Go that guides developers and autonomous AI agents through safe Git workflows.**

`gitguia` streamlines the process of spinning up isolated Git Worktrees, inspecting diffs, staging files, drafting commit proposals (via local deterministic rules or local open LLMs), committing changes with hook/GPG preservation, and safely publishing to remotes.

---

## ⚡ Key Highlights

- **Worktree Lifecycle Management:** Spawn isolated workspaces in standardized directories, copying required unversioned files (`.env`, config files) without manual overhead.
- **Unified Core:** The CLI and TUI call the exact same underlying application use cases. Zero behavior discrepancies between automated scripts, agents, and human interactive usage.
- **Native Git Subprocesses:** Relies directly on the system's installed `git` binary using stable structured outputs (`--porcelain -z`). Fully preserves your SSH keys, GPG signing, Git LFS, and pre-commit hooks.
- **Safety by Default:**
  - Non-destructive teardowns: Worktree deletion strictly inspects dirty tracking, untracked changes, and ignored sensitive files (such as `.env`).
  - Separated Publishing: Creating a commit **never** automatically triggers a push. Push is always an explicit action with clear remote/branch targets.
- **Local-First Commit Proposals:** Proposes commit messages based strictly on the staged diff using deterministic heuristics or local open models (e.g. Ollama). No cloud AI dependencies or unconsented telemetry.

---

## 🏗️ Architecture & Philosophy

```
  Human (TUI) / Agent (CLI)
            │
            ▼
    ┌──────────────────────────────┐
    │     Application Core         │ ───► Local Rule / LLM Engine
    │     (internal/app)           │
    └──────────────────────────────┘
            │               │
            ▼               ▼
      Native Git CLI    Filesystem
```

For full architectural blueprints, see [docs/architecture.md](docs/architecture.md).

---

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Rafael-Albernaz-dev/gitguia.git
cd gitguia

# Build the binary
go build -o bin/gitguia ./cmd/gitguia

# Verify installation
./bin/gitguia --version
```

### Basic Commands (CLI)

```bash
# List all active worktrees and their Git statuses
gitguia list

# Create a new isolated worktree
gitguia spawn feat/my-new-task --base main

# Review status and diffs
gitguia status

# Stage files and create a commit with guided message proposal
gitguia commit

# Explicitly push to remote
gitguia push
```

---

## 🗺️ Roadmap & Slices

- [x] **GGUIA-001 — Foundation:** Core architecture, domain models, CLI skeleton, CI pipeline.
- [ ] **GGUIA-002 — Discovery & Status:** Worktree porcelain parser and staged/unstaged diff inspections.
- [ ] **GGUIA-003 — Workspace Provisioning:** `spawn` engine with declared file replication.
- [ ] **GGUIA-004 — Safe Teardown:** Guardrails preventing data loss on uncommitted or ignored files.
- [ ] **GGUIA-005 — Guided Commits:** Per-file staging and rule-based commit proposals.
- [ ] **GGUIA-006 — Local Model Adapter:** Ollama / local LLM integration for contextual commit drafting.
- [ ] **GGUIA-007 — Explicit Push:** Target confirmation and separated publication flow.
- [ ] **GGUIA-008 — Candidate v1:** Full TUI integration (Bubble Tea), end-to-end tests, and v1 release.

---

## 📄 License

MIT © [Rafael Albernaz](https://github.com/Rafael-Albernaz-dev)
