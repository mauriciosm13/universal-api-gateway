package covercheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfigYAML(t *testing.T) {
	t.Parallel()

	cfg, err := ParseConfigYAML(strings.NewReader(`coverage:
  global_minimum: 85
  changed_code_minimum: 90
  fail_on_regression: true
  allowed_drop_percentage: 0
`))
	if err != nil {
		t.Fatalf("ParseConfigYAML() error = %v", err)
	}
	if cfg.GlobalMinimum != 85 || cfg.ChangedCodeMinimum != 90 || !cfg.FailOnRegression || cfg.AllowedDropPercentage != 0 {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestParseConfigYAMLMissingKey(t *testing.T) {
	t.Parallel()

	_, err := ParseConfigYAML(strings.NewReader("coverage:\n  global_minimum: 85\n"))
	if err == nil {
		t.Fatal("ParseConfigYAML() error = nil, want missing key")
	}
}

func TestParseBaselineJSON(t *testing.T) {
	t.Parallel()

	b, err := ParseBaselineJSON(strings.NewReader(`{"coverage":{"measured_global":90.1}}`))
	if err != nil {
		t.Fatalf("ParseBaselineJSON() error = %v", err)
	}
	if b.MeasuredGlobal != 90.1 {
		t.Fatalf("MeasuredGlobal = %v, want 90.1", b.MeasuredGlobal)
	}
}

func TestParseProfileFiltersInternal(t *testing.T) {
	t.Parallel()

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.2,12.3 2 1
example.com/gw/quality/covercheck/gate.go:1.1,2.2 1 1
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}
	if len(prof.Blocks) != 1 {
		t.Fatalf("Blocks = %d, want 1 internal block", len(prof.Blocks))
	}
	covered, total := prof.totals()
	if covered != 2 || total != 2 {
		t.Fatalf("totals covered=%d total=%d, want 2/2", covered, total)
	}
}

func TestParseUnifiedDiffAddedLines(t *testing.T) {
	t.Parallel()

	diff := `diff --git a/internal/server/server.go b/internal/server/server.go
--- a/internal/server/server.go
+++ b/internal/server/server.go
@@ -10,0 +11,2 @@
+func New() {}
+func Start() {}
diff --git a/internal/server/server_test.go b/internal/server/server_test.go
--- a/internal/server/server_test.go
+++ b/internal/server/server_test.go
@@ -1,0 +2,1 @@
+func TestNew(t *testing.T) {}
`
	lines, err := ParseUnifiedDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2 production lines", len(lines))
	}
	if lines[0].Path != "internal/server/server.go" || lines[0].Line != 11 {
		t.Fatalf("first line = %+v, want path internal/server/server.go line 11", lines[0])
	}
	if lines[1].Line != 12 {
		t.Fatalf("second line = %d, want 12", lines[1].Line)
	}
}

func TestEvaluateGlobalAndChangedStatements(t *testing.T) {
	t.Parallel()

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.1,12.2 2 1
example.com/gw/internal/server/server.go:20.1,22.2 2 0
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	rep := Evaluate(Input{
		Config: Config{
			GlobalMinimum:         50,
			ChangedCodeMinimum:    90,
			FailOnRegression:      true,
			AllowedDropPercentage: 0,
		},
		Baseline:   Baseline{MeasuredGlobal: 50},
		Profile:    prof,
		ModulePath: module,
		DiffOK:     true,
		Changed: []ChangedLine{
			{Path: "internal/server/server.go", Line: 10},
			{Path: "internal/server/server.go", Line: 11},
		},
		Packages: []Package{{ImportPath: module + "/internal/server", HasTests: true}},
	})
	if rep.Status != "PASS" {
		t.Fatalf("status = %s messages=%v", rep.Status, rep.Messages)
	}
	if rep.Global != 50 {
		t.Fatalf("global = %v, want 50", rep.Global)
	}
	if rep.ChangedCoverage != 100 {
		t.Fatalf("changed = %v, want 100", rep.ChangedCoverage)
	}
}

