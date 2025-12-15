package proto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// LoadProtos loads all .proto files from the given path and returns a Registry
func LoadProtos(protoPath string, importPaths []string) (*Registry, error) {
	// Verify proto path exists
	info, err := os.Stat(protoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("proto path does not exist: %s", protoPath)
		}
		return nil, fmt.Errorf("cannot access proto path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("proto path is not a directory: %s", protoPath)
	}

	// Find all .proto files
	var protoFiles []string
	err = filepath.Walk(protoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories (e.g., .git, .idea)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			// Get relative path from protoPath
			relPath, err := filepath.Rel(protoPath, path)
			if err != nil {
				return err
			}
			protoFiles = append(protoFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk proto directory: %w", err)
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no .proto files found in: %s", protoPath)
	}

	// Build import paths: protoPath + user-specified + well-known types
	allImportPaths := []string{protoPath}
	allImportPaths = append(allImportPaths, importPaths...)

	// Create compiler with resolver, including well-known types (google/protobuf/*)
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: allImportPaths,
		}),
	}

	// Compile all proto files
	files, err := compiler.Compile(context.Background(), protoFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile protos: %w", err)
	}

	// Build registry from compiled files
	registry := NewRegistry()
	for _, f := range files {
		registry.AddFile(f)
	}

	return registry, nil
}

// ServiceInfo contains information about a gRPC service
type ServiceInfo struct {
	FullName string
	Methods  []MethodInfo
}

// MethodInfo contains information about a gRPC method
type MethodInfo struct {
	Name       string
	InputType  string
	OutputType string
}

// Registry holds parsed proto file descriptors and provides lookup methods
type Registry struct {
	files    []protoreflect.FileDescriptor
	services map[string]protoreflect.ServiceDescriptor
}

// NewRegistry creates a new empty Registry
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]protoreflect.ServiceDescriptor),
	}
}

// AddFile adds a file descriptor to the registry
func (r *Registry) AddFile(fd protoreflect.FileDescriptor) {
	r.files = append(r.files, fd)

	// Index all services
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		svc := services.Get(i)
		r.services[string(svc.FullName())] = svc
	}
}

// ListServices returns information about all registered services
func (r *Registry) ListServices() []ServiceInfo {
	var result []ServiceInfo

	for name, svc := range r.services {
		info := ServiceInfo{
			FullName: name,
		}

		methods := svc.Methods()
		for i := 0; i < methods.Len(); i++ {
			m := methods.Get(i)
			info.Methods = append(info.Methods, MethodInfo{
				Name:       string(m.Name()),
				InputType:  string(m.Input().FullName()),
				OutputType: string(m.Output().FullName()),
			})
		}

		result = append(result, info)
	}

	return result
}

// FindService finds a service by its fully qualified name
func (r *Registry) FindService(name string) (protoreflect.ServiceDescriptor, error) {
	svc, ok := r.services[name]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", name)
	}
	return svc, nil
}

// FindMethod finds a method by service and method name
func (r *Registry) FindMethod(serviceName, methodName string) (protoreflect.MethodDescriptor, error) {
	svc, err := r.FindService(serviceName)
	if err != nil {
		return nil, err
	}

	methods := svc.Methods()
	for i := 0; i < methods.Len(); i++ {
		m := methods.Get(i)
		if string(m.Name()) == methodName {
			return m, nil
		}
	}

	// Provide helpful error with available methods
	var available []string
	for i := 0; i < methods.Len(); i++ {
		available = append(available, string(methods.Get(i).Name()))
	}
	return nil, fmt.Errorf("method %q not found in service %s. Available methods: %s",
		methodName, serviceName, strings.Join(available, ", "))
}
