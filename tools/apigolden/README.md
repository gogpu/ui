# apigolden

Signature-level golden of this module's public API, one file per build context. It exists so that the v1.0
shape-freeze work (`ECOSYSTEM_READINESS.md`, the lead's ledger) can prove what ui's frozen surface is: a
regeneration that differs from the baseline is a reviewed delta, not a surprise at v1.0.

Adapted from wgpu's `tools/apigolden` (the wgpu v0.35.0 window tooling, `d921c9e`), which is the reference
implementation of the ecosystem's checklist. The output format is deliberately identical so both repos' goldens are
reviewed the same way. The loader is stdlib-only (see the module note); wgpu's original imports `x/tools`, which the
ecosystem replaced after gogpu found its export-data reader lags the toolchain.

## What it records

* **Packages**: every importable package of `github.com/gogpu/ui` — 58 packages. Excluded by first path segment:
  `internal/**` (implementation), `examples/**` (apps), `tmp/**` (scratch), `tools/**` (this module).
* **Per exported identifier**, sorted and stable:
  * constants — `const Name Type = exact-value`;
  * variables — `var Name Type`;
  * types — `type Name <underlying>`; for a struct every exported field with its type **in declaration order**,
    embedded fields as the type alone; for an interface every method signature and every embedded type; an alias is
    printed as `type Name = Rhs   # alias`;
  * functions — the full signature;
  * exported methods declared on exported receivers — the full signature.
* Unexported identifiers, unexported fields and methods, doc comments and struct tags are **not** recorded.
* The header's `identifiers` count counts **exported package-scope names only** (2043); exported methods are printed
  on their receiver's block but not counted.

This is exactly what the freeze ledger's rows 1 and 3 need: signatures, struct fields and method sets are the
shapes that would cost a v2 to change.

## Commands

Run from the repository root; `GOWORK=off` is required because the developer workspace above the repository does not
contain the `tools` module. The generator loads the analysed module with `GOWORK=off` as well, so a golden describes
the module and never a local workspace.

```sh
# windows/amd64 (the CI runner's context)
GOWORK=off go -C tools run ./apigolden -out apigolden/testdata/api_windows.txt

# linux/amd64
GOWORK=off go -C tools run ./apigolden -goos linux -out apigolden/testdata/api_linux.txt

# darwin/arm64 (macos-latest is arm64)
GOWORK=off go -C tools run ./apigolden -goos darwin -goarch arm64 -out apigolden/testdata/api_darwin.txt
```

These are the three contexts ui ships (`.github/workflows/ci.yml`: ubuntu-latest, macos-latest, windows-latest on
Go 1.25). ui has no wasm and no tagged build.

## Determinism

Two runs on the same tree are byte-identical: the output has no timestamp, no absolute path, no Go version and no
map iteration order (packages sorted by path, identifiers by name, methods by name; struct fields keep declaration
order). Check it with:

```sh
GOWORK=off go -C tools run ./apigolden -out /tmp/g1.txt && cmp apigolden/testdata/api_windows.txt /tmp/g1.txt && echo identical
```

## Toolchain

The header records the toolchain the tool binary was built with (`# toolchain: go1.27.4`, from
`runtime.Version()`) — the same toolchain that runs `go list -export` for the load. **A Go upgrade is a
golden-refresh event:** regenerate, review the delta, and commit the regenerated baselines in the same change.
Without the line, a toolchain-induced rendering change would show up as an unexplained API delta or be
re-baselined silently (practice taken from gg's golden tool).

## Floors

A partial or empty load must fail loudly instead of writing a short golden that still looks valid. Each supported
context has its own floor, and a context that is not listed is refused:

| Context | Packages | Identifiers |
|---|---:|---:|
| `windows/amd64` | 58 | 2043 |
| `linux/amd64` | 58 | 2043 |
| `darwin/arm64` | 58 | 2043 |

`-min-packages` and `-min-identifiers` override the context floor when **both** are passed (a half-set override is
refused). They exist for regeneration after a deliberate API change of surface size.

## Baselines

First generation: 2026-11-19 on `195a0fa` (the tree of the PR that added this tool). The numbers were pinned with the
same binary that produced the goldens.

| File | Context | Packages | Identifiers |
|---|---|---:|---:|
| `testdata/api_windows.txt` | windows/amd64 | 58 | 2043 |
| `testdata/api_linux.txt` | linux/amd64 | 58 | 2043 |
| `testdata/api_darwin.txt` | darwin/arm64 | 58 | 2043 |

The three files differ **only in the context header line**: ui's public surface does not vary by GOOS on the
shipped contexts. That is itself a recorded fact of the freeze.

## Honest limits

* **Signatures, not runtime identity.** The golden cannot prove that exported error sentinels (`plugin.Err*`,
  `registry.Err*` — 13 in total) keep their `errors.Is` identity. That is a separate evidence row: the behaviour
  tests assert it (`plugin/manager_test.go`, `plugin/plugin_test.go`, `registry/widget_test.go`).
* **Struct tags are not recorded.** A tag-only change does not move a baseline.
* **Three contexts only**: windows/amd64, linux/amd64, darwin/arm64. A windows/386 build or an OS-specific file
  outside these contexts would not move a baseline.
* **It does not compare.** Diffing a fresh run against the baselines is the review step; the tool writes a file.

## Tests

```sh
GOWORK=off go -C tools test ./apigolden -count=1
```

The formatter tests build packages in memory (`go/types`), so they need no module load and no network.

## Module note

`tools/` is its own module with **no external dependencies**. The loader uses `go list -export` and `go/importer`
from the running toolchain, so the golden cannot drift out of step with the Go version that invokes it
(`golang.org/x/tools` bundles its own export-data reader, which lags a new Go release — on a cold build cache it
fails outright, exactly when this gate has no other proof). The analysed packages are type-checked from source, so
parameter names survive into the golden; their dependencies come from export data. The stdlib loader is adapted from
gogpu's fix (gogpu PR #575, commit `a0d90ca`; first port: gogpu PR #574). The main module's `go.mod`/`go.sum` are
untouched, and CI's root-level `./...` does not build this module — the commands above are the way to run it.
