package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/black-desk/lib/go/errwrap"
	"github.com/black-desk/update-dependabot/internal/logger"
	"github.com/black-desk/update-dependabot/internal/modifier"
	"github.com/black-desk/update-dependabot/internal/scanner"
	"github.com/black-desk/update-dependabot/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cmdOpt.File, "file", "F", ".github/dependabot.yml", "Path to dependabot.yml.")
	rootCmd.Flags().StringVarP(&cmdOpt.Dir, "dir", "D", ".", "Path to github repo root to scan.")
	rootCmd.Flags().BoolVarP(&cmdOpt.DryRun, "dry-run", "d", false, "Dry run.")
}

var log = logger.Log
var cmdOpt types.CmdOpt

var rootCmd = &cobra.Command{
	Use:   "update-dependabot [-F file] [-D dir] [-d]",
	Short: "A tool to update dependabot.yml of your github project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			log.Warn("unused arguments", args)
		}
		return cmdRoot(cmdOpt)
	},
}

func cmdRoot(opt types.CmdOpt) (err error) {

	if !filepath.IsAbs(opt.File) {
		opt.File = filepath.Join(opt.Dir, opt.File)
	}

	dependabotCfgFileContent, err := os.ReadFile(opt.File)
	if err != nil && !os.IsNotExist(err) {
		return Trace(err, "Failed to read %s", opt.File)
	}

	scanner := scanner.New()
	updates, err := scanner.Scan(opt.Dir)
	if err != nil {
		return Trace(err, "Failed to scan %s", opt.Dir)
	}

	log.Debugw("scanner.Scan done.", "updates", updates)

	cfg := yaml.Node{}
	err = yaml.Unmarshal(dependabotCfgFileContent, &cfg)
	if err != nil {
		return Trace(err, "Failed to prase %s", opt.File)
	}

	modifier := modifier.New()
	err = modifier.Modify(&cfg, updates)
	if err != nil {
		return Trace(err, "Failed to generate new dependabot.yml")
	}

	if cfg.IsZero() {
		return
	}

	newCfgContent, err := yaml.Marshal(&cfg)
	if err != nil {
		return Trace(err, "Failed to marshal new dependabot.yml")
	}

	if opt.DryRun {
		reader := bytes.NewReader(newCfgContent)
		_, err := io.Copy(os.Stdout, reader)
		return Trace(err, "Failed to write new dependabot.yml to stdout")
	}

	err = os.MkdirAll(filepath.Dir(opt.File), 0755)
	if err != nil {
		return Trace(err, "Failed to create \"%s\"", filepath.Dir(opt.File))
	}

	err = os.WriteFile(opt.File, newCfgContent, 0644)
	if err != nil {
		return Trace(err, "Failed to write %s", opt.File)
	}

        return
}
