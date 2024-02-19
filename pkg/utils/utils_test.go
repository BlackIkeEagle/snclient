package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUtilsExpandDuration(t *testing.T) {
	tests := []struct {
		in  string
		res float64
	}{
		{"2d", 86400 * 2},
		{"1m", 60},
		{"10s", 10},
		{"100ms", 0.1},
		{"100", 100},
		{"-1h", -3600},
	}

	for _, tst := range tests {
		res, err := ExpandDuration(tst.in)
		require.NoError(t, err)
		assert.InDeltaf(t, tst.res, res, 0.00001, "ExpandDuration: %s", tst.in)
	}
}

func TestUtilsIsFloatVal(t *testing.T) {
	tests := []struct {
		in  float64
		res bool
	}{
		{1.00, false},
		{1.01, true},
		{5, false},
	}

	for _, tst := range tests {
		res := IsFloatVal(tst.in)
		assert.Equalf(t, tst.res, res, "IsFloatVal: %s", tst.in)
	}
}

func TestUtilsExecPath(t *testing.T) {
	execPath, _, _, err := GetExecutablePath()
	require.NoErrorf(t, err, "GetExecutablePath works")

	assert.NotEmptyf(t, execPath, "GetExecutablePath")
}

func TestToPrecision(t *testing.T) {
	tests := []struct {
		in        float64
		precision int
		res       float64
	}{
		{1.001, 0, 1},
		{1.001, 3, 1.001},
		{1.0013, 3, 1.001},
	}

	for _, tst := range tests {
		res := ToPrecision(tst.in, tst.precision)
		assert.InDeltaf(t, tst.res, res, 0.00001, "ToPrecision: %v (precision: %d) -> %v", tst.in, tst.precision, res)
	}
}

func TestTokenizer(t *testing.T) {
	tests := []struct {
		in  string
		res []string
	}{
		{"", []string{""}},
		{"a bc d", []string{"a", "bc", "d"}},
		{"a 'bc' d", []string{"a", "'bc'", "d"}},
		{"a 'b c' d", []string{"a", "'b c'", "d"}},
		{`a "b'c" d`, []string{"a", `"b'c"`, "d"}},
		{`a 'b""c' d`, []string{"a", `'b""c'`, "d"}},
	}

	for _, tst := range tests {
		res := Tokenize(tst.in)
		assert.Equalf(t, tst.res, res, "Tokenize: %v -> %v", tst.in, res)
	}
}

func TestTokenizerShell(t *testing.T) {
	tests := []struct {
		in  string
		res []string
	}{
		{"", []string{""}},
		{" a", []string{"a"}},
		{" a ", []string{"a"}},
		{"a bc d", []string{"a", "bc", "d"}},
		{"a 'bc' d", []string{"a", "bc", "d"}},
		{"a 'b c' d", []string{"a", "b c", "d"}},
		{`a "b'c" d`, []string{"a", `b'c`, "d"}},
		{`a 'b""c' d`, []string{"a", `b""c`, "d"}},
		{`a  """b""" '' ''c'' ''d'' ee""ee f' 'f '" "' "' ''"`, []string{`a`, `b`, ``, `c`, `d`, `eeee`, `f f`, `" "`, `' ''`}},
		{`"\'"`, []string{`\'`}},
		{`"\"'"`, []string{`"'`}},
		{`\'`, []string{`'`}},
		{`\"`, []string{`"`}},
		{`'"\a"'`, []string{`"\a"`}},
		{`\ a`, []string{` a`}},
		{`\\ a`, []string{`\`, `a`}},
		{`\\\ a`, []string{`\ a`}},
		{`\\\\ a`, []string{`\\`, `a`}},
		{`"\\\\ a"`, []string{`\\ a`}},
		{`'\\\\ a'`, []string{`\\\\ a`}},
	}

	for _, tst := range tests {
		res, err := TokenizeShell(tst.in)
		assert.Equalf(t, tst.res, res, "Tokenize: %v -> %v", tst.in, res)
		require.NoError(t, err)
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		in  string
		res float64
	}{
		{"1.0", 1.0},
		{"0.1", 0.001},
		{"0.1.23", 0.001023},
	}

	for _, tst := range tests {
		res := ParseVersion(tst.in)
		assert.InDeltaf(t, tst.res, res, 0.00001, "ParseVersion: %v -> %v", tst.in, res)
	}
}

func TestDurationString(t *testing.T) {
	tests := []struct {
		in  time.Duration
		res string
	}{
		{time.Second * 5, "5000ms"},
		{time.Second * 90, "1m 30s"},
		{time.Minute * 5, "5m 00s"},
		{time.Hour * 5, "05:00h"},
		{time.Hour * 24, "1d 00:00h"},
		{time.Hour * 200, "8d 08:00h"},
		{time.Hour * 800, "4w 5d"},
		{time.Hour * 12345, "1y 21w"},
		{time.Millisecond * -312, "-312ms"},
		{time.Nanosecond * -942, "-942ns"},
	}

	for _, tst := range tests {
		res := DurationString(tst.in)
		assert.Equalf(t, tst.res, res, "DurationString: %v -> %v", tst.in, res)
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		in  string
		res string
		err bool
	}{
		{`"test"`, `test`, false},
		{`'test'`, `test`, false},
		{`'test test'`, `test test`, false},
		{`"test test"`, `test test`, false},
		{`"test test`, "", true},
		{`'test test`, "", true},
		{`test"test`, `test"test`, false},
		{`test'test`, `test'test`, false},
		{`test test'`, "", true},
		{`test test"`, "", true},
	}

	for _, tst := range tests {
		res, err := TrimQuotes(tst.in)
		switch tst.err {
		case true:
			require.Errorf(t, err, "TrimQuotes should error on %s", tst.in)
		case false:
			require.NoErrorf(t, err, "TrimQuotes should not error on %s", tst.in)
			assert.Equalf(t, tst.res, res, "TrimQuotes: %v -> %v", tst.in, res)
		}
	}
}

func TestRankedSort(t *testing.T) {
	keys := []string{
		"/includes",
		"/settings/a",
		"/settings/b",
		"/settings/a/2",
		"/settings/b/1",
		"/settings/default",
		"/paths",
		"/modules",
	}
	expected := []string{
		"/paths",
		"/modules",
		"/settings/default",
		"/settings/a",
		"/settings/a/2",
		"/settings/b",
		"/settings/b/1",
		"/includes",
	}
	ranks := map[string]int{
		"/paths":            1,
		"/modules":          5,
		"/settings/default": 10,
		"/settings":         15,
		"default":           20,
		"/includes":         50,
	}

	sorted := SortRanked(keys, ranks)

	assert.Equalf(t, expected, sorted, "sorted by rank")
}
