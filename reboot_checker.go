package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	AlwaysRebootName = "Always"
	NeedrestartName  = "NeedrestartChecker"
)

type RebootNeededChecker interface {
	NeedsReboot() (bool, error)
	Name() string
}

type AlwaysReboot struct{}

func (r *AlwaysReboot) NeedsReboot() (bool, error) {
	return true, nil
}

func (r *AlwaysReboot) Name() string {
	return AlwaysRebootName
}

// NeedrestartChecker uses https://github.com/liske/needrestart to check whether rebooting is needed
type NeedrestartChecker struct{}

func (n *NeedrestartChecker) Name() string {
	return NeedrestartName
}

func (n *NeedrestartChecker) NeedsReboot() (bool, error) {
	out, err := exec.Command("needrestart", "-b").Output()
	if err != nil {
		return false, fmt.Errorf("could not determine if reboot is needed: %v", err)
	}

	re := regexp.MustCompile(`NEEDRESTART-KSTA: (?P<ksta>\d)`)

	// check for updated kernel
	ma := re.FindStringSubmatch(string(out))
	var kernelUpdate, svcUpdates bool
	if len(ma) > 0 {
		val, err := strconv.Atoi(ma[1])
		if err != nil {
			log.Error().Msgf("could not parse 'NEEDRESTART-KSTA': %v", err)
		} else {
			kernelUpdate = val > 1
			log.Info().Msg("Kernel update detected")
		}
	}

	// check for service upgrades
	if strings.Contains(string(out), "NEEDRESTART-SVC:") {
		svcUpdates = true
		log.Info().Msg("Service updates detected")
	}

	return kernelUpdate || svcUpdates, nil
}
