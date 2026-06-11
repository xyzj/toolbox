# Project Copilot Instructions

These rules apply to all coding tasks in this repository.

## Change policy
- Prefer minimal, targeted edits.
- Do not refactor unrelated code unless explicitly requested.
- Preserve existing behavior unless the task requires a behavior change.

## Safety policy
- Do not run destructive git commands unless explicitly requested.
- Avoid changing build, deployment, or secrets files unless required by the task.

## Go coding policy
- Keep code compile-safe after each change.
- Return early on errors and keep error context clear.
- Keep public APIs stable unless requested.
- For all operations requiring a network connection, the `https_proxy` environment variable must be set to `http://127.0.0.1:8119`.

## Verification policy
- For code changes, run lightweight checks relevant to changed files first.
- If full test/build is not possible, report what was validated and what was not.

## Response policy
- Summarize what changed, why, and any residual risk.
- Include exact file paths touched.

## Git policy
- When using `git add`, explicitly specify file paths and do not add whole directories.
- By default, include only modifications and deletions of tracked files; do not include untracked files.
- If new (untracked) files must be included, add them one by one explicitly and note the reason in the commit message.
- Do not mention changes to `go.mod`, `go.sum`, or `Makefile` in commit messages.
