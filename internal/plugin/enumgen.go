package plugin

import (
	"strconv"

	"go.einride.tech/protoc-gen-typescript-http/internal/codegen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type enumGenerator struct {
	pkg  protoreflect.FullName
	enum protoreflect.EnumDescriptor
}

func (e enumGenerator) Generate(f *codegen.File) {
	commentGenerator{descriptor: e.enum}.generateLeading(f, 0)
	name := scopedDescriptorTypeName(e.pkg, e.enum)
	f.P("export const ", name, " = {")
	rangeEnumValues(e.enum, func(value protoreflect.EnumValueDescriptor, last bool) {
		commentGenerator{descriptor: value}.generateLeading(f, 1)
		f.P(t(1), string(value.Name()), ": ", strconv.Quote(string(value.Name())), ",")
	})
	f.P("} as const;")
	f.P("export type ", name, " = typeof ", name, "[keyof typeof ", name, "];")
	f.P()
}
