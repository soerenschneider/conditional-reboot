package state_evaluator

import (
	"github.com/soerenschneider/conditional-reboot/internal/agent/state"
	"time"
)

const StateCheckerAndName = "and"

type StateCheckerAnd struct {
	wants map[state.StateName]time.Duration
}

func NewStateCheckerAnd(args map[string]string) (*StateCheckerAnd, error) {
	parsed, err := parseArgsMap(args)
	if err != nil {
		return nil, err
	}
	return &StateCheckerAnd{wants: parsed}, nil
}

func (r *StateCheckerAnd) ShouldReboot(group Group) bool {
	for _, agent := range group.Agents() {
		if !r.CheckAgent(agent) {
			return false
		}
	}

	return true
}

func (r *StateCheckerAnd) CheckAgent(agent state.Agent) bool {
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
