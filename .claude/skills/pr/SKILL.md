---
name: pr
description: Create a GitHub pull request with auto-generated title and description
disable-model-invocation: true
allowed-tools: Bash(gh *), Bash(git *)
---

Create a pull request for the current branch.

## Context

Current branch: !`git rev-parse --abbrev-ref HEAD`
Default branch: !`git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo main`
Commits to include: !`git log --oneline origin/HEAD..HEAD 2>/dev/null || git log --oneline -5`

## Instructions

1. If not on main/master, check if branch has unpushed commits
2. Push the branch if needed: `git push -u origin HEAD`
3. Analyze the commits to generate a PR title and description:
   - Title: concise summary of the change
   - Description: bullet points of what changed and why
4. Create the PR:
   ```
   gh pr create --title "..." --body "..."
   ```
5. Return the PR URL

If `$ARGUMENTS` is provided, use it as the PR title hint.

## PR Description Format

```markdown
## Summary
<2-3 bullet points describing what changed>

## Test plan
<How to verify the changes work>
```
