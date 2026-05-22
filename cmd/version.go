package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"bootimus/internal/server"

	"github.com/spf13/cobra"
)

const (
	ipxeVersion = "1.21.1+upstream"
	licence     = "Apache-2.0"
)

var versionVerbose bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the bootimus version",
	Run: func(cmd *cobra.Command, args []string) {
		if !versionVerbose {
			fmt.Println(server.Version)
			return
		}

		commit, dirty, cgo := buildInfo()
		state := "clean"
		if dirty {
			state = "dirty"
		}
		buildTags := "cgo"
		if !cgo {
			buildTags = "static"
		}

		fmt.Printf("bootimus      %s\n", server.Version)
		fmt.Printf("commit        %s (%s)\n", commit, state)
		fmt.Printf("go            %s %s/%s\n", trimGoVersion(runtime.Version()), runtime.GOOS, runtime.GOARCH)
		fmt.Printf("build         %s\n", buildTags)
		fmt.Printf("licence       %s\n", licence)
		fmt.Println()
		fmt.Println("embedded")
		fmt.Printf("  ipxe        %s  GPL-2.0\n", ipxeVersion)
		fmt.Printf("  proprietary 0 blobs\n")
		fmt.Printf("  telemetry   disabled (compile-time)\n")
	},
}

func buildInfo() (commit string, dirty bool, cgo bool) {
	commit = "unknown"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) >= 7 {
				commit = s.Value[:7]
			} else {
				commit = s.Value
			}
		case "vcs.modified":
			dirty = s.Value == "true"
		case "CGO_ENABLED":
			cgo = s.Value == "1"
		}
	}
	return
}

func trimGoVersion(v string) string {
	if len(v) > 2 && v[:2] == "go" {
		return v[2:]
	}
	return v
}

func init() {
	versionCmd.Flags().BoolVarP(&versionVerbose, "verbose", "v", false, "Show detailed build, embedded component, and licence info")
	rootCmd.AddCommand(versionCmd)
}
