package plugin

import (
	"fmt"
	"path"
	"strings"

	"github.com/cortea-ai/protoc-gen-pydantic/internal/codegen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func Generate(request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	generate := make(map[string]struct{})
	registry, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: request.GetProtoFile(),
	})
	if err != nil {
		return nil, fmt.Errorf("create proto registry: %w", err)
	}
	for _, f := range request.GetFileToGenerate() {
		generate[f] = struct{}{}
	}
	packaged := make(map[protoreflect.FullName][]protoreflect.FileDescriptor)
	for _, f := range request.GetFileToGenerate() {
		file, err := registry.FindFileByPath(f)
		if err != nil {
			return nil, fmt.Errorf("find file %s: %w", f, err)
		}
		packaged[file.Package()] = append(packaged[file.Package()], file)
	}

	params := parseParameters(request.GetParameter())

	packageSuffix := params["package_suffix"]

	var filename string
	var ok bool
	if filename, ok = params["filename"]; !ok {
		filename = "pb_models"
	}

	var res pluginpb.CodeGeneratorResponse
	for pkg, files := range packaged {
		if includePath, ok := params["include_path"]; ok {
			if !strings.HasPrefix(string(pkg), includePath) {
				continue
			}
		}
		var index codegen.File
		indexPathElems := append(strings.Split(string(pkg)+packageSuffix, "."), filename+".py")
		(packageGenerator{pkg: pkg, files: files, params: params}).Generate(&index)
		res.File = append(res.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(path.Join(indexPathElems...)),
			Content: proto.String(string(index.Content())),
		})
		indexPathElems = append(strings.Split(string(pkg)+packageSuffix, "."), "__init__.py")
		res.File = append(res.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(path.Join(indexPathElems...)),
			Content: proto.String("from ." + filename + " import *\n"),
		})
		indexPathElems = append(strings.Split(string(pkg)+packageSuffix, "."), "py.typed")
		res.File = append(res.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(path.Join(indexPathElems...)),
			Content: proto.String(""),
		})
	}
	res.SupportedFeatures = proto.Uint64(uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))
	return &res, nil
}

func parseParameters(parameter string) map[string]string {
	params := make(map[string]string)
	for _, param := range strings.Split(parameter, ",") {
		if param == "" {
			continue
		}
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 1 {
			params[parts[0]] = ""
		} else {
			params[parts[0]] = parts[1]
		}
	}
	return params
}
