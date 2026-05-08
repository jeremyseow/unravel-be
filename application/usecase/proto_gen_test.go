package usecase

import (
	"strings"
	"testing"

	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestGenerateProtoText(t *testing.T) {
	tests := []struct {
		name      string
		schemaKey string
		params    []domain.EnrichedSchemaParameter
		contains  []string
		absent    []string
	}{
		{
			name:      "basic scalars",
			schemaKey: "user_clicked_button",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "user_id", DataType: domain.DataTypeString, IsRequired: true},
				{ParameterKey: "click_count", DataType: domain.DataTypeInteger},
			},
			contains: []string{
				`syntax = "proto3";`,
				"package unravel;",
				"message UserClickedButton {",
				"  string user_id = 1;",
				"  int64 click_count = 2;",
				"// user_id (required)",
			},
			absent: []string{`import "google/protobuf/struct.proto";`},
		},
		{
			name:      "json type adds struct import",
			schemaKey: "my_event",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "metadata", DataType: domain.DataTypeJSON},
			},
			contains: []string{
				`import "google/protobuf/struct.proto";`,
				"message MyEvent {",
				"  google.protobuf.Struct metadata = 1;",
			},
		},
		{
			name:      "repeated types",
			schemaKey: "batch_event",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "tags", DataType: domain.DataTypeStringArray},
				{ParameterKey: "counts", DataType: domain.DataTypeIntegerArray},
			},
			contains: []string{
				"  repeated string tags = 1;",
				"  repeated int64 counts = 2;",
			},
		},
		{
			name:      "description embedded in comment",
			schemaKey: "order_placed",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "amount", DataType: domain.DataTypeFloat, Description: strPtr("Total order amount in USD")},
			},
			contains: []string{"// amount — Total order amount in USD"},
		},
		{
			name:      "empty parameters — empty message body",
			schemaKey: "empty_schema",
			params:    []domain.EnrichedSchemaParameter{},
			contains:  []string{"message EmptySchema {\n}"},
		},
		{
			name:      "kebab-case key becomes PascalCase",
			schemaKey: "page-view",
			params:    []domain.EnrichedSchemaParameter{},
			contains:  []string{"message PageView {"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateProtoText(tt.schemaKey, tt.params)
			for _, want := range tt.contains {
				assert.True(t, strings.Contains(got, want), "expected %q in output:\n%s", want, got)
			}
			for _, bad := range tt.absent {
				assert.False(t, strings.Contains(got, bad), "did not expect %q in output:\n%s", bad, got)
			}
		})
	}
}

func TestGenerateFileDescriptor(t *testing.T) {
	tests := []struct {
		name      string
		schemaKey string
		params    []domain.EnrichedSchemaParameter
	}{
		{
			name:      "scalars produce valid bytes",
			schemaKey: "user_event",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "user_id", DataType: domain.DataTypeString},
				{ParameterKey: "ts", DataType: domain.DataTypeTimestamp},
				{ParameterKey: "active", DataType: domain.DataTypeBoolean},
			},
		},
		{
			name:      "json type included",
			schemaKey: "rich_event",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "payload", DataType: domain.DataTypeJSON},
			},
		},
		{
			name:      "repeated types included",
			schemaKey: "multi_event",
			params: []domain.EnrichedSchemaParameter{
				{ParameterKey: "tags", DataType: domain.DataTypeStringArray},
				{ParameterKey: "scores", DataType: domain.DataTypeFloatArray},
			},
		},
		{
			name:      "empty params",
			schemaKey: "empty",
			params:    []domain.EnrichedSchemaParameter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFileDescriptor(tt.schemaKey, tt.params)
			require.NoError(t, err)
			assert.NotEmpty(t, got)
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_clicked_button", "UserClickedButton"},
		{"page-view", "PageView"},
		{"single", "Single"},
		{"already_PascalCase", "AlreadyPascalCase"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toPascalCase(tt.input))
		})
	}
}
