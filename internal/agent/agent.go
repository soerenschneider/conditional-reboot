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
	mutex           sync.Mutex
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
	internal.CheckerLastCheck.WithLabelValues(a.checker.Name()).SetToCurrentTime()

	if !a.precondition.PerformCheck() {
		log.Debug().Msgf("Precondition not met, not invoking checker %s", a.CheckerNiceName())
		return
	}

	isHealthy, err := a.checker.IsHealthy(ctx)
	if err != nil {
		a.state.Error(err)
		return
	}

	if isHealthy {
		a.state.Success()
	} else {
		a.state.Failure()
	}
}

func (a *StatefulAgent) SetState(newState state.State) {
	log.Info().Msgf("Updating state for checker '%s' from '%s' -> '%s'", a.checker.Name(), a.state.Name(), newState.Name())

	internal.AgentState.WithLabelValues(string(newState.Name()), a.CheckerNiceName()).Set(1)
	internal.AgentState.WithLabelValues(string(a.state.Name()), a.CheckerNiceName()).Set(0)
	internal.LastStateChange.WithLabelValues(string(a.state.Name()), a.CheckerNiceName()).SetToCurrentTime()

	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.lastStateChange = time.Now()
	a.state = newState
	a.updateChannel <- a
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
	a.mutex.Lock()
	defer a.mutex.Unlock()

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
