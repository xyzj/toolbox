---
description: "Use when: editing Go source files in this repository"
applyTo: "**/*.go"
---
# Go File Instructions

## Editing style
- Keep functions focused and avoid unnecessary restructuring.
- Keep naming consistent with nearby code.
- Add comments only for non-obvious logic.

## Error handling
- Check and return errors promptly.
- Preserve original error semantics unless change is requested.

## Performance and allocation
- Avoid obvious extra allocations in hot paths.
- Reuse existing patterns for slices/maps where practical.

## Concurrency
- Do not introduce goroutines, channels, or locks unless required.
- If touching concurrent code, preserve thread-safety assumptions.

## Validation
- Ensure changed Go files remain free of immediate diagnostics.
- Prefer targeted verification over expensive global checks.
