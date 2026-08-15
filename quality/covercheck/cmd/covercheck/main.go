package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mauriciomendonca/universal-api-gateway/quality/covercheck"
)

func main() {
	var (
		profilePath  = flag.String("profile", "coverage.out", "go cover profile")
		configPath   = flag.String("config", "quality/config/coverage.yaml", "coverage config")
		baselinePath = flag.String("baseline", "quality/baselines/quality-baseline.json", "quality baseline")
		diffPath     = flag.String("diff", "", "unified diff of changed files")
		reportPath   = flag.String("report", "quality/reports/coverage.json", "JSON report output")
		requireDiff  = flag.Bool("require-diff", false, "fail when the base-ref diff is missing")
		diffOK       = flag.Bool("diff-ok", false, "true when the caller produced a valid base-ref diff")
		modulePath   = flag.String("module", "", "Go module path (default: go list -m)")
	)
	flag.Parse()

	cfgFile, err := os.Open(*configPath)
	must(err, "open config")
	cfg, err := covercheck.ParseConfigYAML(cfgFile)
	_ = cfgFile.Close()
	must(err, "parse config")

	baseFile, err := os.Open(*baselinePath)
	must(err, "open baseline")
	baseline, err := covercheck.ParseBaselineJSON(baseFile)
	_ = baseFile.Close()
	must(err, "parse baseline")

	profFile, err := os.Open(*profilePath)
	must(err, "open cover profile")

	mod := *modulePath
	if mod == "" {
		mod, err = goListModule()
		must(err, "go list -m")
	}

	profile, err := covercheck.ParseProfile(profFile, mod)
	_ = profFile.Close()
	must(err, "parse cover profile")

	pkgs, err := listInternalPackages()
	must(err, "list internal packages")

	var changed []covercheck.ChangedLine
	haveDiff := *diffOK
	if *diffPath != "" {
		raw, readErr := os.ReadFile(*diffPath)
		must(readErr, "read diff")
		changed, err = covercheck.ParseUnifiedDiff(bytes.NewReader(raw))
		must(err, "parse diff")
	}

	rep := covercheck.Evaluate(covercheck.Input{
		Config:      cfg,
		Baseline:    baseline,
		Profile:     profile,
		Packages:    pkgs,
		Changed:     changed,
		ModulePath:  mod,
		DiffOK:      haveDiff,
		RequireDiff: *requireDiff,
	})

	for _, msg := range rep.Messages {
		fmt.Println(msg)
	}

	if *reportPath != "" {
		if err := os.MkdirAll(filepath.Dir(*reportPath), 0o755); err != nil {
			must(err, "create report dir")
		}
		out, err := json.MarshalIndent(rep, "", "  ")
		must(err, "marshal report")
		must(os.WriteFile(*reportPath, append(out, '\n'), 0o644), "write report")
	}

	if rep.Status != "PASS" {
		os.Exit(1)
	}
}

func must(err error, what string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %s: %v\n", what, err)
		os.Exit(1)
	}
}

func goListModule() (string, error) {
	out, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

type listPackage struct {
	ImportPath   string
	Dir          string
	GoFiles      []string
	TestGoFiles  []string
	XTestGoFiles []string
}

func listInternalPackages() ([]covercheck.Package, error) {
	cmd := exec.Command("go", "list", "-json", "./internal/...")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []covercheck.Package
	for {
		var lp listPackage
		if err := dec.Decode(&lp); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		pkgs = append(pkgs, covercheck.Package{
			ImportPath: lp.ImportPath,
			Dir:        lp.Dir,
			GoFiles:    lp.GoFiles,
			HasTests:   len(lp.TestGoFiles) > 0 || len(lp.XTestGoFiles) > 0,
		})
	}
	return pkgs, nil
}
