package main

// ------------------------------------
// entro_parser.go
// --------------------

import (
	"bufio"
	"math"
	"os"
	"regexp"
	"strings"

)

// CodeLine stores a matching line and its index
type CodeLine struct {
	Line  string
	Index int
}:\

// --------------------
// Filters
// --------------------

// LineFilter is a predicate that returns true if a line passes a check
type LineFilter func(string) bool

// entropyFilter returns a filter that checks if a line's entropy > threshold
func entropyFilter(threshold float64) LineFilter {
	return func(s string) bool {
		return calcEntropy(s) > threshold
	}
}

var candidateRe = regexp.MustCompile(
	`^[a-zA-Z0-9_.+/~$-][a-zA-Z0-9_.+/=~$-]+[a-zA-Z0-9_.+/=~$-]$`,
)

func regexFilter() LineFilter {
	return func(s string) bool {
		// Regex structure check
		if !candidateRe.MatchString(s) {
			return false
		}
		// Enforce length between 16 and 1024 in code
		if len(s) < 16 || len(s) > 1024 {
			return false
		}
		// Manual "negative lookahead" checks
		if strings.Contains(s, `\n`) ||
			strings.Contains(s, `\t`) ||
			strings.Contains(s, `\r`) ||
			strings.Contains(s, `\"`) {
			return false
		}
		return true
	}
}

// allFilters combines multiple filters into one (logical AND)
func allFilters(filters ...LineFilter) LineFilter {
	return func(s string) bool {
		for _, f := range filters {
			if !f(s) {
				return false
			}
		}
		return true
	}
}

// --------------------
// Core logic
// --------------------

// calcEntropy computes Shannon entropy of a string
func calcEntropy(line string) float64 {
	if len(line) == 0 {
		return 0.0
	}

	uniqCounter := make(map[rune]int)
	for _, r := range line {
		uniqCounter[r]++
	}

	length := float64(len(line))
	entropy := 0.0

	for _, count := range uniqCounter {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// readAndCalc reads lines from a file and returns those
// that pass the provided filter
func readAndCalc(filename string, filter LineFilter) ([]CodeLine, error) {
	var collection = []CodeLine{}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	currIndex := 0

	for scanner.Scan() {
		currIndex++
		line := scanner.Text()
		if filter(line) {
			collection = append(collection, CodeLine{line, currIndex})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return collection, nil
}
