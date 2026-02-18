package layers

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Layer defines the interface for a container layer.
type Layer interface {
	GetName() string
	GetCommands() []string
	GetHash() string
	GetVolumes() []string
	GetPorts() []string
}

// BaseLayer represents the starting point (FROM ...).
type BaseLayer struct {
	Image   string
	Volumes []string
	Ports   []string
}

func (l *BaseLayer) GetName() string     { return "base" }
func (l *BaseLayer) GetCommands() []string { return []string{fmt.Sprintf("FROM %s", l.Image)} }
func (l *BaseLayer) GetHash() string {
	return hashString(l.Image + strings.Join(l.Volumes, ",") + strings.Join(l.Ports, ","))
}
func (l *BaseLayer) GetVolumes() []string { return l.Volumes }
func (l *BaseLayer) GetPorts() []string   { return l.Ports }

// DependencyLayer installs specific packages.
type DependencyLayer struct {
	Name    string
	Pkgs    []string
	Volumes []string
	Ports   []string
}

func (l *DependencyLayer) GetName() string { return l.Name }
func (l *DependencyLayer) GetCommands() []string {
	return []string{
		"RUN apt-get update && apt-get install -y " + strings.Join(l.Pkgs, " ") + " && rm -rf /var/lib/apt/lists/*",
	}
}
func (l *DependencyLayer) GetHash() string {
	return hashString(strings.Join(l.Pkgs, ",") + strings.Join(l.Volumes, ",") + strings.Join(l.Ports, ","))
}
func (l *DependencyLayer) GetVolumes() []string { return l.Volumes }
func (l *DependencyLayer) GetPorts() []string   { return l.Ports }

// CustomLayer allows for arbitrary commands and metadata.
type CustomLayer struct {
	Name     string
	Commands []string
	Volumes  []string
	Ports    []string
}

func (l *CustomLayer) GetName() string     { return l.Name }
func (l *CustomLayer) GetCommands() []string { return l.Commands }
func (l *CustomLayer) GetHash() string {
	return hashString(l.Name + strings.Join(l.Commands, "") + strings.Join(l.Volumes, "") + strings.Join(l.Ports, ""))
}
func (l *CustomLayer) GetVolumes() []string { return l.Volumes }
func (l *CustomLayer) GetPorts() []string   { return l.Ports }

// TopLayer ensures a binary is available and sets it as the entrypoint.
type TopLayer struct {
	Name       string
	BinaryURL  string // Could be a URL or local path
	BinaryPath string // Destination in container
	Volumes    []string
	Ports      []string
}

func (l *TopLayer) GetName() string { return l.Name }
func (l *TopLayer) GetCommands() []string {
	cmds := []string{}
	if strings.HasPrefix(l.BinaryURL, "http") {
		cmds = append(cmds, fmt.Sprintf("ADD %s %s", l.BinaryURL, l.BinaryPath))
	} else {
		cmds = append(cmds, fmt.Sprintf("COPY %s %s", l.BinaryURL, l.BinaryPath))
	}
	cmds = append(cmds, fmt.Sprintf("RUN chmod +x %s", l.BinaryPath))
	cmds = append(cmds, fmt.Sprintf("ENTRYPOINT [\"%s\"]", l.BinaryPath))
	return cmds
}
func (l *TopLayer) GetHash() string {
	return hashString(l.BinaryURL + ":" + l.BinaryPath + strings.Join(l.Volumes, ",") + strings.Join(l.Ports, ","))
}
func (l *TopLayer) GetVolumes() []string { return l.Volumes }
func (l *TopLayer) GetPorts() []string   { return l.Ports }

// CustomTopLayer allows for arbitrary setup commands and an entrypoint.
type CustomTopLayer struct {
	Name     string
	Commands []string
	Entry    []string
	HashKey  string
}

func (l *CustomTopLayer) GetName() string     { return l.Name }
func (l *CustomTopLayer) GetCommands() []string {
	cmds := append([]string{}, l.Commands...)
	if len(l.Entry) > 0 {
		cmds = append(cmds, fmt.Sprintf("ENTRYPOINT [\"%s\"]", strings.Join(l.Entry, "\", \"")))
	}
	return cmds
}
func (l *CustomTopLayer) GetHash() string {
	return hashString(l.Name + l.HashKey + strings.Join(l.Commands, "") + strings.Join(l.Entry, ""))
}
func (l *CustomTopLayer) GetVolumes() []string { return nil }
func (l *CustomTopLayer) GetPorts() []string   { return nil }

func hashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
