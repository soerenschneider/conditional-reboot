package group

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/internal/agent/state"
	"github.com/soerenschneider/conditional-reboot/internal/group/state_evaluator"
	"time"
)

const tickerInterval = 15 * time.Second

type Group struct {
	agents               []state.Agent
	stateEvaluator       state_evaluator.StateEvaluator
	agentStateUpdateChan chan state.Agent
	rebootRequests       chan *Group
	name                 string
}

func NewGroup(name string, agents []state.Agent, stateEvaluator state_evaluator.StateEvaluator, rebootRequests chan *Group) (*Group, error) {
	if len(name) == 0 {
		return nil, errors.New("could not build group: empty name provided")
	}

	if len(agents) == 0 {
		return nil, errors.New("could not build group: empty agents provided")
	}

	if rebootRequests == nil {
		return nil, errors.New("could not build group: nil channel provided")
	}

	agentUpdates := make(chan state.Agent)

	return &Group{
		name:                 name,
		agents:               agents,
		stateEvaluator:       stateEvaluator,
		agentStateUpdateChan: agentUpdates,
		rebootRequests:       rebootRequests,
	}, nil
}

func (g *Group) GetName() string {
	return g.name
}

func (g *Group) Agents() []state.Agent {
	return g.agents
}

func (g *Group) Start(ctx context.Context) {
	for _, agent := range g.agents {
		go func(a state.Agent) {
			if err := a.Run(ctx, g.agentStateUpdateChan); err != nil {
				log.Fatal().Err(err).Msgf("could start agent %s", a.GetName())
			}
		}(agent)
	}

	ticker := time.NewTicker(tickerInterval)

	go func() {
		for {
			select {
			case agent := <-g.agentStateUpdateChan:
				log.Info().Msgf("Received update from agent %s", agent.GetName())
				if g.stateEvaluator.ShouldReboot(g) {
					g.rebootRequests <- g
				}
			case <-ticker.C:
				if g.stateEvaluator.ShouldReboot(g) {
					g.rebootRequests <- g
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
