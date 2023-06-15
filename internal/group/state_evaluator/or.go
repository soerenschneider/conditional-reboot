package state_evaluator

import (
	"github.com/soerenschneider/conditional-reboot/internal/agent/state"
	"time"
)

const StateCheckerOrName = "or"

type StateCheckerOr struct {
	wants map[state.StateName]time.Duration
}

func NewStateCheckerOr(args map[string]string) (*StateCheckerOr, error) {
	parsed, err := parseArgsMap(args)
	if err != nil {
		return nil, err
	}

	return &StateCheckerOr{wants: parsed}, nil
}

func (r *StateCheckerOr) ShouldReboot(group Group) bool {
	for _, agent := range group.Agents() {
		if r.CheckAgent(agent) {
			return true
		}
	}

	return false
}

func (r *StateCheckerOr) CheckAgent(agent state.Agent) bool {
	currentState := agent.GetState().Name()
	for wantedType, wantedFor := range r.wants {
		if currentState == wantedType {
			if agent.GetStateDuration() >= wantedFor {
				return true
			}
		}
	}

	return false
}
