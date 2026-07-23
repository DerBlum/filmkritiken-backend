# RTK (Rust Token Killer) Instructions

Always prefix CLI commands with `rtk` (including chained commands with `&&`). RTK proxies and optimizes CLI output for 60-90% token savings.

- **Golden Rule**: Always use `rtk <cmd>` (e.g. `rtk git status`, `rtk go test`, `rtk docker ps`).
- **Chained Commands**: `rtk git add . && rtk git commit -m "msg" && rtk git push`
- **Common Wrappers**:
  - **Git**: `rtk git status`, `rtk git log`, `rtk git diff`, `rtk git add`, `rtk git commit`, `rtk git push`
  - **Go / Build**: `rtk go test`, `rtk make`, `rtk docker ps`, `rtk docker logs`
  - **GitHub**: `rtk gh pr view`, `rtk gh pr checks`, `rtk gh issue list`
  - **Files & Search**: `rtk ls`, `rtk read`, `rtk grep`, `rtk find`
  - **Debug & Meta**: `rtk err <cmd>`, `rtk log <file>`, `rtk summary <cmd>`, `rtk proxy <cmd>` (unfiltered), `rtk gain` (savings)

---

## Agent skills

### Issue tracker

Issues and specs live as markdown files under `.scratch/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Canonical triage roles mapped to repo label strings (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout (`CONTEXT.md` + `docs/adr/` at repo root). See `docs/agents/domain.md`.

### UAT & Testing standards

Acceptance criteria & Gherkin UAT specifications (`@e2e` tags, Given/When/Then). See `docs/agents/uat-standards.md`.

