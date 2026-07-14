# Final responsive layout fixes

## RED

Command:

```text
env GOCACHE=/tmp/flux-go-cache go test ./... -run 'TestConstrainedLayout|TestTinyTerminalLayout' -count=1
```

The constrained valid-height test passed, confirming the existing minimum-body/footer priority. The tiny-terminal table failed for every screen height from 0 through 7: returned body/footer/fixed allocations exceeded `WindowHeight` (for example, component height 6 versus `WindowHeight` 3 at screen height 5).

## GREEN

The allocator now reduces an oversized preferred body to the three-row minimum, trims the wrapped footer next, and reduces the body below three only when the terminal cannot fit fixed border/search rows plus the documented minimum. Sizes remain non-negative and banners are hidden below eight rows.

Focused verification:

```text
env GOCACHE=/tmp/flux-go-cache go test ./... -run 'Layout' -count=1
ok github.com/WariKoda/Flux 0.003s
```

Full verification:

```text
env GOCACHE=/tmp/flux-go-cache go test ./... -count=1
ok github.com/WariKoda/Flux 0.005s

env GOCACHE=/tmp/flux-go-cache go vet ./...
exit 0

env GOCACHE=/tmp/flux-go-cache go build -o /tmp/flux-responsive-banner .
exit 0

gofmt -l .
no output

git diff --check
exit 0
```

The build emitted a read-only warning while attempting to update Go's module stat cache, but exited successfully and wrote `/tmp/flux-responsive-banner`.
