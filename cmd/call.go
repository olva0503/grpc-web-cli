package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"grpc_client/internal/client"
	"grpc_client/internal/proto"
)

var (
	address  string
	service  string
	method   string
	data     string
	prefix   string
	headers  []string
	protocol string
	timeout  time.Duration
)

var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Call a gRPC method",
	Long: `Invoke a gRPC method with JSON input and receive JSON output.

Example:
  grpc_client call -p ./protos \
    --address http://localhost:8080 \
    --service example.UserService \
    --method GetUser \
    --data '{"user_id": "123"}' \
    --prefix /api/grpc \
    --header "Authorization: Bearer token123"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load proto definitions
		registry, err := proto.LoadProtos(protoPath, importPaths)
		if err != nil {
			return fmt.Errorf("failed to load protos: %w", err)
		}

		// Find the method descriptor
		methodDesc, err := registry.FindMethod(service, method)
		if err != nil {
			// Provide helpful error with available services
			services := registry.ListServices()
			var available []string
			for _, s := range services {
				available = append(available, s.FullName)
			}
			return fmt.Errorf("%w\n\nAvailable services: %s", err, strings.Join(available, ", "))
		}

		// Parse headers
		headerMap := make(map[string]string)
		for _, h := range headers {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid header format %q, expected 'Key: Value'", h)
			}
			headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}

		// Parse protocol
		proto, err := client.ParseProtocol(protocol)
		if err != nil {
			return err
		}

		// Create the client
		c := client.NewClient(address, prefix, proto, headerMap)

		// Convert JSON input to proto message
		inputMsg, err := client.JSONToProto(data, methodDesc.Input())
		if err != nil {
			return fmt.Errorf("failed to parse JSON input: %w", err)
		}

		// Make the call
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		response, err := c.Call(ctx, methodDesc, inputMsg)
		if err != nil {
			return fmt.Errorf("RPC call failed: %w", err)
		}

		// Convert response to JSON
		jsonOutput, err := client.ProtoToJSON(response)
		if err != nil {
			return fmt.Errorf("failed to format response: %w", err)
		}

		fmt.Println(jsonOutput)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(callCmd)

	callCmd.Flags().StringVarP(&address, "address", "a", "", "server address (required)")
	callCmd.Flags().StringVarP(&service, "service", "s", "", "fully qualified service name (required)")
	callCmd.Flags().StringVarP(&method, "method", "m", "", "method name (required)")
	callCmd.Flags().StringVarP(&data, "data", "d", "{}", "JSON input for the request")
	callCmd.Flags().StringVar(&prefix, "prefix", "", "route prefix for gRPC-Web endpoints (e.g., /api/grpc)")
	callCmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "HTTP headers (format: 'Key: Value', can be repeated)")
	callCmd.Flags().StringVar(&protocol, "protocol", "grpc-web", "protocol: grpc, grpc-web, or connect")
	callCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "request timeout")

	_ = callCmd.MarkFlagRequired("address")
	_ = callCmd.MarkFlagRequired("service")
	_ = callCmd.MarkFlagRequired("method")
}
