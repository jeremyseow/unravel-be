package schema

import (
	"testing"

	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	"github.com/stretchr/testify/assert"
)

// boolPtr is a test helper to get a pointer to a bool literal.
func boolPtr(b bool) *bool { return &b }

// makeMapping builds a minimal EntitySchemasParametersMappings slice for test input.
func makeMapping(params map[string]bool) []model.EntitySchemasParametersMappings {
	out := make([]model.EntitySchemasParametersMappings, 0, len(params))
	for key, required := range params {
		req := required
		out = append(out, model.EntitySchemasParametersMappings{
			ParameterKey: key,
			IsRequired:   &req,
		})
	}
	return out
}

// --- determineVersionBump ---

func TestDetermineVersionBump(t *testing.T) {
	opt := func(key string) (string, paramInfo) { return key, paramInfo{isRequired: false} }
	req := func(key string) (string, paramInfo) { return key, paramInfo{isRequired: true} }
	pm := func(pairs ...func() (string, paramInfo)) map[string]paramInfo {
		m := make(map[string]paramInfo, len(pairs))
		for _, p := range pairs {
			k, v := p()
			m[k] = v
		}
		return m
	}

	tests := []struct {
		name     string
		current  map[string]paramInfo
		next     map[string]paramInfo
		expected string
	}{
		{
			name:     "no structural change — same optional param",
			current:  pm(func() (string, paramInfo) { return opt("a") }),
			next:     pm(func() (string, paramInfo) { return opt("a") }),
			expected: "patch",
		},
		{
			name:     "no structural change — same required param",
			current:  pm(func() (string, paramInfo) { return req("a") }),
			next:     pm(func() (string, paramInfo) { return req("a") }),
			expected: "patch",
		},
		{
			name:     "no structural change — empty both sides",
			current:  map[string]paramInfo{},
			next:     map[string]paramInfo{},
			expected: "patch",
		},
		{
			name:     "parameter removed",
			current:  pm(func() (string, paramInfo) { return opt("a") }),
			next:     map[string]paramInfo{},
			expected: "major",
		},
		{
			name:     "required parameter removed",
			current:  pm(func() (string, paramInfo) { return req("a") }),
			next:     map[string]paramInfo{},
			expected: "major",
		},
		{
			name:     "new required parameter added",
			current:  map[string]paramInfo{},
			next:     pm(func() (string, paramInfo) { return req("a") }),
			expected: "major",
		},
		{
			name:     "new optional parameter added",
			current:  map[string]paramInfo{},
			next:     pm(func() (string, paramInfo) { return opt("a") }),
			expected: "minor",
		},
		{
			name:     "optional → required (tightening)",
			current:  pm(func() (string, paramInfo) { return opt("a") }),
			next:     pm(func() (string, paramInfo) { return req("a") }),
			expected: "major",
		},
		{
			name:     "required → optional (relaxation)",
			current:  pm(func() (string, paramInfo) { return req("a") }),
			next:     pm(func() (string, paramInfo) { return opt("a") }),
			expected: "minor",
		},
		{
			name: "mixed changes — major wins over minor",
			current: pm(
				func() (string, paramInfo) { return opt("a") },
			),
			next: pm(
				func() (string, paramInfo) { return req("a") }, // optional→required = major
				func() (string, paramInfo) { return opt("b") }, // new optional = minor
			),
			expected: "major",
		},
		{
			name: "only additive changes — no existing param changed",
			current: pm(
				func() (string, paramInfo) { return req("a") },
			),
			next: pm(
				func() (string, paramInfo) { return req("a") },
				func() (string, paramInfo) { return opt("b") }, // new optional = minor
			),
			expected: "minor",
		},
		{
			name: "multiple parameters removed — still major",
			current: pm(
				func() (string, paramInfo) { return opt("a") },
				func() (string, paramInfo) { return opt("b") },
			),
			next: map[string]paramInfo{},
			expected: "major",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := determineVersionBump(tc.current, tc.next)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// --- applyBump ---

func TestApplyBump(t *testing.T) {
	tests := []struct {
		version  string
		bump     string
		expected string
	}{
		{"1.2.3", "major", "2.0.0"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "patch", "1.2.4"},
		{"1.0.0", "major", "2.0.0"},
		{"2.5.1", "minor", "2.6.0"},
		{"0.0.0", "patch", "0.0.1"},
		{"10.20.30", "major", "11.0.0"},
	}

	for _, tc := range tests {
		t.Run(tc.version+"+"+tc.bump, func(t *testing.T) {
			got := applyBump(tc.version, tc.bump)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// --- compareSemver ---

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"major a > b", "2.0.0", "1.9.9", 1},
		{"major a < b", "1.9.9", "2.0.0", -1},
		{"minor a > b", "1.2.0", "1.1.9", 1},
		{"minor a < b", "1.1.9", "1.2.0", -1},
		{"patch a > b", "1.0.1", "1.0.0", 1},
		{"patch a < b", "1.0.0", "1.0.1", -1},
		{"release beats pre-release", "1.0.0", "1.0.0-rc.1", 1},
		{"pre-release loses to release", "1.0.0-rc.1", "1.0.0", -1},
		{"pre-release ordering rc.2 > rc.1", "1.0.0-rc.2", "1.0.0-rc.1", 1},
		{"pre-release ordering rc.1 < rc.2", "1.0.0-rc.1", "1.0.0-rc.2", -1},
		{"both pre-release equal", "1.0.0-rc.1", "1.0.0-rc.1", 0},
		{"draft version is lowest", "0.0.0", "1.0.0", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := compareSemver(tc.a, tc.b)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// --- sortByVersion ---

func TestSortByVersion(t *testing.T) {
	records := []SchemaRecord{
		{model.EntitySchemas{SchemaVersion: "1.0.0"}, nil},
		{model.EntitySchemas{SchemaVersion: DraftVersion}, nil},
		{model.EntitySchemas{SchemaVersion: "2.0.0"}, nil},
		{model.EntitySchemas{SchemaVersion: "1.1.0"}, nil},
	}

	sortByVersion(records)

	versions := make([]string, len(records))
	for i, r := range records {
		versions[i] = r.SchemaVersion
	}

	assert.Equal(t, []string{"2.0.0", "1.1.0", "1.0.0", DraftVersion}, versions)
}

// --- toParamMap ---

func TestToParamMap(t *testing.T) {
	mappings := makeMapping(map[string]bool{
		"alpha": true,
		"beta":  false,
	})

	m := toParamMap(mappings)

	assert.Len(t, m, 2)
	assert.True(t, m["alpha"].isRequired)
	assert.False(t, m["beta"].isRequired)
}

func TestToParamMap_NilIsRequired(t *testing.T) {
	// A nil IsRequired pointer should be treated as optional (false).
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: "x", IsRequired: nil},
	}

	m := toParamMap(mappings)
	assert.False(t, m["x"].isRequired)
}
