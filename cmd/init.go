package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command. The getwd function returns the working
// directory where the project will be initialized.
func NewInitCmd(getwd func() (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:          "init",
		Short:        "Initialize a new linemark project in the current directory",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			dir := filepath.Join(cwd, ".linemark")
			info, statErr := os.Stat(dir)
			if statErr == nil && info.IsDir() {
				fmt.Fprintln(cmd.OutOrStdout(), "Linemark project already initialized")
				return nil
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating .linemark directory: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Initialized linemark project")
			return nil
		},
	}
}
