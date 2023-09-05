package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/agent/preconditions"
	"github.com/soerenschneider/conditional-reboot/internal/agent/state"
	"github.com/soerenschneider/conditional-reboot/internal/checkers"
	"sync"
	"time"
)

type StatefulAgent struct {
	checker       checkers.Checker
	precondition  preconditions.Precondition
	checkInterval time.Duration

	//durationUntilRecovered specifies the duration that the state "recovering" needs to be in to become "healthy" again.
	updateChannel chan state.Agent

	streakUntilOk           int
	streakUntilRebootNeeded int

	state           state.State
	lastStateChange time.Time
	mutex           sync.RWMutex
}

func NewAgent(checker checkers.Checker, precondition preconditions.Precondition, conf *internal.AgentConf) (*StatefulAgent, error) {
	if checker == nil {
		return nil, errors.New("could not build agent: empty checker supplied")
	}

	if precondition == nil {
		return nil, errors.New("could not build agent: empty precondition supplied")
	}

	if conf == nil {
		return nil, errors.New("empty agent conf provided")
	}

	parsedCheckInterval, err := time.ParseDuration(conf.CheckInterval)
	if err != nil {
		return nil, fmt.Errorf("can not parse 'checkInterval' duration string '%s'", conf.CheckInterval)
	}

	if parsedCheckInterval < time.Duration(5)*time.Second {
		return nil, fmt.Errorf("'checkInterval' may not be < 5s")
	}

	if parsedCheckInterval > time.Duration(1)*time.Hour {
		return nil, fmt.Errorf("'checkInterval' may not be > 1h")
	}

	agent := &StatefulAgent{
		checker:                 checker,
		precondition:            precondition,
		checkInterval:           parsedCheckInterval,
		streakUntilRebootNeeded: conf.StreakUntilReboot,
		streakUntilOk:           conf.StreakUntilOk,
		lastStateChange:         time.Time{},
	}

	agent.state, err = state.NewInitialState(agent)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *StatefulAgent) Run(ctx context.Context, stateUpdateChannel chan state.Agent) error {
	if stateUpdateChannel == nil {
		return errors.New("empty channel provided")
	}
	a.updateChannel = stateUpdateChannel

	a.performCheck(ctx)
	ticker := time.NewTicker(a.checkInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			a.performCheck(ctx)
		}
	}
}

func (a *StatefulAgent) performCheck(ctx context.Context) {
	log.Debug().Msgf("performCheck() %s", a.CheckerNiceName())
	if !a.precondition.PerformCheck() {
		log.Debug().Msgf("Precondition not met, not invoking checker %s", a.CheckerNiceName())
		return
	}

	internal.CheckerLastCheck.WithLabelValues(a.checker.Name()).SetToCurrentTime()

	log.Debug().Msgf("IsHealthy() %s", a.CheckerNiceName())
	isHealthy, err := a.checker.IsHealthy(ctx)
	if err != nil {
		log.Debug().Msgf("IsHealthy(), err != nil %s", a.CheckerNiceName())
		a.state.Error(err)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "err").Set(1)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "healthy").Set(0)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "unhealthy").Set(0)
		return
	}

	if isHealthy {
		log.Debug().Msgf("IsHealthy(), isHealthy=true %s", a.CheckerNiceName())
		a.state.Success()
		internal.CheckerState.WithLabelValues(a.checker.Name(), "err").Set(0)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "healthy").Set(1)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "unhealthy").Set(0)
	} else {
		log.Debug().Msgf("IsHealthy(), isHealthy=false %s", a.CheckerNiceName())
		a.state.Failure()
		internal.CheckerState.WithLabelValues(a.checker.Name(), "err").Set(0)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "healthy").Set(0)
		internal.CheckerState.WithLabelValues(a.checker.Name(), "unhealthy").Set(1)
	}
}

func (a *StatefulAgent) SetState(newState state.State) {
	log.Info().Msgf("Updating state for checker '%s' from '%s' -> '%s'", a.checker.Name(), a.state.Name(), newState.Name())

	internal.AgentState.WithLabelValues(string(newState.Name()), a.CheckerNiceName()).Set(1)
	internal.AgentState.WithLabelValues(string(a.state.Name()), a.CheckerNiceName()).Set(0)
	internal.LastStateChange.WithLabelValues(string(a.state.Name()), a.CheckerNiceName()).SetToCurrentTime()

	log.Debug().Msgf("SetState(%s) acquire lock (%s)", newState.Name(), a.CheckerNiceName())
	a.mutex.Lock()
	defer a.mutex.Unlock()
	log.Debug().Msgf("SetState(%s) success (%s)", newState.Name(), a.CheckerNiceName())

	a.lastStateChange = time.Now()
	a.state = newState
	log.Debug().Msgf("Updating channel %s", a.checker.Name())
	a.updateChannel <- a
	log.Debug().Msgf("Updated channel %s", a.checker.Name())
}

func (a *StatefulAgent) String() string {
	return fmt.Sprintf("%s checker=%s, checkInterval=%s, streakUntilOk=%d, streakUntilUnhealhty=%d", a.CheckerNiceName(), a.checker.Name(), a.checkInterval, a.streakUntilOk, a.streakUntilRebootNeeded)
}

func (a *StatefulAgent) CheckerNiceName() string {
	return a.checker.Name()
}

func (a *StatefulAgent) StreakUntilOkState() int {
	return a.streakUntilOk
}

func (a *StatefulAgent) StreakUntilRebootState() int {
	return a.streakUntilRebootNeeded
}

func (a *StatefulAgent) GetState() state.State {
	log.Debug().Msgf("GetState() acquire lock (%s)", a.CheckerNiceName())
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	log.Debug().Msgf("GetState() lock success (%s)", a.CheckerNiceName())

	return a.state
}

func (a *StatefulAgent) GetStateDuration() time.Duration {
	return time.Since(a.lastStateChange)
}

func (a *StatefulAgent) Failure() {
	a.state.Failure()
}

func (a *StatefulAgent) Success() {
	a.state.Success()
}

func (a *StatefulAgent) Error(err error) {
	a.state.Error(err)
}