func TestEvaluateChangedCodeBelowMinimumFails(t *testing.T) {
	t.Parallel()

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.1,12.2 2 0
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	rep := Evaluate(Input{
		Config: Config{
			GlobalMinimum:      0,
			ChangedCodeMinimum: 90,
		},
		Profile:    prof,
		ModulePath: module,
		DiffOK:     true,
		Changed:    []ChangedLine{{Path: "internal/server/server.go", Line: 11}},
		Packages:   []Package{{ImportPath: module + "/internal/server", HasTests: true}},
	})
	if rep.Status != "FAIL" || rep.ChangedOK {
		t.Fatalf("status=%s changedOK=%v messages=%v", rep.Status, rep.ChangedOK, rep.Messages)
	}
}

func TestEvaluateRegressionFails(t *testing.T) {
	t.Parallel()

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.1,12.2 2 1
example.com/gw/internal/server/server.go:20.1,22.2 2 0
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	rep := Evaluate(Input{
		Config: Config{
			GlobalMinimum:         0,
			ChangedCodeMinimum:    0,
			FailOnRegression:      true,
			AllowedDropPercentage: 0,
		},
		Baseline:   Baseline{MeasuredGlobal: 90},
		Profile:    prof,
		ModulePath: module,
		DiffOK:     true,
		Packages:   []Package{{ImportPath: module + "/internal/server", HasTests: true}},
	})
	if rep.Status != "FAIL" || rep.RegressionOK {
		t.Fatalf("status=%s regressionOK=%v messages=%v", rep.Status, rep.RegressionOK, rep.Messages)
	}
}

func TestEvaluateUntestedExecutablePackageFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := "package otelsetup\n\nfunc Install() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "setup.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.1,12.2 1 1
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	rep := Evaluate(Input{
		Config:     Config{GlobalMinimum: 0, ChangedCodeMinimum: 0},
		Profile:    prof,
		ModulePath: module,
		DiffOK:     true,
		Packages: []Package{{
			ImportPath: module + "/internal/observability/otelsetup",
			Dir:        dir,
			GoFiles:    []string{"setup.go"},
			HasTests:   false,
		}},
	})
	if rep.Status != "FAIL" || len(rep.UntestedPackages) != 1 {
		t.Fatalf("status=%s untested=%v messages=%v", rep.Status, rep.UntestedPackages, rep.Messages)
	}
}

func TestEvaluateTypesOnlyPackageWithoutTestsPasses(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := "package port\n\ntype Router interface{ Resolve() }\n"
	if err := os.WriteFile(filepath.Join(dir, "router.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.1,12.2 1 1
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	rep := Evaluate(Input{
		Config:     Config{GlobalMinimum: 0, ChangedCodeMinimum: 0},
		Profile:    prof,
		ModulePath: module,
		DiffOK:     true,
		Packages: []Package{{
			ImportPath: module + "/internal/routing/port",
			Dir:        dir,
			GoFiles:    []string{"router.go"},
			HasTests:   false,
		}},
	})
	if rep.Status != "PASS" {
		t.Fatalf("status=%s messages=%v", rep.Status, rep.Messages)
	}
}

func TestEvaluateRequireDiffFailsWhenMissing(t *testing.T) {
	t.Parallel()

	const module = "example.com/gw"
	prof, err := ParseProfile(strings.NewReader(`mode: atomic
example.com/gw/internal/server/server.go:10.1,12.2 1 1
`), module)
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	rep := Evaluate(Input{
		Config:      Config{GlobalMinimum: 0, ChangedCodeMinimum: 90},
		Profile:     prof,
		ModulePath:  module,
		DiffOK:      false,
		RequireDiff: true,
		Packages:    []Package{{ImportPath: module + "/internal/server", HasTests: true}},
	})
	if rep.Status != "FAIL" || rep.ChangedOK {
		t.Fatalf("status=%s changedOK=%v messages=%v", rep.Status, rep.ChangedOK, rep.Messages)
	}
}

func TestFileHasFuncDecl(t *testing.T) {
	t.Parallel()

	if !fileHasFuncDecl([]byte("package p\nfunc F() {}\n")) {
		t.Fatal("expected func decl")
	}
	if fileHasFuncDecl([]byte("package p\ntype T struct{ X int }\n")) {
		t.Fatal("struct-only file should not count as executable")
	}
}
