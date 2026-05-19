## Context

You are in the bcs-terraform-bkprovider repo, which is a gRPC microservice in the BlueKing Container Service (BCS) ecosystem. It acts as an API proxy between the BCS Terraform controller and backend systems — specifically BK NodeMan for host/agent management and Tencent Cloud VPC for IP whitelist management. You are helping implement features, fix bugs, and refactor existing code.

## Source code

* bcs-terraform-bkprovider is a Go gRPC service built on go-micro v4, providing host agent management (via BK NodeMan API Gateway) and IP whitelist management (via Tencent Cloud VPC SDK).
* The main entry point is in `main.go`, which delegates to `cmd/server/`.
* Core packages are organized as follows:
  - `cmd/server/`: Server initialization — Cobra CLI, config loading, gRPC service, HTTP gateway, signal handling
  - `handler/`: gRPC handler implementations for all RPC methods (core business logic)
  - `common/`: Configuration structs, error codes, and utility functions
  - `middleware/xbknodeman/`: BK NodeMan API Gateway client (cloud area, host, and job operations)
  - `middleware/xtencentcloud/`: Tencent Cloud VPC client (IP whitelist management)
  - `middleware/xrequests/`: HTTP request utilities built on grequests
  - `pkg/middleware/auth/`: JWT authentication and authorization middleware
  - `proto/`: Protobuf service definition and generated Go/gRPC/Gateway/Swagger code
  - `sdk/bcsprovider-sdk-go/`: Standalone Go SDK client library for cluster, Helm, and project management
* API documentation is available via Swagger UI and generated from the proto definition.
* Unit tests are placed alongside their source files with `_test.go` suffix.

## Coding style

* For Go files, follow the official Go style guide and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
* Use `goimports` to format imports.
* Run `gofmt` before committing.
* Follow the existing naming conventions:
  - External middleware packages use `x` prefix (e.g., `xbknodeman`, `xtencentcloud`, `xrequests`)
  - Proto-generated files remain in `proto/` package
  - Config structs are defined in `common/options.go`

### Running our tests

* Run all tests: `make test`
* ALWAYS prefer specifying test packages for efficiency, e.g. `go test -v ./middleware/xbknodeman/... -cover`
* Before running any Go commands, ensure gvm-managed Go 1.23.12 is active (not the system Go).

### Building

* Build the binary: `make build`
* Build the Docker image: `make docker`

### Proto code generation

* Install protoc toolchain: `make init`
* Regenerate proto code: `make proto`
* This will regenerate Go, gRPC, gRPC-Gateway, and Swagger files in `proto/`.

### Key dependencies

* `go-micro.dev/v4` — microservice framework
* `github.com/grpc-ecosystem/grpc-gateway` — HTTP REST gateway
* `github.com/spf13/cobra` — CLI framework
* `github.com/tencentcloud/tencentcloud-sdk-go` — Tencent Cloud SDK (VPC)
* `github.com/levigross/grequests` — HTTP client
