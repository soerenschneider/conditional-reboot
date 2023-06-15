package state

type ErrorState struct {
	stateful Agent
}

func (s *ErrorState) Name() StateName {
	return ErrorStateName
}

func (s *ErrorState) Success() {
	newState := NewUncertainState(s.stateful)
	s.stateful.SetState(newState)
}

func (s *ErrorState) Failure() {
	s.stateful.SetState(&RebootNeeded{stateful: s.stateful})
}

func (s *ErrorState) Error(err error) {
}
