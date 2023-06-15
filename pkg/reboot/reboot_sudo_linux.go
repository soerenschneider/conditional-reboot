package reboot

import (
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
)

type SystemctlReboot struct {
}

func (l *SystemctlReboot) Reboot() error {
	uid := os.Getuid()

	if uid == 0 {
		log.Info().Msg("Running as root, attempting direct reboot...")

		cmd := exec.Command("systemctl", "reboot")
		err := cmd.Run()
		if err != nil {
			return err
		}

		return nil
	}

	log.Info().Msg("Not running as root, rebooting system via sudo... ")
	cmd := exec.Command("sudo", "systemctl", "reboot")
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
