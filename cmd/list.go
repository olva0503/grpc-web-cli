package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"grpc_client/internal/proto"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available services and methods from proto files",
	Long: `Parse proto files and display all available gRPC services and their methods.

Example:
  grpc_client list -p ./protos
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		registry, err := proto.LoadProtos(protoPath, importPaths)
		if err != nil {
			return fmt.Errorf("failed to load protos: %w", err)
		}

		services := registry.ListServices()
		if len(services) == 0 {
			fmt.Println("No services found in proto files.")
			return nil
		}

		fmt.Println("Services:")
		for _, svc := range services {
			fmt.Printf("  %s\n", svc.FullName)
			for _, method := range svc.Methods {
				fmt.Printf("    - %s (%s) → %s\n",
					method.Name,
					method.InputType,
					method.OutputType,
				)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
