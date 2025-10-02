package core

// ------------------------------------
// entro_parser.go
// --------------------

import (
	"math"
	"regexp"
)

// CodeLine stores a matching line and its index
type CodeLine struct {
	Line    string
	Index   int
	Column  int
	Entropy float64
}

// LineFilter is a predicate that returns true if a line passes a check
type LineFilter func(string) bool

// entropyFilter returns a filter that checks if a line's entropy > threshold
func EntropyFilter(threshold float64) LineFilter {
	return func(s string) bool {
		if calcEntropy(s) > threshold {
			println("PAY ANTTETION TO THIS: ", calcEntropy(s))
			return true
		} else {
			return false
		}
	}
}

var apiKeyRegex = regexp.MustCompile(`[a-zA-Z0-9_.+/~$-][a-zA-Z0-9_.+/~$=!%:-]{10,1000}[a-zA-Z0-9_.+/=~$!%-]`)

func RegexFilter() LineFilter {
	return func(s string) bool {
		if len(s) >= 24 && len(s) <= 51 {
			return apiKeyRegex.MatchString(s)
		}
		return false
	}
}

// allFilters combines multiple filters into one (logical AND)
func AllFilters(filters ...LineFilter) LineFilter {
	return func(s string) bool {
		for _, f := range filters {
			if f(s) {
				return true
			}
		}
		return false
	}
}


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
