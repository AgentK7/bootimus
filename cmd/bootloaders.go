package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"bootimus/bootloaders"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type bootloaderConfig struct {
	ActiveSet string `json:"active_set"`
}

var bootloadersCmd = &cobra.Command{
	Use:   "bootloaders",
	Short: "Manage bootloader sets used by the TFTP/HTTP server",
}

var bootloadersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available bootloader sets (embedded + custom)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sets, err := allSets()
		if err != nil {
			return err
		}
		active, err := readActiveSet()
		if err != nil {
			return err
		}
		if active == "" {
			active = bootloaders.DefaultSet
		}
		for _, s := range sets {
			marker := "  "
			if s.name == active {
				marker = "* "
			}
			origin := "custom"
			if s.builtIn {
				origin = "embedded"
			}
			fmt.Printf("%s%-20s %s\n", marker, s.name, origin)
		}
		return nil
	},
}

var bootloadersUseCmd = &cobra.Command{
	Use:   "use <set>",
	Short: "Set the active bootloader set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		sets, err := allSets()
		if err != nil {
			return err
		}
		found := false
		for _, s := range sets {
			if s.name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown bootloader set %q (run 'bootimus bootloaders list')", name)
		}

		persisted := name
		if name == bootloaders.DefaultSet {
			persisted = ""
		}
		if err := writeActiveSet(persisted); err != nil {
			return err
		}
		fmt.Printf("✓ active set: %s\n", name)
		fmt.Println("  takes effect on next server start")
		return nil
	},
}

var bootloadersActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "Print the currently active bootloader set",
	RunE: func(cmd *cobra.Command, args []string) error {
		active, err := readActiveSet()
		if err != nil {
			return err
		}
		if active == "" {
			active = bootloaders.DefaultSet
		}
		fmt.Println(active)
		return nil
	},
}

type setEntry struct {
	name    string
	builtIn bool
}

func allSets() ([]setEntry, error) {
	seen := map[string]bool{}
	var out []setEntry

	embedded, err := bootloaders.ListSets()
	if err != nil {
		return nil, err
	}
	for _, s := range embedded {
		out = append(out, setEntry{name: s, builtIn: true})
		seen[s] = true
	}

	bootDir := bootloadersDir()
	entries, err := os.ReadDir(bootDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() || seen[e.Name()] {
				continue
			}
			out = append(out, setEntry{name: e.Name(), builtIn: false})
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, nil
}

func dataDir() string {
	d := viper.GetString("data_dir")
	if d == "" {
		d = "./data"
	}
	return d
}

func bootloadersDir() string {
	return filepath.Join(dataDir(), "bootloaders")
}

func configPath() string {
	return filepath.Join(dataDir(), "bootloader-config.json")
}

func readActiveSet() (string, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var cfg bootloaderConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.ActiveSet, nil
}

func writeActiveSet(name string) error {
	if err := os.MkdirAll(dataDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(bootloaderConfig{ActiveSet: name}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

func init() {
	bootloadersCmd.AddCommand(bootloadersListCmd, bootloadersUseCmd, bootloadersActiveCmd)
	rootCmd.AddCommand(bootloadersCmd)
}
