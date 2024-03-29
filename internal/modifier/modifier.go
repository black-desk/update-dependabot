package modifier

import (
	. "github.com/black-desk/lib/go/errwrap"
	"github.com/black-desk/update-dependabot/internal/logger"
	"github.com/black-desk/update-dependabot/internal/types"
	"gopkg.in/yaml.v3"
)

var log = logger.Log

type Modifier struct {
}

func New() *Modifier {
	return nil
}

func (m *Modifier) Modify(cfg *yaml.Node, updates []types.Update) (err error) {
	defer Wrap(&err, "modify yaml")
	if err = m.initCfg(cfg, updates); err != nil {
		return
	}

	if updates == nil || len(updates) == 0 {
		return
	}

	for i := range cfg.Content[0].Content {
		if cfg.Content[0].Content[i].Value != "updates" {
			continue
		}

		if err = m.modifyUpdates(
			cfg.Content[0].Content[i+1], updates,
		); err != nil {
			return
		}
		break
	}

	return
}

func (m *Modifier) initCfg(cfg *yaml.Node, updates []types.Update) (err error) {
	defer Wrap(&err)

	found := false

	if cfg != nil && len(cfg.Content) != 0 {
		for i := range cfg.Content[0].Content {
			if cfg.Content[0].Content[i].Value != "updates" {
				continue
			}
			found = true
			break
		}
	}

	if found {
		return
	}

	if updates == nil || len(updates) == 0 {
		return
	}

	if err = yaml.Unmarshal(
		[]byte("version: 2\nupdates: []"), cfg,
	); err != nil {
		return
	}

	cfg.Content[0].Content[3].Style = yaml.Style(yaml.FoldedStyle)

	return
}

type update struct {
	Directory        string `yaml:"directory"`
	PackageEcosystem string `yaml:"package-ecosystem"`

	Schedule struct {
		Interval string `yaml:"interval"`
	} `yaml:"schedule"`
}

func (u *update) node() (node *yaml.Node, err error) {
	defer Wrap(&err)

	var buf []byte
	if buf, err = yaml.Marshal(u); err != nil {
		return
	}

	node = &yaml.Node{}
	if err = yaml.Unmarshal(buf, node); err != nil {
		return
	}

	node = node.Content[0]

	return
}

func (m *Modifier) modifyUpdates(
	updatesNode *yaml.Node, updates []types.Update,
) (
	err error,
) {
	for i := 0; i < len(updatesNode.Content); i++ {
		found := false
		updateNode := updatesNode.Content[i]

		var update update
		if err = updateNode.Decode(&update); err != nil {
			return
		}

		for j := 0; j < len(updates); j++ {
			if updates[j].Directory != update.Directory {
				continue
			}

			if updates[j].PackageEcosystem.String() !=
				update.PackageEcosystem {
				continue
			}

			found = true
			log.Debugw("Already exists", "update", update)
			updates = append(updates[:j], updates[j+1:]...)
		}

		if !found {
			log.Debugw("Remove", "update", update)
			updatesNode.Content = append(
				updatesNode.Content[:i],
				updatesNode.Content[i+1:]...,
			)
			i--
		}
	}

	for i := range updates {
		update := update{
			Directory:        updates[i].Directory,
			PackageEcosystem: updates[i].PackageEcosystem.String(),
			Schedule: struct {
				Interval string `yaml:"interval"`
			}{
				Interval: "daily",
			},
		}

		var node *yaml.Node
		if node, err = update.node(); err != nil {
			return
		}

		log.Debugw("Add", "update", update)
		updatesNode.Content = append(updatesNode.Content, node)
	}

	return
}
