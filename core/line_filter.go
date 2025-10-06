package core

import (
	"math"
	"regexp"
	"runtime"
	"sync"
)

// CodeLine stores a matching line and its metadata
type CodeLineList struct {
	Line     string
	Index    int
	Column   int
	Entropy  float64
	Meta map[string]any
	}

// LineFilter applies a condition to a line and returns a CodeLine if matched
type LineFilter func(line string, index int) *CodeLin


// EntropyFilter returns lines whose entropy exceeds a threshold
func EntropyFilter(threshold float64) LineFilter {
	return func(s string, idx int) *CodeLine {
		e := CalcEntropy(s)
		if e > threshold {
			return &CodeLine{
				Line:     s,
				Index:    idx,
				Entropy:  e,
				Header:   "entropy_threshold",
			}
		}
		return nil
	}
}

// BasicFilter checks for minimum complexity (length, digit, case, symbol)
func BasicFilter() LineFilter {
	reDigit := regexp.MustCompile(`[0-9]`)
	reUpper := regexp.MustCompile(`[A-Z]`)
	reLower := regexp.MustCompile(`[a-z]`)
	reSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`)

	return func(s string, idx int) *CodeLine {
		if len(s) < 15 {
			return nil
		}

		classes := 0
		if reDigit.MatchString(s) {
			classes++
		}
		if reUpper.MatchString(s) {
			classes++
		}
		if reLower.MatchString(s) {
			classes++
		}
		if reSymbol.MatchString(s) {
			classes++
		}

		if classes >= 3 {
			return &CodeLine{
				Line:     s,
				Index:    idx,
				Header:   "basic_filter",
			}
		}
		return nil
	}
}

// RegexFilter describes a named regular expression pattern
type RegexFilter struct {
	Header string
	Regex  *regexp.Regexp
}

// AddRegexFilters builds a LineFilter from multiple regex patterns
func AddRegexFilters(patterns []RegexFilter) LineFilter {
	return func(s string, idx int) *CodeLine {
		for _, rf := range patterns {
			loc := rf.Regex.FindStringIndex(s)
			if loc != nil {
				return &CodeLine{
					Line:     s,
					Index:    idx,
					Column:   loc[0],
					Header:   rf.Header,
				}
			}
		}
		return nil
	}
}

// LoadRegex compiles a single regex into a LineFilter
func LoadRegex(header, regexPattern string) (LineFilter, error) {
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}
	return AddRegexFilters([]RegexFilter{{Header: header, Regex: re}}), nil
}

// =====================
//   Filter Combinators
// =====================

// AllFilters returns a filter that passes only if *all* filters match
func AllFilters(filters ...LineFilter) LineFilter {
	return func(s string, idx int) *CodeLine {
		var combined *CodeLine
		for _, f := range filters {
			result := f(s, idx)
			if result == nil {
				return nil
			}
			if combined == nil {
				combined = result
			} else {
				// Combine metadata
				if result.Entropy > combined.Entropy {
					combined.Entropy = result.Entropy
				}
				if result.Header != "" && combined.Header == "" {
					combined.Header = result.Header
				}
			}
		}
		return combined
	}
}

// AnyFilters returns a filter that passes if *any* filter matches
func AnyFilters(filters ...LineFilter) LineFilter {
	return func(s string, idx int) *CodeLine {
		for _, f := range filters {
			if result := f(s, idx); result != nil {
				return result
			}
		}
		return nil
	}
}

// =====================
//     Entropy Utils
// =====================

// CalcEntropy computes Shannon entropy (single-threaded)
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

// CalcEntropyParallel computes entropy using multiple CPU cores
func CalcEntropyParallel(line string) float64 {
	x := len(line)
	if x == 0 {
		return 0.0
	}

	numCPU := runtime.NumCPU()
	chunkSize := (x + numCPU - 1) / numCPU

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
