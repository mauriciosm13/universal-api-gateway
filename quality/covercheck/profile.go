package covercheck

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"path"
	"strconv"
	"strings"
)

// Block is one instrumented statement range from a Go cover profile.
type Block struct {
	File       string
	StartLine  int
	EndLine    int
	Statements int
	Count      int
}

// Profile is a parsed Go cover profile limited to module internal/ files.
type Profile struct {
	Blocks []Block
}

// ParseProfile reads a go coverprofile and keeps blocks under modulePath/internal/.
func ParseProfile(r io.Reader, modulePath string) (Profile, error) {
	prefix := strings.TrimSuffix(modulePath, "/") + "/internal/"
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Profile{}, err
		}
		return Profile{}, fmt.Errorf("empty cover profile")
	}
	if !strings.HasPrefix(scanner.Text(), "mode:") {
		return Profile{}, fmt.Errorf("cover profile missing mode line")
	}

	var blocks []Block
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		block, err := parseBlock(line)
		if err != nil {
			return Profile{}, err
		}
		if strings.HasPrefix(block.File, prefix) {
			blocks = append(blocks, block)
		}
	}
	if err := scanner.Err(); err != nil {
		return Profile{}, err
	}
	return Profile{Blocks: blocks}, nil
}

func parseBlock(line string) (Block, error) {
	file, rest, ok := strings.Cut(line, ":")
	if !ok {
		return Block{}, fmt.Errorf("invalid cover line %q", line)
	}
	rangePart, counts, ok := strings.Cut(rest, " ")
	if !ok {
		return Block{}, fmt.Errorf("invalid cover line %q", line)
	}
	start, end, ok := strings.Cut(rangePart, ",")
	if !ok {
		return Block{}, fmt.Errorf("invalid cover line %q", line)
	}
	startLine, _, ok := strings.Cut(start, ".")
	if !ok {
		return Block{}, fmt.Errorf("invalid cover line %q", line)
	}
	endLine, _, ok := strings.Cut(end, ".")
	if !ok {
		return Block{}, fmt.Errorf("invalid cover line %q", line)
	}
	stmtRaw, countRaw, ok := strings.Cut(counts, " ")
	if !ok {
		return Block{}, fmt.Errorf("invalid cover line %q", line)
	}
	sl, err := strconv.Atoi(startLine)
	if err != nil {
		return Block{}, fmt.Errorf("invalid cover start line %q", line)
	}
	el, err := strconv.Atoi(endLine)
	if err != nil {
		return Block{}, fmt.Errorf("invalid cover end line %q", line)
	}
	stmts, err := strconv.Atoi(stmtRaw)
	if err != nil {
		return Block{}, fmt.Errorf("invalid cover statements %q", line)
	}
	count, err := strconv.Atoi(countRaw)
	if err != nil {
		return Block{}, fmt.Errorf("invalid cover count %q", line)
	}
	return Block{
		File:       file,
		StartLine:  sl,
		EndLine:    el,
		Statements: stmts,
		Count:      count,
	}, nil
}

func (p Profile) totals() (covered, total int) {
	for _, b := range p.Blocks {
		total += b.Statements
		if b.Count > 0 {
			covered += b.Statements
		}
	}
	return covered, total
}

func (p Profile) packageTotals(importPath string) (covered, total int) {
	for _, b := range p.Blocks {
		if path.Dir(b.File) != importPath {
			continue
		}
		total += b.Statements
		if b.Count > 0 {
			covered += b.Statements
		}
	}
	return covered, total
}

func (p Profile) hasPackage(importPath string) bool {
	_, total := p.packageTotals(importPath)
	return total > 0
}

func (p Profile) statementCovered(file string, line int) (found, covered bool) {
	for _, b := range p.Blocks {
		if b.File != file {
			continue
		}
		if line < b.StartLine || line > b.EndLine {
			continue
		}
		found = true
		if b.Count > 0 {
			return true, true
		}
	}
	return found, false
}

func percent(covered, total int) float64 {
	if total == 0 {
		return 100
	}
	return math.Round(float64(covered)*1000/float64(total)) / 10
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
