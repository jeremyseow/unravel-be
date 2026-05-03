package usecase

import (
	"testing"

	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func params(keys ...string) []domain.SchemaParameter {
	p := make([]domain.SchemaParameter, len(keys))
	for i, k := range keys {
		required := k[0] == '!'
		if required {
			k = k[1:]
		}
		p[i] = domain.SchemaParameter{ParameterKey: k, IsRequired: required}
	}
	return p
}

func TestDetermineBump(t *testing.T) {
	tests := []struct {
		name     string
		old      []domain.SchemaParameter
		next     []domain.SchemaParameter
		expected string
	}{
		{
			name:     "no structural change",
			old:      params("a"),
			next:     params("a"),
			expected: "patch",
		},
		{
			name:     "parameter removed",
			old:      params("a", "b"),
			next:     params("a"),
			expected: "major",
		},
		{
			name:     "new required parameter",
			old:      params("a"),
			next:     params("a", "!b"),
			expected: "major",
		},
		{
			name:     "optional to required",
			old:      params("a"),
			next:     params("!a"),
			expected: "major",
		},
		{
			name:     "new optional parameter",
			old:      params("a"),
			next:     params("a", "b"),
			expected: "minor",
		},
		{
			name:     "required to optional",
			old:      params("!a"),
			next:     params("a"),
			expected: "minor",
		},
		{
			name:     "major beats minor — remove param and add optional",
			old:      params("a", "b"),
			next:     params("a", "c"),
			expected: "major",
		},
		{
			name:     "only additive optional changes",
			old:      params("!a"),
			next:     params("!a", "b"),
			expected: "minor",
		},
		{
			name:     "empty to empty",
			old:      params(),
			next:     params(),
			expected: "patch",
		},
		{
			name:     "first parameter added as optional",
			old:      params(),
			next:     params("a"),
			expected: "minor",
		},
		{
			name:     "first parameter added as required",
			old:      params(),
			next:     params("!a"),
			expected: "major",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, determineBump(tt.old, tt.next))
		})
	}
}

func TestApplyBump(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		bump     string
		expected string
	}{
		{"major bump", "1.2.3", "major", "2.0.0"},
		{"minor bump", "1.2.3", "minor", "1.3.0"},
		{"patch bump", "1.2.3", "patch", "1.2.4"},
		{"major resets minor and patch", "2.9.9", "major", "3.0.0"},
		{"minor resets patch", "1.9.9", "minor", "1.10.0"},
		{"pre-release stripped before bump", "1.2.3-rc.1", "patch", "1.2.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyBump(tt.version, tt.bump)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestComputeNextVersion(t *testing.T) {
	active := func(version string, p ...string) *domain.Schema {
		return &domain.Schema{SchemaVersion: version, Parameters: params(p...)}
	}

	tests := []struct {
		name     string
		latest   *domain.Schema
		draft    domain.Schema
		expected string
	}{
		{
			name:     "first publish — no active version",
			latest:   nil,
			draft:    domain.Schema{Parameters: params("a")},
			expected: "1.0.0",
		},
		{
			name:     "patch — no param changes",
			latest:   active("1.0.0", "a"),
			draft:    domain.Schema{Parameters: params("a")},
			expected: "1.0.1",
		},
		{
			name:     "minor — new optional param",
			latest:   active("1.0.0", "a"),
			draft:    domain.Schema{Parameters: params("a", "b")},
			expected: "1.1.0",
		},
		{
			name:     "major — param removed",
			latest:   active("1.1.0", "a", "b"),
			draft:    domain.Schema{Parameters: params("a")},
			expected: "2.0.0",
		},
		{
			name:     "major — new required param",
			latest:   active("2.0.0", "a"),
			draft:    domain.Schema{Parameters: params("a", "!b")},
			expected: "3.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeNextVersion(tt.latest, tt.draft)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		major, minor  int
		patch         int
		expectErr     bool
	}{
		{"valid", "1.2.3", 1, 2, 3, false},
		{"zeros", "0.0.0", 0, 0, 0, false},
		{"large numbers", "10.20.30", 10, 20, 30, false},
		{"pre-release stripped", "1.2.3-rc.1", 1, 2, 3, false},
		{"invalid — missing part", "1.2", 0, 0, 0, true},
		{"invalid — non-numeric", "a.b.c", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := parseSemver(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.major, major)
			assert.Equal(t, tt.minor, minor)
			assert.Equal(t, tt.patch, patch)
		})
	}
}
