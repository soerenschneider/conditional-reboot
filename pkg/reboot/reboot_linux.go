package reboot

import "syscall"

type DefaultRebootImpl struct {
}

func (l *DefaultRebootImpl) Reboot() error {
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	return nil
}
