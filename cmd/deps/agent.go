package deps

import (
	"fmt"
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/agent"
	"github.com/soerenschneider/conditional-reboot/internal/agent/preconditions"
	"github.com/soerenschneider/conditional-reboot/internal/checkers"
)

func BuildAgent(c *internal.AgentConf) (*agent.StatefulAgent, error) {
	checker, err := BuildChecker(c)
	if err != nil {
		return nil, fmt.Errorf("could not build checker: %w", err)
	}

	precondition, err := BuildPrecondition(c)
	if err != nil {
		return nil, fmt.Errorf("could not build precondition: %w", err)
	}

	agent, err := agent.NewAgent(checker, precondition, c)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func BuildChecker(c *internal.AgentConf) (checkers.Checker, error) {
	switch c.CheckerName {
	case checkers.NeedrestartCheckerName:
		return checkers.NewNeedrestartChecker(), nil
	case checkers.FileCheckerName:
		return checkers.FileCheckerFromMap(c.CheckerArgs)
	case checkers.DnsCheckerName:
		return checkers.DnsCheckerFromMap(c.CheckerArgs)
	case checkers.PrometheusName:
		return checkers.PrometheusCheckerFromMap(c.CheckerArgs)
	case checkers.TcpName:
		return checkers.TcpCheckerFromMap(c.CheckerArgs)
	case checkers.IcmpCheckerName:
		return checkers.IcmpCheckerFromMap(c.CheckerArgs)
	}

	return nil, fmt.Errorf("unknown checker: %s", c.CheckerName)
}

func BuildPrecondition(c *internal.AgentConf) (preconditions.Precondition, error) {
	switch c.PreconditionName {
	case preconditions.WindowedPreconditionName:
		return preconditions.WindowPreconditionFromMap(c.PreconditionArgs)
	case preconditions.AlwaysPreconditionName:
		return &preconditions.AlwaysPrecondition{}, nil
	default:
		return &preconditions.AlwaysPrecondition{}, nil
	}
}
