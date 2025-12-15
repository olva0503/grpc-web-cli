package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	protoPath   string
	importPaths []string
)

var rootCmd = &cobra.Command{
	Use:   "grpc_client",
	Short: "A dynamic gRPC-Web client CLI",
	Long: `A command-line tool for invoking gRPC services via gRPC-Web protocol.

Load proto files at runtime, discover services and methods, and call them
with JSON input without needing pre-generated code.

Examples:
  # List all services and methods from proto files
  grpc_client list -p ./protos

  # Call a gRPC method
  grpc_client call -p ./protos \
    --address http://localhost:8080 \
    --service example.UserService \
    --method GetUser \
    --data '{"user_id": "123"}'
`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&protoPath, "proto-path", "p", "", "path to folder containing .proto files (required)")
	rootCmd.PersistentFlags().StringArrayVarP(&importPaths, "import-path", "I", nil, "additional import paths for proto dependencies")
	_ = rootCmd.MarkPersistentFlagRequired("proto-path")
}
