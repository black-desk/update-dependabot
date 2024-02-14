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
	defer Wrap(&err)

	if !filepath.IsAbs(opt.File) {
		opt.File = filepath.Join(opt.Dir, opt.File)
	}

	dependabotCfgFileContent, err := os.ReadFile(opt.File)
	if err != nil && !os.IsNotExist(err) {
		Wrap(&err, "read %s", opt.File)
		return
	}

	scanner := scanner.New()
	updates, err := scanner.Scan(opt.Dir)
	if err != nil {
		Wrap(&err, "scan %s", opt.Dir)
		return
	}

	log.Debugw("scanner.Scan done.", "updates", updates)

	cfg := yaml.Node{}
	err = yaml.Unmarshal(dependabotCfgFileContent, &cfg)
	if err != nil {
		Wrap(&err, "prase %s", opt.File)
		return
	}

	modifier := modifier.New()
	err = modifier.Modify(&cfg, updates)
	if err != nil {
		return err
	}

	if cfg.IsZero() {
		return
	}

	newCfgContent, err := yaml.Marshal(&cfg)
	if err != nil {
		Wrap(&err, "marshal new dependabot.yml")
		return
	}

	if opt.DryRun {
		reader := bytes.NewReader(newCfgContent)
		_, err = io.Copy(os.Stdout, reader)
		Wrap(&err, "write new dependabot.yml to stdout")
		return
	}

	dir := filepath.Dir(opt.File)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		Wrap(&err, "create %s", dir)
		return
	}

	err = os.WriteFile(opt.File, newCfgContent, 0644)
	if err != nil {
		Wrap(&err, "write %s", opt.File)
		return
	}

	return
}
