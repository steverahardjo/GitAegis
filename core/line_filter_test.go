package core

import (
	"math"
	"testing"
)

func TestCalcEntropySingle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
		wantHigh bool
	}{
		{"empty", "", true, false},
		{"single char", "a", true, false},
		{"repeated", "aaaaaaaa", true, false},
		{"two chars", "ab", false, false},
		{"high entropy", "aB3$kL9@mX2#pQ5!", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcEntropySingle(tt.input)

			if tt.wantZero && got > 0.01 {
				t.Errorf("calcEntropySingle(%q) = %f, want ~0", tt.input, got)
			}
			if tt.wantHigh && got < 2.0 {
				t.Errorf("calcEntropySingle(%q) = %f, want high entropy", tt.input, got)
			}
		})
	}
}

func TestCalcEntropyParallel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{"empty", "", true},
		{"single char", "a", true},
		{"repeated", "aaaaaaaa", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcEntropyParallel(tt.input)

			if tt.wantZero && got > 0.01 {
				t.Errorf("calcEntropyParallel(%q) = %f, want ~0", tt.input, got)
			}
		})
	}
}

func TestCalcEntropy_AutoSelect(t *testing.T) {
	short := "abc123XYZ"
	long := "aB3$kL9@mX2#pQ5!rT8&nV1^wY4*uI7"

	shortEnt := CalcEntropy(short)
	longEnt := CalcEntropy(long)

	if shortEnt < 0 {
		t.Error("short string entropy should not be negative")
	}
	if longEnt < 0 {
		t.Error("long string entropy should not be negative")
	}

	t.Logf("short entropy: %f, long entropy: %f", shortEnt, longEnt)
}

func TestCalcEntropy_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unicode", "你好世界"},
		{"emoji", "😀😁😂"},
		{"mixed", "abc123你好"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcEntropy(tt.input)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("CalcEntropy(%q) returned invalid value: %f", tt.input, got)
			}
		})
	}
}

func TestBasicFilter(t *testing.T) {
	filter := BasicFilter()

	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"too short", "abc123", false},
		{"no digits", "abcdefghijklmn", false},
		{"no uppercase", "abc123!@#defgh", false},
		{"valid all classes", "abcDEF123!@#xyz", true},
		{"valid 3 classes", "abcDEF123456xyz", true},
		{"exactly 15 chars", "abcDEF123456789", true},
		{"14 chars", "abcDEF12345678", false},
		{"only lowercase and digits", "abcdefghij12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := filter(tt.input)
			if got != tt.wantMatch {
				t.Errorf("BasicFilter(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
		})
	}
}

func TestEntropyFilter(t *testing.T) {
	filter := EntropyFilter(4.0)

	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"low entropy", "aaaaaaaaaa", false},
		{"medium entropy", "abcabcabc", false},
		{"high entropy", "aB3$kL9@mX2#pQ5!rT8&nV1^wY4", true},
		{"random looking", "xK9#mP2$vL5@nQ8!rT3&uI7", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, got := filter(tt.input)
			if got != tt.wantMatch {
				t.Errorf("EntropyFilter(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
			if got && payload == nil {
				t.Error("expected payload when match is true")
			}
		})
	}
}

func TestAnyFilters(t *testing.T) {
	filter1 := func(s string) (Payload, bool) {
		if s == "match1" {
			return Payload{"type": "filter1"}, true
		}
		return nil, false
	}
	filter2 := func(s string) (Payload, bool) {
		if s == "match2" {
			return Payload{"type": "filter2"}, true
		}
		return nil, false
	}

	combined := AnyFilters(filter1, filter2)

	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"match first", "match1", true},
		{"match second", "match2", true},
		{"no match", "nomatch", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, got := combined(tt.input)
			if got != tt.wantMatch {
				t.Errorf("AnyFilters(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
			if got && payload == nil {
				t.Error("expected payload when match is true")
			}
		})
	}
}

func TestAllFilters(t *testing.T) {
	filter1 := func(s string) (Payload, bool) {
		if len(s) > 5 {
			return Payload{"len": "long"}, true
		}
		return nil, false
	}
	filter2 := func(s string) (Payload, bool) {
		if s[0] == 'a' {
			return Payload{"start": "a"}, true
		}
		return nil, false
	}

	combined := AllFilters(filter1, filter2)

	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"both match", "abcdef", true},
		{"only len", "bcdefg", false},
		{"only start", "a", false},
		{"neither", "b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := combined(tt.input)
			if got != tt.wantMatch {
				t.Errorf("AllFilters(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
		})
	}
}

func TestAddTargetRegexPattern(t *testing.T) {
	filter := AddTargetRegexPattern("aws_key", `AKIA[0-9A-Z]{16}`)

	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"valid aws key", "AKIAIOSFODNN7EXAMPLE", true},
		{"invalid aws key", "AKIA123", false},
		{"no match", "notakey", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, got := filter(tt.input)
			if got != tt.wantMatch {
				t.Errorf("AddTargetRegexPattern(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
			if got && payload == nil {
				t.Error("expected payload when match is true")
			}
		})
	}
}

func TestAddTargetRegexPattern_InvalidRegex(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	filter := AddTargetRegexPattern("test", `[invalid`)
	if filter != nil {
		_, _ = filter("test")
	}
}
