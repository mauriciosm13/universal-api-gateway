package covercheck

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// ChangedLine is an added line from a unified diff.
type ChangedLine struct {
	Path string
	Line int
}

// ParseUnifiedDiff returns added lines in internal/*.go files, excluding tests.
func ParseUnifiedDiff(r io.Reader) ([]ChangedLine, error) {
	var (
		out      []ChangedLine
		path     string
		newLine  int
		inHunk   bool
		skipFile bool
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "+++ "):
			inHunk = false
			raw := strings.TrimPrefix(line, "+++ ")
			raw = strings.TrimPrefix(raw, "b/")
			if raw == "/dev/null" || !isCoveredPath(raw) {
				path = ""
				skipFile = true
				continue
			}
			path = raw
			skipFile = false
		case strings.HasPrefix(line, "@@ "):
			if skipFile || path == "" {
				inHunk = false
				continue
			}
			start, ok := parseHunkNewStart(line)
			if !ok {
				inHunk = false
				continue
			}
			newLine = start
			inHunk = true
		case !inHunk || skipFile || path == "":
			continue
		case strings.HasPrefix(line, "+"):
			out = append(out, ChangedLine{Path: path, Line: newLine})
			newLine++
		case strings.HasPrefix(line, "-"):
			// deleted line does not advance the new-file cursor
		default:
			newLine++
		}
	}
	return out, scanner.Err()
}

func isCoveredPath(p string) bool {
	if !strings.HasPrefix(p, "internal/") || !strings.HasSuffix(p, ".go") {
		return false
	}
	return !strings.HasSuffix(p, "_test.go")
}

func parseHunkNewStart(hunk string) (int, bool) {
	// @@ -old[,len] +new[,len] @@
	plus := strings.Index(hunk, "+")
	if plus < 0 {
		return 0, false
	}
	rest := hunk[plus+1:]
	rest = strings.TrimSpace(strings.SplitN(rest, " ", 2)[0])
	rest = strings.TrimPrefix(rest, "+")
	startRaw, _, _ := strings.Cut(rest, ",")
	n, err := strconv.Atoi(startRaw)
	if err != nil {
		return 0, false
	}
	return n, true
}
