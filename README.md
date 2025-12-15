# gRPC Client CLI

A dynamic command-line gRPC client that loads proto files at runtime and invokes gRPC services without pre-generated code. Supports gRPC, gRPC-Web, and Connect protocols.

## Features

- **Dynamic Proto Loading** – Parse `.proto` files at runtime without code generation
- **Multiple Protocols** – Support for gRPC, gRPC-Web, and Connect
- **File-Based Requests** – Define requests in `.grpc` files (inspired by [Hurl](https://hurl.dev/))
- **JSON I/O** – Send JSON input and receive JSON output
- **Custom Headers** – Add authentication and custom headers
- **Service Discovery** – List all available services and methods

## Installation

```bash
go build -o grpc_client .
```

## Usage

### List Services

Discover available gRPC services and methods from proto files:

```bash
grpc_client list -p ./path/to/protos
```

**Example output:**
```
Services:
  example.UserService
    - GetUser (GetUserRequest) → GetUserResponse
    - CreateUser (CreateUserRequest) → CreateUserResponse
```

### Call a Method

Invoke a gRPC method directly from the command line:

```bash
grpc_client call -p ./protos \
  --address http://localhost:8080 \
  --service example.UserService \
  --method GetUser \
  --data '{"user_id": "123"}'
```

**With headers and custom prefix:**
```bash
grpc_client call -p ./protos \
  --address http://localhost:8080 \
  --prefix /api/grpc \
  --service example.UserService \
  --method GetUser \
  --data '{"user_id": "123"}' \
  --header "Authorization: Bearer token123" \
  --header "X-Tenant: my-tenant"
```

### Run from File

Execute gRPC requests defined in `.grpc` files:

```bash
grpc_client run -p ./protos ./request.grpc
```

## Request File Format

The `.grpc` file format provides a clean, declarative way to define gRPC requests:

```
# Comment line (ignored)
GRPC http://localhost:8080/api/grpc
Service: example.UserService
Method: GetUser
Authorization: Bearer token123
X-Custom-Header: custom-value

{
  "user_id": "123",
  "include_details": true
}
```

### File Structure

| Line | Description |
|------|-------------|
| `# ...` | Comment (ignored) |
| `GRPC <url>` | Server address with optional path prefix |
| `Service: <name>` | Fully qualified service name |
| `Method: <name>` | Method to call |
| `Protocol: <type>` | Optional: `grpc`, `grpc-web`, or `connect` (default: `grpc-web`) |
| `Timeout: <duration>` | Optional: Request timeout (default: `30s`) |
| `<Header>: <Value>` | HTTP headers (any other key-value pairs) |
| `{ ... }` | JSON request body |

### Example Files

**Simple request:**
```
GRPC http://localhost:8080
Service: api.HealthService
Method: Check
{}
```

**With authentication:**
```
GRPC http://localhost:8082/web-grpc
Service: myapp.api.v1.CustomerService
Method: GetList
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
X-Tenant: production
Timeout: 60s

{
  "page": 1,
  "limit": 50
}
```

### Multiple Requests and Chaining

You can define multiple requests in a single file separated by `---`. This allows for request chaining where values from one response can be captured and used in subsequent requests.

**Supported Features:**
- **Captures**: Extract values from JSON response using `[Captures]` section.
- **Variables**: Use captured values with `{{variable_name}}` syntax.
- **JSONPath**: Use dot notation (`user.id`) or array indexing (`users[0].name`) to extract values.

**Example with Chaining:**
```
# Request 1: Login
GRPC http://localhost:8080
Service: example.AuthService
Method: Login

{
  "username": "admin",
  "password": "secret"
}

[Captures]
auth_token: token
user_id: user.details.id

---

# Request 2: Get User (uses captured token and ID)
GRPC http://localhost:8080
Service: example.UserService
Method: GetUser
Authorization: Bearer {{auth_token}}

{
  "id": "{{user_id}}"
}
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--proto-path` | `-p` | Path to folder containing `.proto` files (required) |
| `--import-path` | `-I` | Additional import paths for proto dependencies |

## Call Command Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--address` | `-a` | Server address (required) | - |
| `--service` | `-s` | Fully qualified service name (required) | - |
| `--method` | `-m` | Method name (required) | - |
| `--data` | `-d` | JSON input for the request | `{}` |
| `--prefix` | | Route prefix for gRPC-Web endpoints | - |
| `--header` | `-H` | HTTP headers (repeatable) | - |
| `--protocol` | | Protocol: `grpc`, `grpc-web`, `connect` | `grpc-web` |
| `--timeout` | | Request timeout | `30s` |

## Protocols

| Protocol | Description |
|----------|-------------|
| `grpc` | Native gRPC over HTTP/2 |
| `grpc-web` | gRPC-Web for browser-compatible endpoints |
| `connect` | Connect protocol (Buf Connect) |

## Project Structure

```
grpc_client/
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # Root command and global flags
│   ├── list.go          # List services command
│   ├── call.go          # Call method command
│   └── run.go           # Run from file command
├── internal/
│   ├── client/          # gRPC client implementation
│   ├── file/            # .grpc file parser
│   └── proto/           # Proto file loading and registry
└── testdata/            # Test proto files
```

## License

MIT
