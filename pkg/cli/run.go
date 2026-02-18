package cli

import (
	"context"
	"strings"

	"aig/pkg/docker"
	"aig/pkg/layers"

	"github.com/spf13/cobra"
)

var (
	baseImage    string
	layerNames   []string
	topLayerName string
	volumes      []string
	ports        []string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and run a customized container",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder, err := docker.NewBuilder()
		if err != nil {
			return err
		}

		base := &layers.BaseLayer{
			Image:   baseImage,
			Volumes: volumes,
			Ports:   ports,
		}
		
		var selectedLayers []layers.Layer
		for _, name := range layerNames {
			l, err := layers.Get(strings.TrimSpace(name))
			if err != nil {
				return err
			}
			selectedLayers = append(selectedLayers, l)
		}

		if topLayerName != "" {
			top, err := layers.Get(topLayerName)
			if err != nil {
				return err
			}
			selectedLayers = append(selectedLayers, top)
		}

		return builder.BuildAndRun(context.Background(), base, selectedLayers)
	},
}

func init() {
	runCmd.Flags().StringVarP(&baseImage, "base", "b", "ubuntu:22.04", "Base docker image")
	runCmd.Flags().StringSliceVarP(&layerNames, "layers", "l", []string{}, "Comma-separated list of layers to include")
	runCmd.Flags().StringVarP(&topLayerName, "top", "t", "", "Top layer (binary layer)")
	runCmd.Flags().StringSliceVarP(&volumes, "volume", "v", []string{}, "Bind mount a volume (e.g. /host:/container)")
	runCmd.Flags().StringSliceVarP(&ports, "port", "p", []string{}, "Publish a container's port(s) to the host (e.g. 8080:80)")
	
	rootCmd.AddCommand(runCmd)
}
