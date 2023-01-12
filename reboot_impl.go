package main

type Reboot interface {
	Reboot() error
}

type NoReboot struct {
}

func (d *NoReboot) Reboot() error {
	return nil
}
