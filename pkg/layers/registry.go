package layers

import "fmt"

var (
	registry = make(map[string]Layer)
)

// Register adds a layer to the registry.
func Register(l Layer) {
	registry[l.GetName()] = l
}

// Get retrieves a layer by name.
func Get(name string) (Layer, error) {
	l, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("layer %s not found", name)
	}
	return l, nil
}

// GetAllNames returns names of all registered layers.
func GetAllNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

func init() {
	// Register some default layers for demonstration
	Register(&DependencyLayer{
		Name: "python",
		Pkgs: []string{"python3", "python3-pip"},
	})
	Register(&DependencyLayer{
		Name: "node",
		Pkgs: []string{"nodejs", "npm"},
	})
	Register(&DependencyLayer{
		Name: "go",
		Pkgs: []string{"golang"},
	})
	Register(&DependencyLayer{
		Name:  "nginx",
		Pkgs:  []string{"nginx"},
		Ports: []string{"80"},
	})
	Register(&CustomLayer{
		Name: "node-pnpm",
		Commands: []string{
			"RUN apt-get update && apt-get install -y nodejs npm && rm -rf /var/lib/apt/lists/*",
			"RUN npm install -g pnpm",
		},
	})
	Register(&CustomTopLayer{
		Name: "opencode",
		Commands: []string{
			"RUN apt-get update && apt-get install -y nodejs npm && rm -rf /var/lib/apt/lists/*",
			"RUN npm install -g pnpm",
			"RUN pnpm add -g opencode-ai",
		},
		Entry:   []string{"opencode"},
		HashKey: "v1.1.59", // Using version as hash key
	})
	
	// Example Top Layer
	Register(&TopLayer{
		Name:       "hello-world",
		BinaryURL:  "https://github.com/docker-library/hello-world/raw/master/hello",
		BinaryPath: "/hello",
	})
}
