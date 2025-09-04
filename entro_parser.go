package main

import (
	"fmt"
	"math"
)

// calcEntropy computes Shannon entropy of a string
func calcEntropy(line string) float64 {
	if len(line) == 0 {
		return 0.0
	}

	// Count frequencies of runes (Unicode-safe)
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

func main() {
	fmt.Println("Entropy of 'api key':", calcEntropy("OPENAI_KEY = "))
	fmt.Println("Entropy of 'normal python syntax ':", calcEntropy("for key, value in maps.items():"))
	fmt.Println("Entropy of 'aaaaaa':", calcEntropy("aaaaaaaa"))
}

