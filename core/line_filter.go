package core

import (
	"math"
	"regexp"
	"runtime"
	"sync"
)

type Payload map[string]any

type CodeLine struct {
	Line      string
	Index     int
	Column    int
	Extracted Payload
}

// LineFilter returns a Payload and a boolean indicating if the line matched
type LineFilter func(line string) (Payload, bool)

// EntropyFilter returns lines whose entropy exceeds a threshold
func EntropyFilter(threshold float64) LineFilter {
	return func(s string) (Payload, bool) {
		e := CalcEntropy(s)
		if e > threshold {
			return Payload{
				"entropy": e, // only the computed value
			}, true
		}
		return nil, false
	}
}

// BasicFilter checks for minimum complexity (length, digit, case, symbol)
func BasicFilter() LineFilter {
	reDigit := regexp.MustCompile(`[0-9]`)
	reUpper := regexp.MustCompile(`[A-Z]`)
	reLower := regexp.MustCompile(`[a-z]`)
	reSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`)

	return func(s string) (Payload, bool) {
		if len(s) < 15 {
			return nil, false
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
			return Payload{
				"complexity": classes,
			}, true
		}
		return nil, false
	}
}

// RegexFilter describes a named regular expression pattern
type RegexFilter struct {
	Header string
	Regex  *regexp.Regexp
}

// AddRegexFilters builds a LineFilter from multiple regex patterns
func AddRegexFilters(patterns []RegexFilter) LineFilter {
	return func(s string) (Payload, bool) {
		matches := make(map[string]string)
		for _, rf := range patterns {
			loc := rf.Regex.FindStringIndex(s)
			if loc != nil {
				matches[rf.Header] = s[loc[0]:loc[1]]
			}
		}
		if len(matches) > 0 {
			return Payload(matches), true
		}
		return nil, false
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

// AnyFilters returns a filter that passes if any filter matches
func AnyFilters(filters ...LineFilter) LineFilter {
	return func(s string) (Payload, bool) {
		merged := make(Payload)
		matched := false
		for _, f := range filters {
			if pl, ok := f(s); ok {
				matched = true
				for k, v := range pl {
					merged[k] = v
				}
			}
		}
		if matched {
			return merged, true
		}
		return nil, false
	}
}

// AllFilters returns a filter that passes only if all filters match
func AllFilters(filters ...LineFilter) LineFilter {
	return func(s string) (Payload, bool) {
		merged := make(Payload)
		for _, f := range filters {
			pl, ok := f(s)
			if !ok {
				return nil, false
			}
			for k, v := range pl {
				merged[k] = v
			}
		}
		return merged, true
	}
}

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
