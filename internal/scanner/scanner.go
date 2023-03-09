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

func (s *Scanner) Scan(root string) (updates []types.Update, err error) {

	scanFns := [](func(string) ([]types.Update, error)){
		s.scanGitSubmodule,
		s.scanGithubActions,
		s.scanGoModules,
	}

	for i := range scanFns {
		pendingUpdates, scanFnErr := scanFns[i](root)
		if scanFnErr != nil {
			err = Trace(scanFnErr, "Failed to scan in %s", root)
			return
		}
		updates = append(updates, pendingUpdates...)
	}

	return
}

func (s *Scanner) pushd(root string, fn func()) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		err = Trace(err, "Failed to get working dir")
		return
	}

	err = os.Chdir(root)
	if err != nil {
		err = Trace(err, "Failed to chdir to %s", root)
		return
	}

	fn()

	defer func() {
		popdErr := os.Chdir(pwd)
		if popdErr != nil {
			if err != nil {
				log.Warn("Failed to chdir back to %s", pwd)
			}
			err = popdErr
		}
	}()
	return

}

func (s *Scanner) scanGitSubmodule(root string) (updates []types.Update, err error) {

	err = Trace(s.pushd(root, func() {

		_, statErr := os.Stat(".gitmodules")
		if os.IsNotExist(statErr) {
			return
		}

		updates = []types.Update{{
			Directory:        "/",
			PackageEcosystem: types.PackageEcosystemGitSubmodule,
		}}

	}))

	return
}

func (s *Scanner) scanGithubActions(root string) (updates []types.Update, err error) {

	err = Trace(s.pushd(root, func() {

		_, statErr := os.Stat(filepath.Join(".github", "workflows"))
		if os.IsNotExist(statErr) {
			return
		}

		updates = []types.Update{{
			Directory:        "/",
			PackageEcosystem: types.PackageEcosystemGitHubActions,
		}}

	}))

	return
}

func (s *Scanner) scanGoModules(root string) (updates []types.Update, err error) {

	err = filepath.WalkDir(root, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return Trace(err)
			}
			return nil
		}

		name := f.Name()
		if name == "go.mod" {
			updates = append(updates, types.Update{
				Directory:        "/" + strings.TrimPrefix(filepath.Dir(path), root),
				PackageEcosystem: types.PackageEcosystemGoModules,
			})
		}

		return nil
	})

	if err != nil {
		err = Trace(err, "Failed to walk directory.")
		return
	}
	return
}
