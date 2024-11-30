package plugin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cortea-ai/protoc-gen-pydantic/internal/codegen"
	"github.com/cortea-ai/protoc-gen-pydantic/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type descriptorGenerator struct {
	name   string
	pkg    protoreflect.FullName
	desc   protoreflect.Descriptor
	indent int
}

func (d descriptorGenerator) GenerateHeader(f *codegen.File) {
	switch d.desc.(type) {
	case protoreflect.EnumDescriptor:
		d.generateEnumHeader(f)
	case protoreflect.MessageDescriptor:
		d.generateMessageHeader(f)
	}
}

func (d descriptorGenerator) GenerateFields(f *codegen.File) {
	switch t := d.desc.(type) {
	case protoreflect.EnumDescriptor:
		d.generateEnumFields(f, t)
	case protoreflect.MessageDescriptor:
		d.generateMessageFields(f, t)
	}

	f.P()
}

func (d descriptorGenerator) generateEnumHeader(f *codegen.File) {
	f.P(t(d.indent), "class ", d.name, "(StrEnum):")
}

func (d descriptorGenerator) generateMessageHeader(f *codegen.File) {
	f.P(t(d.indent), "class ", d.name, "(BaseModel):")
}

func (d descriptorGenerator) generateEnumFields(f *codegen.File, enum protoreflect.EnumDescriptor) {
	rangeEnumValues(enum, func(value protoreflect.EnumValueDescriptor, last bool) {
		commentGenerator{descriptor: value}.generateLeading(f, d.indent+2)
		f.P(t(d.indent+2), string(value.Name()), " = ", strconv.Quote(string(value.Name())))
	})
}

