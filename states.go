package main

import (
	"time"
)

const (
	HealthyStateName    = "Healthy"
	RecoveringStateName = "Recovering"
	UnhealthyStateName  = "Unhealthy"
	InitialStateName    = "Initial"
)

type State interface {
	SuccessfulEvaluation()
	QueryError(err error)
	Name() string
}

type HealthyState struct {
	cond *Condition
}

func (s *HealthyState) Name() string {
	return HealthyStateName
}

func (s *HealthyState) SuccessfulEvaluation() {
}

func (s *HealthyState) QueryError(err error) {
	s.cond.recoveringSince = time.Time{}
	s.cond.setState(s.cond.unhealthyState)
}

type RecoveringState struct {
	cond *Condition
}

func (s *RecoveringState) Name() string {
	return RecoveringStateName
}

func (s *RecoveringState) SuccessfulEvaluation() {
	if !s.cond.recoveringSince.IsZero() && time.Now().After(s.cond.recoveringSince.Add(s.cond.durationUntilHealthy)) {
		s.cond.setState(s.cond.healthyState)
	}
}

func (s *RecoveringState) QueryError(err error) {
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

type InitialState struct {
	cond *Condition
}

func (s *InitialState) Name() string {
	return InitialStateName
}

func (s *InitialState) SuccessfulEvaluation() {
	s.cond.recoveringSince = time.Now()
	s.cond.setState(s.cond.recoveringState)
}

func (s *InitialState) QueryError(err error) {
	s.cond.setState(s.cond.unhealthyState)
}
