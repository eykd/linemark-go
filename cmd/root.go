// Package cmd contains the CLI commands for the lmk application.
package cmd

import (
	"context"
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/eykd/linemark-go/internal/fs"
	"github.com/eykd/linemark-go/internal/lock"
	"github.com/eykd/linemark-go/internal/outline"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

// verbose holds the global --verbose flag state.
var verbose bool

// jsonFlag holds the global --json flag state.
var jsonFlag bool

// dryRun holds the global --dry-run flag state.
var dryRun bool

func init() {
	rootCmd = BuildCommandTree(nil, nil)
}

// GetVerbose returns the current verbose flag state.
// This is used by other packages to check if debug logging is enabled.
func GetVerbose() bool {
	return verbose
}

// GetJSON returns the current global --json flag state.
func GetJSON() bool {
	return jsonFlag
}

// GetDryRun returns the current global --dry-run flag state.
func GetDryRun() bool {
	return dryRun
}

// NewRootCmd creates a new root command instance.
// This is useful for testing to get a fresh command tree.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "lmk",
		Short:         "Manage long-form prose projects with organized Markdown files",
		Long:          "lmk is a CLI tool for managing long-form prose projects using organized Markdown files.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Add persistent flags (available to all subcommands)
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging to stderr")
	cmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying files")

	return cmd
}

// BuildCommandTree creates a complete root command with all subcommands wired
// to the given service. If svc is nil, all runners will be nil (nil guards in
// each command will return ErrNotInProject). The bootstrapAdd parameter is
// accepted for API compatibility but is no longer used.
func BuildCommandTree(svc *outline.OutlineService, bootstrapAdd AddRunner) *cobra.Command {
	root := NewRootCmd()

	var aa AddRunner
	var ca CheckRunner
	var ra RepairRunner
	var la ListRunner
	var da DeleteRunner
	var ma MoveRunner
	var rna RenameRunner
	var cpa CompactRunner
	var ta TypesService

	if svc != nil {
		aa = &addAdapter{svc: svc}
		ca = &checkAdapter{svc: svc}
		ra = &repairAdapter{svc: svc}
		la = &listAdapter{svc: svc}
		da = &deleteAdapter{svc: svc}
		ma = &moveAdapter{svc: svc}
		rna = &renameAdapter{svc: svc}
		cpa = &compactAdapter{svc: svc}
		ta = &typesAdapter{svc: svc}
	}

	// Commands that work without a project
	root.AddCommand(NewInitCmd(os.Getwd))

	// Commands that require a project
	root.AddCommand(NewAddCmd(aa))
	root.AddCommand(NewCheckCmd(ca))
	root.AddCommand(NewDoctorCmd(ca, ra))
	root.AddCommand(NewTypesCmd(ta))
	root.AddCommand(NewCompactCmd(cpa))
	root.AddCommand(NewListCmd(la))
	root.AddCommand(NewDeleteCmd(da))
	root.AddCommand(NewMoveCmd(ma))
	root.AddCommand(NewRenameCmd(rna))

	return root
}

// wireServiceFromRootImpl creates an OutlineService for the given project root.
func wireServiceFromRootImpl(projectRoot string) (*outline.OutlineService, error) {
	reader := &fs.OSReader{Root: projectRoot}
	writer := &fs.OSWriter{Root: projectRoot}
	deleter := &fs.OSDeleter{Root: projectRoot}
	renamer := &fs.OSRenamer{Root: projectRoot}
	contentReader := &fs.OSContentReader{Root: projectRoot}
	reserver := &fs.SIDReserver{Rand: rand.Reader}
	locker := lock.NewFromPath(filepath.Join(projectRoot, lock.DefaultPath))

	reservationStore := &fs.OSReservationStore{Root: projectRoot}

	svc := outline.NewOutlineService(reader, writer, locker, reserver,
		outline.WithDeleter(deleter),
		outline.WithRenamer(renamer),
		outline.WithContentReader(contentReader),
		outline.WithSlugifier(fs.SlugAdapter{}),
		outline.WithFrontmatterHandler(fs.FMAdapter{}),
		outline.WithReservationStore(reservationStore),
	)

	return svc, nil
}

// wireFromCwdImpl detects the project root and wires the OutlineService.
func wireFromCwdImpl() (*outline.OutlineService, error) {
	projectRoot, err := fs.FindProjectRootImpl()
	if err != nil {
		return nil, err
	}
	return wireServiceFromRootImpl(projectRoot)
}

// Execute runs the root command and returns any error.
// Deprecated: Use ExecuteContext instead for proper signal handling.
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContextImpl runs the root command with the given context.
// This enables graceful shutdown via context cancellation (e.g., on SIGINT).
// It is an Impl function because it wraps OS operations (wireFromCwdImpl).
func ExecuteContextImpl(ctx context.Context) error {
	svc, _ := wireFromCwdImpl()
	root := BuildCommandTree(svc, nil)
	return root.ExecuteContext(ctx)
}

// ExecuteContext delegates to ExecuteContextImpl.
func ExecuteContext(ctx context.Context) error {
	return ExecuteContextImpl(ctx)
}
