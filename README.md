# Protoc-gen-pydantic

Generate Pydantic models from Protobuf definitions.

## Build

Install `protoc-gen-go`:

```sh
env GO111MODULE=on GOBIN=$PWD/bin go install google.golang.org/protobuf/cmd/protoc-gen-go
```

Build the plugin:

```sh
buf generate
```

Regenerate `py_validate.proto` go stubs:

```sh
protoc \
  --go_opt=paths=source_relative \
  --go_out=./ \
  validate/py_validate.proto
```

## Known Limitations

1. Well-known types are not supported.
2. Import paths are static and cannot be configured.
3. Import paths across packages are not supported.
4. Self-referencing types are not supported

```proto
message Chat {
  message Type {
    string name = 1;
  }
  message Salutation {
    Chat.Type type = 1;  // <-- this is *not* supported
    string greeting = 2;
  }
  Type type = 1;  // <-- this is supported
  Salutation salutation = 2;
}
```
