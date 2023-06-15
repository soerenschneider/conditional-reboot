package checkers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const NeedrestartCheckerName = "needrestart"

var regex = regexp.MustCompile(`NEEDRESTART-KSTA: (?P<ksta>\d)`)

type Needrestart interface {
	Result(ctx context.Context) (string, error)
}

type NeedrestartCmd struct{}

func (n *NeedrestartCmd) Result(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "needrestart", "-b").Output()
	if err != nil {
		return "", fmt.Errorf("could not determine if reboot is needed: %w", err)
	}

	return string(out), nil
}

// NeedrestartChecker uses https://github.com/liske/needrestart to check whether rebooting is needed
type NeedrestartChecker struct {
	rebootNeeded bool
	sync         sync.Mutex
	needrestart  Needrestart
}

func NewNeedrestartChecker() *NeedrestartChecker {
	return &NeedrestartChecker{
		sync:        sync.Mutex{},
		needrestart: &NeedrestartCmd{},
	}
}

func (n *NeedrestartChecker) Name() string {
	return NeedrestartCheckerName
}

func (n *NeedrestartChecker) IsHealthy(ctx context.Context) (bool, error) {
	n.sync.Lock()
	defer n.sync.Unlock()

	// use cached reply
	if n.rebootNeeded {
		return false, nil
	}

	out, err := n.needrestart.Result(ctx)
	if err != nil {
		return false, err
	}

	kernelUpdate, svcUpdates := n.detectUpdates(out)

	// cache response - we won't recover from a needed reboot until we actually reboot
	n.rebootNeeded = kernelUpdate || svcUpdates

	// reboot is needed, report unhealthy status
	if n.rebootNeeded {
		return false, nil
	}
	return true, nil
}

func (n *NeedrestartChecker) detectUpdates(out string) (bool, bool) {
	// check for updated kernel
	ma := regex.FindStringSubmatch(out)
	var kernelUpdate, svcUpdates bool
	if len(ma) > 0 {
		val, err := strconv.Atoi(ma[1])
		if err != nil {
			log.Error().Msgf("could not parse 'NEEDRESTART-KSTA': %v", err)
		} else if val > 1 {
			kernelUpdate = true
			log.Info().Msg("Kernel update detected")
		}
	}

	// check for service upgrades
	if strings.Contains(out, "NEEDRESTART-SVC:") {
		svcUpdates = true
		log.Info().Msg("Service updates detected")
	}

	return kernelUpdate, svcUpdates
}
