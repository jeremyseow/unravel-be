package usecase

import (
	"testing"

	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineBump(t *testing.T) {
	tests := []struct {
		name     string
		old      []domain.SchemaParameter
		next     []domain.SchemaParameter
		expected string
	}{
		{
			name:     "no structural change",
			old:      []domain.SchemaParameter{{ParameterKey: "a", IsRequired: false}},
			next:     []domain.SchemaParameter{{ParameterKey: "a", IsRequired: false}},
			expected: "patch",
		},
		{
			name:     "parameter removed",
			old:      []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b"}},
			next:     []domain.SchemaParameter{{ParameterKey: "a"}},
			expected: "major",
		},
		{
			name:     "new required parameter",
			old:      []domain.SchemaParameter{{ParameterKey: "a"}},
			next:     []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b", IsRequired: true}},
			expected: "major",
		},
		{
			name:     "optional to required",
			old:      []domain.SchemaParameter{{ParameterKey: "a", IsRequired: false}},
			next:     []domain.SchemaParameter{{ParameterKey: "a", IsRequired: true}},
			expected: "major",
		},
		{
			name:     "new optional parameter",
			old:      []domain.SchemaParameter{{ParameterKey: "a"}},
			next:     []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b", IsRequired: false}},
			expected: "minor",
		},
		{
			name:     "required to optional",
			old:      []domain.SchemaParameter{{ParameterKey: "a", IsRequired: true}},
			next:     []domain.SchemaParameter{{ParameterKey: "a", IsRequired: false}},
			expected: "minor",
		},
		{
			name:     "major beats minor — remove param and add optional",
			old:      []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b"}},
			next:     []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "c", IsRequired: false}},
			expected: "major",
		},
		{
			name:     "only additive optional changes",
			old:      []domain.SchemaParameter{{ParameterKey: "a", IsRequired: true}},
			next:     []domain.SchemaParameter{{ParameterKey: "a", IsRequired: true}, {ParameterKey: "b", IsRequired: false}},
			expected: "minor",
		},
		{
			name:     "empty to empty",
			old:      []domain.SchemaParameter{},
			next:     []domain.SchemaParameter{},
			expected: "patch",
		},
		{
			name:     "first parameter added as optional",
			old:      []domain.SchemaParameter{},
			next:     []domain.SchemaParameter{{ParameterKey: "a", IsRequired: false}},
			expected: "minor",
		},
		{
			name:     "first parameter added as required",
			old:      []domain.SchemaParameter{},
			next:     []domain.SchemaParameter{{ParameterKey: "a", IsRequired: true}},
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
	tests := []struct {
		name     string
		latest   *domain.Schema
		draft    domain.Schema
		expected string
	}{
		{
			name:     "first publish — no active version",
			latest:   nil,
			draft:    domain.Schema{Parameters: []domain.SchemaParameter{{ParameterKey: "a"}}},
			expected: "1.0.0",
		},
		{
			name: "patch — no param changes",
			latest: &domain.Schema{
				SchemaVersion: "1.0.0",
				Parameters:    []domain.SchemaParameter{{ParameterKey: "a"}},
			},
			draft:    domain.Schema{Parameters: []domain.SchemaParameter{{ParameterKey: "a"}}},
			expected: "1.0.1",
		},
		{
			name: "minor — new optional param",
			latest: &domain.Schema{
				SchemaVersion: "1.0.0",
				Parameters:    []domain.SchemaParameter{{ParameterKey: "a"}},
			},
			draft:    domain.Schema{Parameters: []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b", IsRequired: false}}},
			expected: "1.1.0",
		},
		{
			name: "major — param removed",
			latest: &domain.Schema{
				SchemaVersion: "1.1.0",
				Parameters:    []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b"}},
			},
			draft:    domain.Schema{Parameters: []domain.SchemaParameter{{ParameterKey: "a"}}},
			expected: "2.0.0",
		},
		{
			name: "major — new required param",
			latest: &domain.Schema{
				SchemaVersion: "2.0.0",
				Parameters:    []domain.SchemaParameter{{ParameterKey: "a"}},
			},
			draft:    domain.Schema{Parameters: []domain.SchemaParameter{{ParameterKey: "a"}, {ParameterKey: "b", IsRequired: true}}},
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
		name      string
		input     string
		major     int
		minor     int
		patch     int
		expectErr bool
	}{
		{"valid", "1.2.3", 1, 2, 3, false},
		{"zeros", "0.0.0", 0, 0, 0, false},
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

func TestAnnotateChanges(t *testing.T) {
	paramA := domain.SchemaParameter{ParameterKey: "a", IsRequired: true}
	paramB := domain.SchemaParameter{ParameterKey: "b", IsRequired: false}
	paramC := domain.SchemaParameter{ParameterKey: "c", IsRequired: true}

	tests := []struct {
		name    string
		input   []domain.Schema
		asserts func(t *testing.T, got []domain.Schema)
	}{
		{
			name:    "single version — no changes",
			input:   []domain.Schema{{SchemaVersion: "1.0.0", Parameters: []domain.SchemaParameter{paramA}}},
			asserts: func(t *testing.T, got []domain.Schema) { assert.Nil(t, got[0].Changes) },
		},
		{
			name: "two published versions — second has changes",
			input: []domain.Schema{
				{SchemaVersion: "1.0.0", Parameters: []domain.SchemaParameter{paramA}},
				{SchemaVersion: "1.1.0", Parameters: []domain.SchemaParameter{paramA, paramB}},
			},
			asserts: func(t *testing.T, got []domain.Schema) {
				require.Len(t, got, 2)
				assert.Nil(t, got[0].Changes)
				require.NotNil(t, got[1].Changes)
				assert.Equal(t, "minor", got[1].Changes.BumpType)
				assert.Equal(t, []string{"b"}, got[1].Changes.AddedParams)
			},
		},
		{
			name: "out-of-order input — sorted before annotating",
			input: []domain.Schema{
				{SchemaVersion: "2.0.0", Parameters: []domain.SchemaParameter{paramA, paramC}},
				{SchemaVersion: "1.0.0", Parameters: []domain.SchemaParameter{paramA}},
			},
			asserts: func(t *testing.T, got []domain.Schema) {
				require.Len(t, got, 2)
				assert.Equal(t, "1.0.0", got[0].SchemaVersion)
				assert.Nil(t, got[0].Changes)
				assert.Equal(t, "2.0.0", got[1].SchemaVersion)
				require.NotNil(t, got[1].Changes)
				assert.Equal(t, "major", got[1].Changes.BumpType)
			},
		},
		{
			name: "draft included — appears last with changes vs latest published",
			input: []domain.Schema{
				{SchemaVersion: "1.0.0", Parameters: []domain.SchemaParameter{paramA}},
				{SchemaVersion: domain.DraftVersion, Parameters: []domain.SchemaParameter{paramA, paramB}},
			},
			asserts: func(t *testing.T, got []domain.Schema) {
				require.Len(t, got, 2)
				assert.Equal(t, "1.0.0", got[0].SchemaVersion)
				assert.Equal(t, domain.DraftVersion, got[1].SchemaVersion)
				require.NotNil(t, got[1].Changes)
				assert.Equal(t, "minor", got[1].Changes.BumpType)
				assert.Equal(t, []string{"b"}, got[1].Changes.AddedParams)
			},
		},
		{
			name: "removed parameter shows in removed_params",
			input: []domain.Schema{
				{SchemaVersion: "1.0.0", Parameters: []domain.SchemaParameter{paramA, paramB}},
				{SchemaVersion: "2.0.0", Parameters: []domain.SchemaParameter{paramA}},
			},
			asserts: func(t *testing.T, got []domain.Schema) {
				require.NotNil(t, got[1].Changes)
				assert.Equal(t, "major", got[1].Changes.BumpType)
				assert.Equal(t, []string{"b"}, got[1].Changes.RemovedParams)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.asserts(t, annotateChanges(tt.input))
		})
	}
}
