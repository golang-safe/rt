# pkg.safego.dev/rt — the SafeGo compiler intrinsics

These four packages are the parts of SafeGo the compiler *recognises* rather than transpiles.
A SafeGo program calls `volatile.Reg32(...).Store(x)` and the compiler emits a width-exact
volatile store; it does not compile the Go body you see here.

| Package | What the compiler lowers it to |
| :--- | :--- |
| `volatile` | width-exact volatile loads and stores of memory-mapped registers |
| `time` | the runtime's monotonic clock, tickers and deadlines |
| `sync` | the runtime's mutexes, sized and priority-ceiling-analysed at compile time |
| `errors` | the runtime's code-comparing error values |

So the Go implementations here are not the target behaviour — they are the **host** half of the
same contract. That is the point of them: a driver written against this API is ordinary Go on a
workstation, unit-testable, fuzzable and race-detectable without a board, and the compiler makes
the same API mean the same thing on the target.

Two halves of one contract only hold if they are versioned as one. This module is therefore
pinned by the compiler at an **exact** version, not a range, and is checked out inside the
compiler repository as a submodule. Import it from a project the ordinary way:

```go
import "pkg.safego.dev/rt/volatile"
```

`embed.go` carries the tree into the `safego` binary, so a build can answer "which intrinsic
source is this compiler's contract" from the executable alone, with no network and no module
cache.

The set is closed. A path under `pkg.safego.dev/rt/` that is not one of the four above is
rejected by the compiler (rule R-123) rather than silently treated as an intrinsic.

See [Spec 012](https://github.com/golang-safe/safego/blob/main/spec/012_safego_canonical_intrinsic_packages/spec.md)
for the namespace rules and [Spec 001](https://github.com/golang-safe/safego/blob/main/spec/001_safego_language_and_architecture/spec.md)
for the language.
