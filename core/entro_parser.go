package core

import (
	"math"
	"regexp"
	"runtime"
	"sync"
)

// CodeLine stores a matching line and its index
type CodeLine struct {
	Line    string
	Index   int
	Column  int
	Entropy float64
}

// RegexFilter represents a regex + optional label
type RegexFilter struct {
	Header string
	Regex  *regexp.Regexp
}

// LineFilter is a predicate that returns true if a line passes a check
type LineFilter func(string) bool

// EntropyFilter returns a filter that checks if a line's entropy > threshold
func EntropyFilter(threshold float64) LineFilter {
	return func(s string) bool {
		return CalcEntropy(s) > threshold
	}
}

// BasicFilter checks for minimum complexity: length, digits, cases, symbols
func BasicFilter() LineFilter {
	return func(s string) bool {
		if len(s) < 15 {
			return false
		}
		hasDigit := regexp.MustCompile(`[0-9]`).MatchString(s)
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(s)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(s)
		hasSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(s)

		// Require at least 3 out of 4 classes
		classes := 0
		if hasDigit {
			classes++
		}
		if hasUpper {
			classes++
		}
		if hasLower {
			classes++
		}
		if hasSymbol {
			classes++
		}
		return classes >= 3
	}
}

// AllFilters (AND)
func AllFilters(filters ...LineFilter) LineFilter {
	return func(s string) bool {
		for _, f := range filters {
			if !f(s) {
				return false
			}
		}
		return true
	}
}

// AnyFilters (OR)
func AnyFilters(filters ...LineFilter) LineFilter {
	return func(s string) bool {
		for _, f := range filters {
			if f(s) {
				return true
			}
		}
		return false
	}
}

// AddRegexFilters builds a filter from regex patterns
func AddRegexFilters(patterns []RegexFilter) LineFilter {
	return func(s string) bool {
		for _, rf := range patterns {
			if rf.Regex.MatchString(s) {
				return true
			}
		}
		return false
	}
}

// CalcEntropy computes Shannon entropy safely
func CalcEntropyParallel(line string) float64 {
	x := len(line)
	if x == 0 {
		return 0.0
	}

	numCPU := runtime.NumCPU()
	chunkSize := (x + numCPU - 1) / numCPU
	/*	
	if 300 > x{
		return calcEntropySingle(line)
	}*/

	var wg sync.WaitGroup
	wg.Add(numCPU)

	chResults := make(chan [256]float64, numCPU)

	for i := 0; i < numCPU; i++ {
		go func(chunkNum int) {
			defer wg.Done()
			var uniqCounter [256]float64

			start := chunkNum * chunkSize
			end := start + chunkSize
			if end > x {
				end = x
			}
			if start >= end {
				chResults <- uniqCounter
				return
			}
			for _, val := range line[start:end] {
				uniqCounter[val]++
			}
			chResults <- uniqCounter
		}(i)
	}

	go func() {
		wg.Wait()
		close(chResults)
	}()

	var totalFreq [256]float64
	for freqArray := range chResults {
		for j, val := range freqArray {
			totalFreq[j] += val
		}
	}

	length := float64(x)
	entropy := 0.0
	for _, count := range totalFreq {
		if count > 0 {
			p := count / length
			entropy -= p * math.Log2(p)
		}
	}

	if math.IsNaN(entropy) || math.IsInf(entropy, 0) {
		return 0.0
	}
	return entropy
}

func CalcEntropy(line string) float64 {
    freq := make(map[rune]float64)
    for _, val := range line {
        freq[val]++
    }
    length := float64(len(line))
    if length == 0 {
        return 0.0
    }

    entropy := 0.0
    for _, count := range freq {
        if count > 0 {
            p := count / length
            entropy -= p * math.Log2(p)
        }
    }
    return entropy
}