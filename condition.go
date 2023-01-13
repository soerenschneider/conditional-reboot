package main

import (
	"context"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
	"math/rand"
	"sync"
	"time"
)

type Condition struct {
	client               v1.API
	name                 string
	query                string
	durationUntilHealthy time.Duration

	mutex   sync.Mutex
	queries int

	recoveringSince time.Time
	queryInterval   time.Duration

	recoveringState State
	unhealthyState  State
	healthyState    State

	currentState State
}

const (
	defaultQueryInterval = time.Second * 30
)

func NewCondition(client v1.API, name, query string, durationUntilHealthy time.Duration) (*Condition, error) {
	cond := &Condition{
		name:                 name,
		client:               client,
		query:                query,
		durationUntilHealthy: durationUntilHealthy,

		mutex:           sync.Mutex{},
		queries:         0,
		recoveringSince: time.Time{},
		queryInterval:   defaultQueryInterval,
	}

	cond.unhealthyState = &UnhealthyState{cond: cond}
	cond.recoveringState = &RecoveringState{cond: cond}
	cond.healthyState = &HealthyState{cond: cond}

	cond.setState(&InitialState{cond: cond})

	return cond, nil
}

func (c *Condition) UpdateCondition(ctx context.Context) {
	result, warnings, err := c.client.Query(ctx, c.query, time.Now(), v1.WithTimeout(5*time.Second))
	if err != nil {
		c.QueryError(err)
		time.Sleep(c.queryInterval)
		return
	}

	if len(warnings) > 0 {
		log.Warn().Msgf("Warnings: %v\n", warnings)
	}

	vec := result.(model.Vector)
	log.Debug().Msgf("%v", vec)
	if len(vec) > 0 {
		c.QuerySuccess()
	} else {
		c.QueryError(nil)
	}
}

func (c *Condition) Run(ctx context.Context) {
	ticker := time.NewTicker(c.queryInterval)
	c.UpdateCondition(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("%s received signal, packing it up", c.name)
			return
		case <-ticker.C:
			// Try to slightly distribute reads
			rand.Seed(time.Now().UTC().UnixNano())
			sleepDuration := time.Millisecond * time.Duration(rand.Intn(2500))
			log.Debug().Msgf("Sleeping for %v", sleepDuration)
			time.Sleep(sleepDuration)

			c.UpdateCondition(ctx)
		}
	}

	return
}

func (c *Condition) getState() State {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentState
}

func (c *Condition) setState(state State) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.currentState != nil {
		log.Info().Msgf("State for '%s' changed from %s to %s", c.name, c.currentState.Name(), state.Name())
	}
	MetricStatus.WithLabelValues(c.name, state.Name()).Set(1)
	c.currentState = state
}

func (c *Condition) QuerySuccess() {
	if c.queries > 0 && c.queries%5 == 0 {
		log.Info().Msgf("Trying %d times", c.queries)
	}
	c.queries++
	c.currentState.SuccessfulEvaluation()
}

func (c *Condition) QueryError(err error) {
	if err != nil {
		log.Error().Msgf("query error for name '%s': %v", c.name, err)
	}
	if c.queries > 0 && c.queries%5 == 0 {
		log.Info().Msgf("Trying %d times", c.queries)
	}
	c.queries++
	c.currentState.QueryError(err)
}
