package deps

import (
	"fmt"

	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/group"
)

func BuildGroups(groupUpdates chan *group.Group, conf *internal.ConditionalRebootConfig) ([]*group.Group, error) {
	var groups []*group.Group

	for _, groupConf := range conf.Groups {
		group, err := BuildGroup(groupUpdates, &groupConf)
		if err != nil {
			return nil, fmt.Errorf("could not build group '%s': %w", groupConf.Name, err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}
