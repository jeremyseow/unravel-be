package usecase

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"

	"github.com/jeremyseow/unravel-be/application/domain"
)

// GenerateProtoText returns a proto3 file as a string for the given schema.
// Parameters are listed in the order provided; callers should sort for stability.
func GenerateProtoText(schemaKey string, params []domain.EnrichedSchemaParameter) string {
	var sb strings.Builder
	messageName := toPascalCase(schemaKey)

	needsStructImport := false
	for _, p := range params {
		if p.DataType == domain.DataTypeJSON {
			needsStructImport = true
			break
		}
	}

	sb.WriteString("syntax = \"proto3\";\n\npackage unravel;\n")
	if needsStructImport {
		sb.WriteString("\nimport \"google/protobuf/struct.proto\";\n")
	}
	sb.WriteString(fmt.Sprintf("\nmessage %s {\n", messageName))

	for i, p := range params {
		comment := buildFieldComment(p)
		sb.WriteString(fmt.Sprintf("  // %s\n", comment))
		protoType := domain.ToProtoType(p.DataType)
		sb.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, p.ParameterKey, i+1))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// GenerateFileDescriptor returns a serialised FileDescriptorProto for the schema.
func GenerateFileDescriptor(schemaKey string, params []domain.EnrichedSchemaParameter) ([]byte, error) {
	messageName := toPascalCase(schemaKey)

	var deps []string
	fields := make([]*descriptorpb.FieldDescriptorProto, len(params))

	for i, p := range params {
		num := int32(i + 1)
		protoType := domain.ToProtoType(p.DataType)
		fieldType, label, typeName := resolveFieldDescriptor(protoType)

		if typeName != nil && *typeName == ".google.protobuf.Struct" && len(deps) == 0 {
			deps = []string{"google/protobuf/struct.proto"}
		}

		name := p.ParameterKey
		fields[i] = &descriptorpb.FieldDescriptorProto{
			Name:     &name,
			Number:   &num,
			Type:     &fieldType,
			Label:    &label,
			TypeName: typeName,
		}
	}

	pkg := "unravel"
	fileName := schemaKey + ".proto"
	fd := &descriptorpb.FileDescriptorProto{
		Name:       &fileName,
		Package:    &pkg,
		Syntax:     proto.String("proto3"),
		Dependency: deps,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name:  &messageName,
				Field: fields,
			},
		},
	}

	return proto.Marshal(fd)
}

// resolveFieldDescriptor maps a ProtoType string (from domain.ProtoType) to
// the corresponding descriptor field type, label, and optional message type name.
func resolveFieldDescriptor(protoType string) (descriptorpb.FieldDescriptorProto_Type, descriptorpb.FieldDescriptorProto_Label, *string) {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	if strings.HasPrefix(protoType, "repeated ") {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		protoType = strings.TrimPrefix(protoType, "repeated ")
	}

	switch protoType {
	case "int64":
		return descriptorpb.FieldDescriptorProto_TYPE_INT64, label, nil
	case "double":
		return descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, label, nil
	case "bool":
		return descriptorpb.FieldDescriptorProto_TYPE_BOOL, label, nil
	case "bytes":
		return descriptorpb.FieldDescriptorProto_TYPE_BYTES, label, nil
	case "google.protobuf.Struct":
		typeName := ".google.protobuf.Struct"
		return descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, label, &typeName
	default: // "string" and unknown fall through to string
		return descriptorpb.FieldDescriptorProto_TYPE_STRING, label, nil
	}
}

func buildFieldComment(p domain.EnrichedSchemaParameter) string {
	var sb strings.Builder
	sb.WriteString(p.ParameterKey)
	if p.IsRequired {
		sb.WriteString(" (required)")
	}
	if p.Description != nil && *p.Description != "" {
		sb.WriteString(" — ")
		sb.WriteString(*p.Description)
	}
	return sb.String()
}

// toPascalCase converts snake_case or kebab-case to PascalCase.
func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
