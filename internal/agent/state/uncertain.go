package state

import "github.com/rs/zerolog/log"

type UncertainState struct {
	stateful      Agent
	successStreak int
	failureStreak int
}

func NewUncertainState(agent Agent) *UncertainState {
	return &UncertainState{
		stateful:      agent,
		successStreak: 0,
		failureStreak: 0,
	}
}

func (s *UncertainState) Name() StateName {
	return UncertainStateName
}

func (s *UncertainState) Failure() {
	s.failureStreak += 1

	if s.failureStreak >= s.stateful.StreakUntilRebootState() {
		s.stateful.SetState(&RebootNeeded{stateful: s.stateful})
	}
}

func (s *UncertainState) Success() {
	s.successStreak += 1

	if s.successStreak >= s.stateful.StreakUntilOkState() {
		s.stateful.SetState(&NoRebootNeeded{stateful: s.stateful})
	}
}

func (s *UncertainState) Error(err error) {
	log.Error().Err(err).Msgf("'%s' encountered error", s.stateful.GetName())
	s.stateful.SetState(&ErrorState{stateful: s.stateful})
}
