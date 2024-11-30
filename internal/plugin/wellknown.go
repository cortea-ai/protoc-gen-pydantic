package plugin

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	wellKnownPrefix = "google.protobuf."
)

type WellKnown string

// https://developers.google.com/protocol-buffers/docs/reference/google.protobuf
const (
	WellKnownAny       WellKnown = "google.protobuf.Any"
	WellKnownDuration  WellKnown = "google.protobuf.Duration"
	WellKnownEmpty     WellKnown = "google.protobuf.Empty"
	WellKnownFieldMask WellKnown = "google.protobuf.FieldMask"
	WellKnownStruct    WellKnown = "google.protobuf.Struct"
	WellKnownTimestamp WellKnown = "google.protobuf.Timestamp"

	// Wrapper types.
	WellKnownFloatValue  WellKnown = "google.protobuf.FloatValue"
	WellKnownInt64Value  WellKnown = "google.protobuf.Int64Value"
	WellKnownInt32Value  WellKnown = "google.protobuf.Int32Value"
	WellKnownUInt64Value WellKnown = "google.protobuf.UInt64Value"
	WellKnownUInt32Value WellKnown = "google.protobuf.UInt32Value"
	WellKnownBytesValue  WellKnown = "google.protobuf.BytesValue"
	WellKnownDoubleValue WellKnown = "google.protobuf.DoubleValue"
	WellKnownBoolValue   WellKnown = "google.protobuf.BoolValue"
	WellKnownStringValue WellKnown = "google.protobuf.StringValue"

	// Descriptor types.
	WellKnownValue     WellKnown = "google.protobuf.Value"
	WellKnownNullValue WellKnown = "google.protobuf.NullValue"
	WellKnownListValue WellKnown = "google.protobuf.ListValue"
)

func IsWellKnownType(desc protoreflect.Descriptor) bool {
	switch desc.(type) {
	case protoreflect.MessageDescriptor, protoreflect.EnumDescriptor:
		return strings.HasPrefix(string(desc.FullName()), wellKnownPrefix)
	default:
		return false
	}
}

func WellKnownType(desc protoreflect.Descriptor) (WellKnown, bool) {
	if !IsWellKnownType(desc) {
		return "", false
	}
	return WellKnown(desc.FullName()), true
}

func (wkt WellKnown) Name() string {
	switch wkt {
	case WellKnownTimestamp:
		return "datetime.datetime"
	case WellKnownDuration:
		return "datetime.timedelta"
	default:
		panic(fmt.Sprintf("unknown well known type: %s", wkt))
	}
}
