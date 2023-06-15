package deps

import (
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/agent/state"
	"github.com/soerenschneider/conditional-reboot/internal/group"
	"github.com/soerenschneider/conditional-reboot/internal/group/state_evaluator"
	"strings"
)

func BuildGroup(groupUpdates chan *group.Group, conf *internal.GroupConf) (*group.Group, error) {
	agents, err := BuildAgents(conf)
	if err != nil {
		return nil, err
	}

	evaluator, err := BuildStateEvaluator(conf)
	if err != nil {
		return nil, err
	}

	group, err := group.NewGroup(conf.Name, agents, evaluator, groupUpdates)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func BuildAgents(conf *internal.GroupConf) ([]state.Agent, error) {
	var agents []state.Agent
	for _, agentConf := range conf.Agents {
		agent, err := BuildAgent(agentConf)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func BuildStateEvaluator(conf *internal.GroupConf) (state_evaluator.StateEvaluator, error) {
	switch strings.ToLower(conf.StateEvaluatorName) {
	case state_evaluator.StateCheckerAndName:
		return state_evaluator.NewStateCheckerAnd(conf.StateEvaluatorArgs)
	}

	return state_evaluator.NewStateCheckerOr(conf.StateEvaluatorArgs)
}
