package api

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ../../oapi-codegen.yaml ../../api.json
//go:generate go run ./genreadonly ../../api.json readonly.gen.go
