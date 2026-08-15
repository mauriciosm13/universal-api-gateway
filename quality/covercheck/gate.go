package covercheck

import (
	"fmt"
	"sort"
	"strings"
)

// Input is a complete coverage-gate evaluation.
type Input struct {
	Config      Config
	Baseline    Baseline
	Profile     Profile
	Packages    []Package
	Changed     []ChangedLine
	ModulePath  string
	DiffOK      bool
	RequireDiff bool
}

// Report is the machine-readable coverage-gate result.
type Report struct {
	Status            string   `json:"status"`
	Global            float64  `json:"global"`
	GlobalMinimum     float64  `json:"global_minimum"`
	GlobalOK          bool     `json:"global_ok"`
	BaselineGlobal    float64  `json:"baseline_global"`
	RegressionOK      bool     `json:"regression_ok"`
	ChangedCoverage   float64  `json:"changed_coverage"`
	ChangedMinimum    float64  `json:"changed_minimum"`
	ChangedOK         bool     `json:"changed_ok"`
	ChangedSkipped    bool     `json:"changed_skipped"`
	ChangedStatements int      `json:"changed_statements"`
	UntestedPackages  []string `json:"untested_packages"`
	Messages          []string `json:"messages"`
}

// Evaluate applies global, regression, untested-package, and changed-statement gates.
func Evaluate(in Input) Report {
	rep := Report{
		GlobalMinimum:  in.Config.GlobalMinimum,
		BaselineGlobal: round1(in.Baseline.MeasuredGlobal),
		ChangedMinimum: in.Config.ChangedCodeMinimum,
		ChangedOK:      true,
		RegressionOK:   true,
	}

	covered, total := in.Profile.totals()
	rep.Global = percent(covered, total)
	rep.GlobalOK = rep.Global+1e-9 >= in.Config.GlobalMinimum
	if !rep.GlobalOK {
		rep.Messages = append(rep.Messages,
			fmt.Sprintf("FAIL: global coverage %.1f%% is below minimum %.0f%%", rep.Global, in.Config.GlobalMinimum))
	} else {
		rep.Messages = append(rep.Messages,
			fmt.Sprintf("PASS: global coverage %.1f%% (minimum %.0f%%)", rep.Global, in.Config.GlobalMinimum))
	}

	if in.Config.FailOnRegression {
		floor := round1(in.Baseline.MeasuredGlobal - in.Config.AllowedDropPercentage)
		rep.RegressionOK = round1(rep.Global)+1e-9 >= floor
		if !rep.RegressionOK {
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("FAIL: coverage %.1f%% regressed below baseline %.1f%% (allowed drop %.0f%%)",
					rep.Global, round1(in.Baseline.MeasuredGlobal), in.Config.AllowedDropPercentage))
		} else {
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("PASS: coverage %.1f%% meets baseline %.1f%%", rep.Global, round1(in.Baseline.MeasuredGlobal)))
		}
	}

	untested := untestedExecutable(in.Profile, in.Packages)
	rep.UntestedPackages = untested
	if len(untested) > 0 {
		rep.Messages = append(rep.Messages,
			fmt.Sprintf("FAIL: executable internal packages without tests: %s", strings.Join(untested, ", ")))
	}

	switch {
	case in.RequireDiff && !in.DiffOK:
		rep.ChangedOK = false
		rep.Messages = append(rep.Messages, "FAIL: could not diff against base ref for changed-code coverage")
	case !in.DiffOK:
		rep.ChangedSkipped = true
		rep.Messages = append(rep.Messages, "SKIP: changed-code coverage (base ref not available)")
	default:
		changedCovered, changedTotal := changedStatementCoverage(in.Profile, in.Changed, in.ModulePath)
		rep.ChangedStatements = changedTotal
		if changedTotal == 0 {
			rep.ChangedSkipped = true
			rep.ChangedCoverage = 100
			rep.Messages = append(rep.Messages, "SKIP: no coverable statements in changed internal/ files")
		} else {
			rep.ChangedCoverage = percent(changedCovered, changedTotal)
			rep.ChangedOK = rep.ChangedCoverage+1e-9 >= in.Config.ChangedCodeMinimum
			if !rep.ChangedOK {
				rep.Messages = append(rep.Messages,
					fmt.Sprintf("FAIL: changed-code coverage %.1f%% is below minimum %.0f%% (%d/%d statements)",
						rep.ChangedCoverage, in.Config.ChangedCodeMinimum, changedCovered, changedTotal))
			} else {
				rep.Messages = append(rep.Messages,
					fmt.Sprintf("PASS: changed-code coverage %.1f%% (minimum %.0f%%)",
						rep.ChangedCoverage, in.Config.ChangedCodeMinimum))
			}
		}
	}

	rep.Status = "PASS"
	if !rep.GlobalOK || !rep.RegressionOK || !rep.ChangedOK || len(untested) > 0 {
		rep.Status = "FAIL"
	}
	return rep
}

func untestedExecutable(profile Profile, pkgs []Package) []string {
	out := make([]string, 0)
	for _, pkg := range pkgs {
		if pkg.HasTests {
			continue
		}
		if profile.hasPackage(pkg.ImportPath) {
			continue
		}
		if !packageHasFuncs(pkg) {
			continue
		}
		out = append(out, pkg.ImportPath)
	}
	sort.Strings(out)
	return out
}

func changedStatementCoverage(profile Profile, changed []ChangedLine, modulePath string) (covered, total int) {
	seen := map[string]bool{}
	prefix := strings.TrimSuffix(modulePath, "/") + "/"
	for _, cl := range changed {
		file := prefix + cl.Path
		key := fmt.Sprintf("%s:%d", file, cl.Line)
		if seen[key] {
			continue
		}
		found, hit := profile.statementCovered(file, cl.Line)
		if !found {
			continue
		}
		seen[key] = true
		total++
		if hit {
			covered++
		}
	}
	return covered, total
}
