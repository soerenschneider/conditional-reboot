package deps

import "github.com/soerenschneider/conditional-reboot/pkg/reboot"

func BuildRebootImpl(dryRun bool) (reboot.Reboot, error) {
	if dryRun {
		return &reboot.NoReboot{}, nil
	}

	return &reboot.DefaultRebootImpl{}, nil
}
