package plugin

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Type struct {
	IsNamed bool
	Name    string

	IsList     bool
	IsMap      bool
	Underlying *Type
}

func (t Type) Reference(isUUID bool) string {
	switch {
	case t.IsMap:
		return "dict[str, " + t.Underlying.Reference(isUUID) + "]"
	case t.IsList:
		return "list[" + t.Underlying.Reference(isUUID) + "]"
	default:
		if isUUID {
			return "UUID"
		}
		return t.Name
	}
}

func typeFromField(pkg protoreflect.FullName, field protoreflect.FieldDescriptor) Type {
	switch {
	case field.IsMap():
		underlying := namedTypeFromField(pkg, field.MapValue())
		return Type{
			IsMap:      true,
			Underlying: &underlying,
		}
	case field.IsList():
		underlying := namedTypeFromField(pkg, field)
		return Type{
			IsList:     true,
			Underlying: &underlying,
		}
	default:
		return namedTypeFromField(pkg, field)
	}
}

func namedTypeFromField(pkg protoreflect.FullName, field protoreflect.FieldDescriptor) Type {
	switch field.Kind() {
	case protoreflect.StringKind, protoreflect.BytesKind:
		return Type{IsNamed: true, Name: "str"}
	case protoreflect.BoolKind:
		return Type{IsNamed: true, Name: "bool"}
	case
		protoreflect.Int32Kind,
		protoreflect.Int64Kind,
		protoreflect.Uint32Kind,
		protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Sfixed64Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sint64Kind:
		return Type{IsNamed: true, Name: "int"}
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return Type{IsNamed: true, Name: "float"}
	case protoreflect.MessageKind:
		return typeFromMessage(pkg, field.Message())
	case protoreflect.EnumKind:
		desc := field.Enum()
		if wkt, ok := WellKnownType(field.Enum()); ok {
			return Type{IsNamed: true, Name: wkt.Name()}
		}
		return Type{IsNamed: true, Name: string(desc.Name())}
	default:
		panic(fmt.Sprintf("unknown field kind: %s", field.Kind()))
	}
}

func typeFromMessage(pkg protoreflect.FullName, message protoreflect.MessageDescriptor) Type {
	if wkt, ok := WellKnownType(message); ok {
		return Type{IsNamed: true, Name: wkt.Name()}
	}
	return Type{IsNamed: true, Name: string(message.Name())}
}
