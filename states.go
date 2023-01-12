package main

import "time"

const (
	HealthyStateName    = "Healthy"
	RecoveringStateName = "Recovering"
	UnhealthyStateName  = "Unhealthy"
)

type State interface {
	SuccessfulEvaluation()
	QueryError(err error)
	Name() string
}

type HealthyStatus struct {
	cond *Condition
}

func (s *HealthyStatus) Name() string {
	return HealthyStateName
}

func (s *HealthyStatus) SuccessfulEvaluation() {
}

func (s *HealthyStatus) QueryError(err error) {
	s.cond.recoveringSince = time.Time{}
	s.cond.setState(s.cond.unhealthyState)
}

type RecoveringStatus struct {
	cond *Condition
}

func (s *RecoveringStatus) Name() string {
	return RecoveringStateName
}

func (s *RecoveringStatus) SuccessfulEvaluation() {
	if time.Now().After(s.cond.recoveringSince.Add(s.cond.durationUntilHealthy)) {
		s.cond.setState(s.cond.healthyState)
	}
}

func (s *RecoveringStatus) QueryError(err error) {
	s.cond.recoveringSince = time.Time{}
	s.cond.setState(s.cond.unhealthyState)
}

type UnhealthyState struct {
	cond *Condition
}

func (s *UnhealthyState) Name() string {
	return UnhealthyStateName
}

func (s *UnhealthyState) SuccessfulEvaluation() {
	s.cond.recoveringSince = time.Now()
	s.cond.setState(s.cond.recoveringState)
}

func (s *UnhealthyState) QueryError(err error) {
}
