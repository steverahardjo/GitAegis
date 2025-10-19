package tests

import (
	"strings"
	"testing"
	"time"

	core "github.com/steverahardjo/GitAegis/core"
)

func TestParallel_vs_NormalEntropy(t *testing.T) {
	text := strings.Repeat("abcdefghijklmnopqrstuvwxyz0123456789", 100000)

	startNormal := time.Now()
	normalRun := core.CalcEntropy(text)
	durationNormal := time.Since(startNormal)
	t.Logf("Normal entropy: %.6f, time taken: %v", normalRun, durationNormal)

	startParal := time.Now()
	paralRun := core.CalcEntropyParallel(text) 
	durationParal := time.Since(startParal)
	t.Logf("Parallel entropy: %.6f, time taken: %v", paralRun, durationParal)

	if diff := paralRun - normalRun; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("Entropy mismatch between normal and parallel runs (diff=%f)", diff)
	}

	// Compare speed
	if durationParal >= durationNormal {
		t.Errorf("Our parallel isnt faster")
	} else {
		t.Logf("Parallel entropy is faster (normal=%v, parallel=%v)", durationNormal, durationParal)
	}
}
