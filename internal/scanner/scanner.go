package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/black-desk/lib/go/errwrap"
	"github.com/black-desk/update-dependabot/internal/logger"
	"github.com/black-desk/update-dependabot/internal/types"
)

var log = logger.Log

type Scanner struct {
}

func New() *Scanner {
	return nil
}

func (s *Scanner) Scan(dir string) (updates []types.Update, err error) {
	defer Wrap(&err, "scan dir %s", dir)

	scanFns := [](func(string) ([]types.Update, error)){
		s.scanGitSubmodule,
		s.scanGithubActions,
		s.scanGoModules,
		s.scanNPM,
	}

	for i := range scanFns {
		pendingUpdates, scanFnErr := scanFns[i](dir)
		if scanFnErr != nil {
			err = scanFnErr
			Wrap(&err, "Failed to scan in %s", dir)
			return
		}
		updates = append(updates, pendingUpdates...)
	}

	return
}

func (s *Scanner) workInDir(dir string, fn func()) (err error) {
	defer Wrap(&err, "work in dir %s", dir)

	var pwd string
	pwd, err = os.Getwd()
	if err != nil {
		Wrap(&err, "get working dir")
		return
	}

	err = os.Chdir(dir)
	if err != nil {
		Wrap(&err, "chdir to %s", dir)
		return
	}

	fn()

	err = os.Chdir(pwd)
	if err != nil {
		Wrap(&err, "chdir back to %s", pwd)
		return
	}

	return
}

func (s *Scanner) scanGitSubmodule(
	dir string,
) (
	updates []types.Update, err error,
) {
	defer Wrap(&err)

	if err = s.workInDir(dir, func() {
		_, statErr := os.Stat(".gitmodules")
		if os.IsNotExist(statErr) {
			return
		}

		updates = []types.Update{{
			Directory:        "/",
			PackageEcosystem: types.PackageEcosystemGitSubmodule,
		}}
	}); err != nil {
		return
	}

	return
}

func (s *Scanner) scanGithubActions(
	dir string,
) (
	updates []types.Update, err error,
) {
	defer Wrap(&err)

	if err = s.workInDir(dir, func() {
		_, statErr := os.Stat(filepath.Join(".github", "workflows"))
		if os.IsNotExist(statErr) {
			return
		}

		updates = []types.Update{{
			Directory:        "/",
			PackageEcosystem: types.PackageEcosystemGitHubActions,
		}}
	}); err != nil {
		return
	}

	return
}

func (s *Scanner) scanGoModules(
	root string,
) (
	updates []types.Update, err error,
) {
	defer Wrap(&err)

	if err = filepath.WalkDir(root, func(
		path string, f fs.DirEntry, walkDirErr error,
	) (err error) {
		if walkDirErr != nil {
			if path == root {
				err = walkDirErr
				return
			}
			return
		}

		name := f.Name()
		if name == "go.mod" {
			updates = append(updates, types.Update{
				Directory: filepath.Clean(
					"/" + strings.TrimPrefix(
						filepath.Dir(path), root,
					)),
				PackageEcosystem: types.PackageEcosystemGoModules,
			})
		}

		return
	}); err != nil {
		Wrap(&err, "walk directory.")
		return
	}

	return
}

func (s *Scanner) scanNPM(
	root string,
) (
	updates []types.Update, err error,
) {
	defer Wrap(&err)

	if err = filepath.WalkDir(root, func(
		path string, f fs.DirEntry, walkDirErr error,
	) (err error) {
		if walkDirErr != nil {
			if path == root {
				err = walkDirErr
				return
			}
			return
		}

		if f.IsDir() && f.Name() == "node_modules" {
			return filepath.SkipDir
		}

		name := f.Name()
		if name == "package.json" {
			updates = append(updates, types.Update{
				Directory: filepath.Clean(
					"/" + strings.TrimPrefix(
						filepath.Dir(path), root,
					)),
				PackageEcosystem: types.PackageEcosystemNpm,
			})
		}

		return
	}); err != nil {
		Wrap(&err, "walk directory.")
		return
	}

	return
}
