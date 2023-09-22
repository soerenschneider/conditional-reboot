package deps

import (
	"fmt"

	"github.com/soerenschneider/conditional-reboot/internal/config"
	"github.com/soerenschneider/conditional-reboot/internal/group"
)

func BuildGroups(groupUpdates chan *group.Group, conf *config.ConditionalRebootConfig) ([]*group.Group, error) {
	var groups []*group.Group

	for _, groupConf := range conf.Groups {
		groupConf := groupConf
		group, err := BuildGroup(groupUpdates, &groupConf)
		if err != nil {
			return nil, fmt.Errorf("could not build group '%s': %w", groupConf.Name, err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}
