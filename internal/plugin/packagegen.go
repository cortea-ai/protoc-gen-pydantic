package plugin

import (
	"strings"

	"github.com/cortea-ai/protoc-gen-pydantic/internal/codegen"
	"github.com/cortea-ai/protoc-gen-pydantic/internal/protowalk"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type packageGenerator struct {
	pkg    protoreflect.FullName
	files  []protoreflect.FileDescriptor
	params map[string]string
}

type descNode struct {
	name      string
	generator descriptorGenerator
	children  []*descNode
}

func (p packageGenerator) Generate(f *codegen.File) {
	p.generateHeader(f)

	root := &descNode{name: "root", children: []*descNode{}}
	current := root

	protowalk.WalkFiles(p.files, func(desc protoreflect.Descriptor) bool {
		switch t := desc.(type) {
		case protoreflect.MessageDescriptor:
			if t.IsMapEntry() {
				return false
			}
		case protoreflect.EnumDescriptor:
		default:
			return true
		}

		id := scopedDescriptorTypeName(p.pkg, desc)
		parts := strings.Split(id, "_")
		current = root

		var g descriptorGenerator
		for i, part := range parts {
			var child *descNode
			for _, c := range current.children {
				if c.name == part {
					child = c
					break
				}
			}
			isLeaf := i == len(parts)-1
			if isLeaf {
				g = descriptorGenerator{
					name:   part,
					pkg:    p.pkg,
					desc:   desc,
					indent: (len(parts) - 1) * 2,
				}
				if child != nil {
					// if node was already created as non-leaf the generator
					// currently set is nil
					child.generator = g
				} else {
					child = &descNode{name: part, generator: g}
					current.children = append(current.children, child)
				}
			} else if child == nil {
				child = &descNode{name: part}
				current.children = append(current.children, child)
			}
			current = child
		}

		if len(parts) > 1 {
			return true
		}

		current.generator.GenerateHeader(f)

		var visitChildren func(node *descNode)
		visitChildren = func(node *descNode) {
			for _, child := range node.children {
				child.generator.GenerateHeader(f)
				visitChildren(child)
				child.generator.GenerateFields(f)
			}
		}
		visitChildren(current)

		current.generator.GenerateFields(f)
		f.P()

		return true
	})
}

func (p packageGenerator) generateHeader(f *codegen.File) {
	f.P("####################################################################")
	f.P("### This is an automatically generated file.        DO NOT EDIT  ###")
	f.P("####################################################################")
	f.P()
	f.P("import datetime")
	f.P("import json")
	f.P()
	f.P("from enum import StrEnum")
	if p.params["pydantic_base_path"] != "" {
		f.P("from ", p.params["pydantic_base_path"], " import BaseModel")
		f.P("from pydantic import Field, field_serializer, model_validator, SerializationInfo")
	} else {
		f.P("from pydantic import BaseModel, Field, field_serializer, model_validator, SerializationInfo")
	}
	f.P("from typing import Optional, Self")
	f.P("from uuid import UUID")
	f.P()
}
