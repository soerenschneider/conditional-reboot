package state

type NoRebootNeeded struct {
	stateful Agent
}

func (s *NoRebootNeeded) Name() StateName {
	return OkStateName
}

func (s *NoRebootNeeded) Success() {
}

func (s *NoRebootNeeded) Failure() {
	s.stateful.SetState(&RebootNeeded{stateful: s.stateful})
}

func (s *NoRebootNeeded) Error(err error) {
	s.stateful.SetState(&ErrorState{stateful: s.stateful})
}
