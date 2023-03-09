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
	err = Trace(m.initCfg(cfg, updates))
	if err != nil {
		return
	}

	for i := range cfg.Content[0].Content {
		if cfg.Content[0].Content[i].Value == "updates" {
			err = Trace(m.modifyUpdates(cfg.Content[0].Content[i+1], updates))
			if err != nil {
				return
			}
		}
	}

	return
}

func (m *Modifier) initCfg(cfg *yaml.Node, updates []types.Update) (err error) {
	if cfg == nil || len(cfg.Content) == 0 {

		if updates == nil || len(updates) == 0 {
			return
		}

		err = yaml.Unmarshal([]byte(`
                        version: 2
                        updates: []
                `), cfg)

		if err != nil {
			err = Trace(err)
			return
		}

		cfg.Content[0].Content[3].Style = yaml.Style(yaml.FoldedStyle)
	}
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
	buf, err := yaml.Marshal(u)
	if err != nil {
		err = Trace(err)
		return
	}

	node = &yaml.Node{}
	err = yaml.Unmarshal(buf, node)
	if err != nil {
		err = Trace(err)
		return
	}

	node = node.Content[0]

	return
}

func (m *Modifier) modifyUpdates(updatesNode *yaml.Node, updates []types.Update) (err error) {
	for i := 0; i < len(updatesNode.Content); i++ {
		found := false
		updateNode := updatesNode.Content[i]

		var update update
		err = Trace(updateNode.Decode(&update))
		if err != nil {
			return
		}

		for j := 0; j < len(updates); j++ {
			if updates[j].Directory == update.Directory &&
				updates[j].PackageEcosystem.String() == update.PackageEcosystem {

				found = true

				updates = append(updates[:j], updates[j+1:]...)
			}
		}

		if !found {
			log.Debugw("Remove", "update", update)
			updatesNode.Content = append(updatesNode.Content[:i], updateNode.Content[i+1:]...)
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

		node, nodeErr := update.node()
		if nodeErr != nil {
			err = Trace(nodeErr)
			return
		}

		log.Debugw("Add", "update", update)
		updatesNode.Content = append(updatesNode.Content, node)
	}

	return
}
