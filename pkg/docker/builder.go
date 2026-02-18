package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"aig/pkg/layers"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Builder struct {
	cli *client.Client
}

func NewBuilder() (*Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Builder{cli: cli}, nil
}

func (b *Builder) BuildAndRun(ctx context.Context, base layers.Layer, selectedLayers []layers.Layer) error {
	// Generate Dockerfile
	dockerfile := b.generateDockerfile(base, selectedLayers)
	
	// Calculate Hash
	tag := b.calculateTag(base, selectedLayers)
	imageName := fmt.Sprintf("aig-image:%s", tag)

	// Collect volumes and ports
	var volumes []string
	var ports []string

	volumes = append(volumes, base.GetVolumes()...)
	ports = append(ports, base.GetPorts()...)

	for _, l := range selectedLayers {
		volumes = append(volumes, l.GetVolumes()...)
		ports = append(ports, l.GetPorts()...)
	}

	// Check if image exists
	exists, err := b.imageExists(ctx, imageName)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Building image %s...\n", imageName)
		if err := b.buildImage(ctx, dockerfile, imageName); err != nil {
			return err
		}
	} else {
		fmt.Printf("Using cached image %s\n", imageName)
	}

	// Run container
	return b.runContainer(ctx, imageName, volumes, ports)
}

func (b *Builder) generateDockerfile(base layers.Layer, selected []layers.Layer) string {
	var sb strings.Builder
	for _, cmd := range base.GetCommands() {
		sb.WriteString(cmd + "\n")
	}
	for _, l := range selected {
		for _, cmd := range l.GetCommands() {
			sb.WriteString(cmd + "\n")
		}
	}
	return sb.String()
}

func (b *Builder) calculateTag(base layers.Layer, selected []layers.Layer) string {
	h := sha256.New()
	h.Write([]byte(base.GetHash()))
	for _, l := range selected {
		h.Write([]byte(l.GetHash()))
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

func (b *Builder) imageExists(ctx context.Context, name string) (bool, error) {
	_, _, err := b.cli.ImageInspectWithRaw(ctx, name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b *Builder) buildImage(ctx context.Context, dockerfile, tag string) error {
	// Create build context (tar)
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	
	header := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	options := types.ImageBuildOptions{
		Context:    buf,
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
	}

	resp, err := b.cli.ImageBuild(ctx, buf, options)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Stream build output to stdout
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func (b *Builder) runContainer(ctx context.Context, imageName string, volumes []string, ports []string) error {
	exposedPorts := make(nat.PortSet)
	portBindings := make(nat.PortMap)

	for _, p := range ports {
		parts := strings.Split(p, ":")
		var hostPort, containerPort string
		if len(parts) == 2 {
			hostPort = parts[0]
			containerPort = parts[1]
		} else {
			containerPort = parts[0]
		}

		if !strings.Contains(containerPort, "/") {
			containerPort = containerPort + "/tcp"
		}

		cPort := nat.Port(containerPort)
		exposedPorts[cPort] = struct{}{}

		if hostPort != "" {
			portBindings[cPort] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			}
		}
	}

	resp, err := b.cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		Tty:          true,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Binds:        volumes,
		PortBindings: portBindings,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := b.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	fmt.Printf("Container started (ID: %s)\n", resp.ID[:12])

	// Wait for container to exit and stream logs
	statusCh, errCh := b.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	
	// Stream logs
	out, err := b.cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err == nil {
		defer out.Close()
		go io.Copy(os.Stdout, out)
	}

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}
