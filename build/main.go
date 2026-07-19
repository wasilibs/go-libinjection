package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/curioswitch/go-build"
	"github.com/google/go-github/github"
	"github.com/goyek/goyek/v3"
	"github.com/goyek/x/boot"
	"github.com/goyek/x/cmd"
)

// libraryRepo is the upstream repository checked for new releases by the update task.
const libraryRepo = "libinjection/libinjection"

func main() {
	tags := modeTags()

	// Register a mode-aware test task before DefineTasks so it is wired into the
	// test and check aggregates. The mode is selected with LIBINJECTION_TEST_MODE
	// (wazero (default), cgo, tinygo) to match the CI matrix.
	build.RegisterTestTask(goyek.Define(goyek.Task{
		Name:  "test-go",
		Usage: "Runs Go tests. LIBINJECTION_TEST_MODE selects the mode (wazero, cgo, tinygo).",
		Action: func(a *goyek.A) {
			if testMode() == modeTinyGo {
				cmd.Exec(a, fmt.Sprintf(`tinygo test -target=wasi -v -tags "%s" ./...`, strings.Join(tags, ",")))
				return
			}
			cmd.Exec(a, fmt.Sprintf(`go test -v -timeout=20m -tags "%s" ./...`, strings.Join(tags, ",")))
		},
	}))

	goyek.Define(goyek.Task{
		Name:  "wasm",
		Usage: "Builds the WebAssembly module.",
		Action: func(a *goyek.A) {
			buildWasm(a)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "update",
		Usage: "Checks upstream repo for a new version and updates and builds if so.",
		Action: func(a *goyek.A) {
			updateWasm(a)
		},
	})

	// Microbenchmarks live in the root module; the wafbench WAF benchmark is its
	// own module and runs in the wafbench directory.
	defineBenchTasks("bench", "")
	defineBenchTasks("wafbench", "wafbench")

	build.DefineTasks(
		build.Tags(tags...),
		build.ExcludeTasks("test-go"),
	)

	boot.Main()
}

type mode byte

const (
	modeWazero mode = iota
	modeCgo
	modeTinyGo
)

func testMode() mode {
	switch strings.ToLower(os.Getenv("LIBINJECTION_TEST_MODE")) {
	case "cgo":
		return modeCgo
	case "tinygo":
		return modeTinyGo
	default:
		return modeWazero
	}
}

// modeTags returns the build tags for the selected test mode. The wazero and
// tinygo modes need no explicit tags (tinygo defines tinygo.wasm itself).
func modeTags() []string {
	if testMode() == modeCgo {
		return []string{"libinjection_cgo"}
	}
	return nil
}

func buildWasm(a *goyek.A) {
	if !cmd.Exec(a, fmt.Sprintf("docker build -t wasilibs-build -f %s .", filepath.Join("buildtools", "wasm", "Dockerfile"))) {
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		a.Fatal(err)
	}
	wasmDir := filepath.Join(wd, "internal", "wasm")
	if err := os.MkdirAll(wasmDir, 0o755); err != nil { //nolint:gosec
		a.Fatal(err)
	}
	cmd.Exec(a, fmt.Sprintf("docker run --rm -v %s:/out wasilibs-build", wasmDir))
}

func updateWasm(a *goyek.A) {
	verPath := filepath.Join("buildtools", "wasm", "version.txt")
	currBytes, err := os.ReadFile(verPath) //nolint:gosec // fixed in-repo path
	if err != nil {
		a.Fatal(err)
	}
	curr := strings.TrimSpace(string(currBytes))

	gh, err := api.DefaultRESTClient()
	if err != nil {
		a.Fatal(err)
	}

	var latest string
	var release *github.RepositoryRelease
	if err := gh.Get(fmt.Sprintf("repos/%s/releases/latest", libraryRepo), &release); err != nil {
		a.Log(err)
	}

	if release != nil {
		latest = release.GetTagName()
	} else {
		a.Log("could not find releases, falling back to tag")

		var tags []github.RepositoryTag
		if err := gh.Get(fmt.Sprintf("repos/%s/tags", libraryRepo), &tags); err != nil {
			a.Error(err)
		}
		if len(tags) == 0 {
			a.Fatal("could not find tags")
		}
		latest = tags[0].GetName()
	}

	if latest == curr {
		fmt.Println("up to date")
		return
	}

	fmt.Println("updating to", latest)
	if err := os.WriteFile(verPath, []byte(latest+"\n"), 0o644); err != nil { //nolint:gosec
		a.Error(err)
	}

	buildWasm(a)
}

// renovate: github.com/golang/perf
const verBenchstat = "v0.0.0-20230221235046-aebcfb61e84c"

type benchMode int

const (
	benchModeWazero benchMode = iota
	benchModeCGO
	benchModeDefault
)

// benchArgs builds the "go test" command for a benchmark run. benchModeDefault
// runs against the reference libinjection-go for comparison, and benchModeCGO
// wraps the C library instead of the WebAssembly module.
func benchArgs(count int, m benchMode) string {
	args := []string{"go", "test", "-bench=.", "-run=^$", "-v", "-timeout=60m"}
	if count > 0 {
		args = append(args, fmt.Sprintf("-count=%d", count))
	}
	switch m {
	case benchModeCGO:
		args = append(args, "-tags=libinjection_cgo")
	case benchModeDefault:
		args = append(args, "-tags=libinjection_bench_default")
	case benchModeWazero:
	}
	args = append(args, "./...")

	return strings.Join(args, " ")
}

// defineBenchTasks registers the benchmark tasks for a module. dir is the module
// directory to run in, or "" for the repository root module.
func defineBenchTasks(name, dir string) {
	var dirOpts []cmd.Option
	if dir != "" {
		dirOpts = append(dirOpts, cmd.Dir(dir))
	}

	goyek.Define(goyek.Task{
		Name:  name,
		Usage: "Runs " + name + " benchmarks in the default configuration, using wazero.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, benchArgs(1, benchModeWazero), dirOpts...)
		},
	})

	goyek.Define(goyek.Task{
		Name:  name + "-cgo",
		Usage: "Runs " + name + " benchmarks using cgo. A C toolchain and libinjection must be installed.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, benchArgs(1, benchModeCGO), dirOpts...)
		},
	})

	goyek.Define(goyek.Task{
		Name:  name + "-default",
		Usage: "Runs " + name + " benchmarks against the reference libinjection-go for comparison.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, benchArgs(1, benchModeDefault), dirOpts...)
		},
	})

	goyek.Define(goyek.Task{
		Name:  name + "-all",
		Usage: "Runs all " + name + " modes and compares them with benchstat. Requires libinjection for cgo.",
		Action: func(a *goyek.A) {
			if err := os.MkdirAll("out", 0o750); err != nil {
				a.Fatalf("create out directory: %v", err)
			}

			run := func(m benchMode, suffix string) string {
				var stdout bytes.Buffer
				cmd.Exec(a, benchArgs(5, m), append(dirOpts, cmd.Stdout(&stdout))...)
				out := filepath.Join("out", name+suffix+".txt")
				if err := os.WriteFile(out, stdout.Bytes(), 0o600); err != nil {
					a.Fatalf("write %s: %v", out, err)
				}
				return out
			}

			def := run(benchModeDefault, "_default")
			wazero := run(benchModeWazero, "")
			cgo := run(benchModeCGO, "_cgo")

			cmd.Exec(a, fmt.Sprintf("go run golang.org/x/perf/cmd/benchstat@%s %s %s %s", verBenchstat, def, wazero, cgo))
		},
	})
}
