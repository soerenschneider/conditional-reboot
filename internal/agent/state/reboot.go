package state

type RebootNeeded struct {
	stateful Agent
}

func (s *RebootNeeded) Name() StateName {
	return RebootStateName
}

func (s *RebootNeeded) Success() {
	newState := NewUncertainState(s.stateful)
	s.stateful.SetState(newState)
}

func (s *RebootNeeded) Failure() {
}

func (s *RebootNeeded) Error(err error) {
	s.stateful.SetState(&ErrorState{stateful: s.stateful})
}
