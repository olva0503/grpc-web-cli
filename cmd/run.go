package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"grpc_client/internal/assert"
	"grpc_client/internal/client"
	"grpc_client/internal/file"
	"grpc_client/internal/proto"
	"grpc_client/internal/template"
)

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Execute a gRPC request from a .grpc file",
	Long: `Execute a gRPC request defined in a .grpc file.

The file format is inspired by Hurl and contains all request details:
- Server address
- Service and method names
- Headers
- JSON request body

Example file (get_user.grpc):
  GRPC http://localhost:8080/api/grpc
  Service: example.UserService
  Method: GetUser
  Authorization: Bearer token123

  {
    "user_id": "123"
  }

Usage:
  grpc_client run -p ./protos ./get_user.grpc
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Parse the request file (may contain multiple requests)
		requests, err := file.ParseMultiple(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse request file: %w", err)
		}

		// Load proto definitions
		registry, err := proto.LoadProtos(protoPath, importPaths)
		if err != nil {
			return fmt.Errorf("failed to load protos: %w", err)
		}

		// Variable store for captures
		variables := make(map[string]interface{})

		// Execute each request
		for i, reqFile := range requests {
			// Print separator between requests
			if i > 0 {
				fmt.Println("\n---")
			}

			// Substitute variables in Address, Headers, and Body
			reqFile.Address = template.Substitute(reqFile.Address, variables)
			reqFile.Body = template.Substitute(reqFile.Body, variables)
			for k, v := range reqFile.Headers {
				reqFile.Headers[k] = template.Substitute(v, variables)
			}

			// Print request header
			if reqFile.Name != "" {
				fmt.Printf("# %s\n", reqFile.Name)
			} else {
				fmt.Printf("# Request %d\n", i+1)
			}
			fmt.Printf("# %s/%s\n\n", reqFile.Service, reqFile.Method)

			// Find the method descriptor
			methodDesc, err := registry.FindMethod(reqFile.Service, reqFile.Method)
			if err != nil {
				// Provide helpful error with available services
				services := registry.ListServices()
				var available []string
				for _, s := range services {
					available = append(available, s.FullName)
				}
				return fmt.Errorf("%w\n\nAvailable services: %s", err, strings.Join(available, ", "))
			}

			// Parse protocol
			proto, err := client.ParseProtocol(reqFile.Protocol)
			if err != nil {
				return err
			}

			// Extract prefix from address if present
			address, prefix := parseAddressAndPrefix(reqFile.Address)

			// Create the client
			c := client.NewClient(address, prefix, proto, reqFile.Headers)

			// Convert JSON input to proto message
			inputMsg, err := client.JSONToProto(reqFile.Body, methodDesc.Input())
			if err != nil {
				return fmt.Errorf("failed to parse JSON input: %w", err)
			}

			// Make the call
			ctx, cancel := context.WithTimeout(context.Background(), reqFile.Timeout)
			response, err := c.Call(ctx, methodDesc, inputMsg)
			cancel()

			if err != nil {
				return fmt.Errorf("RPC call failed: %w", err)
			}

			// Convert response to JSON
			jsonOutput, err := client.ProtoToJSON(response)
			if err != nil {
				return fmt.Errorf("failed to format response: %w", err)
			}

			fmt.Println(jsonOutput)

			// Handle Captures
			if len(reqFile.Captures) > 0 {
				fmt.Println("\n# Captures:")
				for varName, path := range reqFile.Captures {
					val, err := client.EvaluateJSONPath(jsonOutput, path)
					if err != nil {
						fmt.Printf("# Warning: failed to capture variable '%s' from path '%s': %v\n", varName, path, err)
						continue
					}
					variables[varName] = val
					fmt.Printf("# %s = %v\n", varName, val)
				}
			}

			// Handle Asserts
			if len(reqFile.Asserts) > 0 {
				fmt.Println("\n# Asserts:")
				allPassed := true
				for _, a := range reqFile.Asserts {
					result, err := assert.Check(a, jsonOutput)
					if err != nil {
						// Error executing check (e.g. invalid jsonpath)
						fmt.Printf("# ERROR: %v\n", err)
						allPassed = false
						continue
					}

					fmt.Printf("# %s\n", result.Message)
					if !result.Pass {
						allPassed = false
					}
				}

				if !allPassed {
					return fmt.Errorf("one or more assertions failed")
				}
			}
		}

		return nil
	},
}

// parseAddressAndPrefix splits a URL into base address and path prefix
// e.g., "http://localhost:8080/api/grpc" -> ("http://localhost:8080", "/api/grpc")
func parseAddressAndPrefix(address string) (string, string) {
	// Find the third slash (after http://)
	count := 0
	for i, c := range address {
		if c == '/' {
			count++
			if count == 3 {
				return address[:i], address[i:]
			}
		}
	}
	return address, ""
}

func init() {
	rootCmd.AddCommand(runCmd)
}
