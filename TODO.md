# TODO

## High Priority

- [x] Remove debug `fmt.Printf("HERE %+v", msg)` left in production code (`internal/tui/choice.go:70`)
- [x] Fix inverted spinner logic — spinner never displays for normal users because `if !shared.IsDebugMode() { return }` skips it outside debug mode (`internal/tui/spinner.go:76`)
- [x] Fix `CallLLM` ignoring the passed `ctx` parameter — it discards it and uses `context.Background()`, breaking context propagation/cancellation (`internal/provider/provider.go:91`)

## Medium Priority

- [x] Fix typo in function names: `SprinnerStart` / `SprinnerStop` → `SpinnerStart` / `SpinnerStop` (`internal/tui/spinner.go:73-74, 114-115` + all call sites in `cmd/root.go`)
- [x] Fix typo in struct name: `TextAreModel` → `TextAreaModel` (`internal/tui/textarea.go:20` + all references)
- [x] Fix grammar in commit prompt: `"No more change 240 characters"` → `"No more than 240 characters"` (`internal/provider/commit.prompt:33`)

## Low Priority

- [x] Remove unused `metrics` package — not imported anywhere (`internal/metrics/`)
- [x] Wrap error in `GitGetLatestTags` with error catalog for consistency with other git functions (`internal/git/git.go:141-142`)
- [x] Replace unsafe unchecked type assertions in `CommitMessageTextArea` with checked assertions (`internal/tui/textarea.go:130-131`)