func (d descriptorGenerator) generateMessageFields(f *codegen.File, message protoreflect.MessageDescriptor) {
	if IsWellKnownType(message) {
		return
	}

	mapFields := make([]string, 0)
	oneofFields := make([]string, 0)
	rangeFields(message, func(field protoreflect.FieldDescriptor) {
		defaultValue := ""
		defaultFactory := ""
		isUUID := false
		opts := []string{}

		fieldType := typeFromField(d.pkg, field)

		commentGenerator{descriptor: field}.generateLeading(f, d.indent+2)

		rules := proto.GetExtension(field.Options(), validate.E_Rules)
		switch r := rules.(type) {
		case *validate.FieldRules:
			if r.GetFloat() != nil {
				if hasPresence(r.GetFloat().ProtoReflect(), "default") {
					defaultValue = fmt.Sprintf("default=%.1f", r.GetFloat().GetDefault())
				}
				if hasPresence(r.GetFloat().ProtoReflect(), "lt") {
					opts = append(opts, fmt.Sprintf("lt=%.1f", r.GetFloat().GetLt()))
				}
				if hasPresence(r.GetFloat().ProtoReflect(), "gt") {
					opts = append(opts, fmt.Sprintf("gt=%.1f", r.GetFloat().GetGt()))
				}
				if hasPresence(r.GetFloat().ProtoReflect(), "gte") {
					opts = append(opts, fmt.Sprintf("gte=%.1f", r.GetFloat().GetGte()))
				}
				if hasPresence(r.GetFloat().ProtoReflect(), "lte") {
					opts = append(opts, fmt.Sprintf("lte=%.1f", r.GetFloat().GetLte()))
				}
			}
			if r.GetInt32() != nil {
				if hasPresence(r.GetInt32().ProtoReflect(), "default") {
					defaultValue = fmt.Sprintf("default=%d", r.GetInt32().GetDefault())
				}
				if hasPresence(r.GetInt32().ProtoReflect(), "lt") {
					opts = append(opts, fmt.Sprintf("lt=%d", r.GetInt32().GetLt()))
				}
				if hasPresence(r.GetInt32().ProtoReflect(), "gt") {
					opts = append(opts, fmt.Sprintf("gt=%d", r.GetInt32().GetGt()))
				}
				if hasPresence(r.GetInt32().ProtoReflect(), "gte") {
					opts = append(opts, fmt.Sprintf("gte=%d", r.GetInt32().GetGte()))
				}
				if hasPresence(r.GetInt32().ProtoReflect(), "lte") {
					opts = append(opts, fmt.Sprintf("lte=%d", r.GetInt32().GetLte()))
				}
			}
			if r.GetString_() != nil {
				if hasPresence(r.GetString_().ProtoReflect(), "default") {
					defaultValue = fmt.Sprintf(`default="%s"`, r.GetString_().GetDefault())
				}
				if hasPresence(r.GetString_().ProtoReflect(), "len") {
					opts = append(opts, fmt.Sprintf("len=%d", r.GetString_().GetLen()))
				}
				if hasPresence(r.GetString_().ProtoReflect(), "min_length") {
					opts = append(opts, fmt.Sprintf("min_length=%d", r.GetString_().GetMinLength()))
				}
				if hasPresence(r.GetString_().ProtoReflect(), "max_length") {
					opts = append(opts, fmt.Sprintf("max_length=%d", r.GetString_().GetMaxLength()))
				}
				isUUID = r.GetString_().GetUuid()
			}
			if r.GetMessage() != nil {
				if hasPresence(r.GetMessage().ProtoReflect(), "default_factory") {
					defaultFactory = fmt.Sprintf(`default_factory="%s"`, r.GetMessage().GetDefaultFactory())
				}
				if hasPresence(r.GetMessage().ProtoReflect(), "default_empty") && r.GetMessage().GetDefaultEmpty() {
					defaultValue = fmt.Sprintf(`default_factory=%s`, fieldType.Reference(isUUID))
				}
			}
			if r.GetRepeated() != nil {
				if hasPresence(r.GetRepeated().ProtoReflect(), "len") {
					opts = append(opts, fmt.Sprintf("len=%d", r.GetRepeated().GetLen()))
				}
				if hasPresence(r.GetRepeated().ProtoReflect(), "min_length") {
					opts = append(opts, fmt.Sprintf("min_length=%d", r.GetRepeated().GetMinLength()))
				}
				if hasPresence(r.GetRepeated().ProtoReflect(), "max_length") {
					opts = append(opts, fmt.Sprintf("max_length=%d", r.GetRepeated().GetMaxLength()))
				}
				isUUID = r.GetRepeated().GetItems().GetString_().GetUuid()
			}
		}

		isOneOf := field.ContainingOneof() != nil && !field.HasOptionalKeyword()
		if isOneOf {
			oneofFields = append(oneofFields, "self."+string(field.Name()))
		}
		isOptional := field.HasOptionalKeyword() || isOneOf
		fmtOpts := strings.Join(opts, ", ")
		extras := getExtras(field, fmtOpts, defaultValue, defaultFactory)

		if isOptional && defaultValue == "" && defaultFactory == "" {
			f.P(t(d.indent+2), field.Name(), ": Optional[", fieldType.Reference(isUUID), "] = Field(", extras, ")")
		} else if field.IsList() {
			f.P(t(d.indent+2), field.Name(), ": ", fieldType.Reference(isUUID), " = Field(", extras, ")")
		} else if field.IsMap() {
			f.P(t(d.indent+2), field.Name(), ": ", fieldType.Reference(isUUID), " = Field(", extras, ")")
			mapFields = append(mapFields, string(field.Name()))
		} else {
			f.P(t(d.indent+2), field.Name(), ": ", fieldType.Reference(isUUID), " = Field(", extras, ")")
		}
	})

	if len(mapFields) > 0 {
		f.P("")
		f.P(t(d.indent+2), "@field_serializer(")
	}
	for _, field := range mapFields {
		f.P(t(d.indent+4), `"`, field, `",`)
	}
	if len(mapFields) > 0 {
		f.P(t(d.indent+2), ")")
		f.P(t(d.indent+2), "def json_dump(self, v: dict, info: SerializationInfo):")
		f.P(t(d.indent+4), "if info.context == 'bigquery':")
		f.P(t(d.indent+6), "return json.dumps(v)")
		f.P(t(d.indent+4), "return v")
	}

	if len(oneofFields) > 0 {
		f.P("")
		f.P(t(d.indent+2), `@model_validator(mode="before")`)
	}
	if len(oneofFields) > 0 {
		f.P(t(d.indent+2), "def validate_one_ofs(self) -> Self:")
		f.P(t(d.indent+4), `assert sum(x is not None for x in [`+strings.Join(oneofFields, ", ")+`]) == 1, \`)
		f.P(t(d.indent+6), `ValueError("OneOf condition not met")`)
		f.P(t(d.indent+4), "return self")
	}
}

func getExtras(field protoreflect.FieldDescriptor, fmtOpts, defaultValue, defaultFactory string) string {
	if field.HasOptionalKeyword() && defaultValue == "" {
		defaultValue = "default=None"
	}
	if field.ContainingOneof() != nil && !field.HasOptionalKeyword() {
		defaultValue = "default=None"
	}
	if field.IsList() && defaultFactory == "" {
		defaultFactory = "default_factory=list"
	}
	if field.IsMap() && defaultFactory == "" {
		defaultFactory = "default_factory=dict"
	}
	var extras []string
	if fmtOpts != "" {
		extras = append(extras, fmtOpts)
	}
	if defaultValue != "" {
		extras = append(extras, defaultValue)
	}
	if defaultFactory != "" {
		extras = append(extras, defaultFactory)
	}
	return strings.Join(extras, ", ")
}

func hasPresence(msg protoreflect.Message, field string) bool {
	nameField := msg.Descriptor().Fields().ByName(protoreflect.Name(field))
	return msg.Has(nameField)
}
