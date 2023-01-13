package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"os"
	"strings"
	"time"
)

const (
	defaultWaitOnConditionsTimeout = 4 * 60 * 60
	defaultSecondsUntilHealthy     = 120
)

type Query struct {
	Name                string `json:"name"`
	Expression          string `json:"expression"`
	SecondsUntilHealthy int    `json:"seconds_until_healthy,omitempty"`
}

type PrometheusConditions struct {
	Address string  `json:"prom_address"`
	Queries []Query `json:"queries"`
}

type Conf struct {
	Conditions               []PrometheusConditions `json:"conditions"`
	ConditionsTimeoutSeconds int                    `json:"conditions_timeout_s,omitempty"`
	RebootCheckStrategy      string                 `json:"reboot_check_strategy,omitempty"`
	PushgatewayUrl           string                 `json:"pushgateway_url,omitempty"`
}

func (c *Conf) BuildConditions() ([]*Condition, error) {
	var conditions []*Condition = []*Condition{}
	for _, crit := range c.Conditions {
		client, err := api.NewClient(api.Config{
			Address: crit.Address,
			Client:  defaultHttpclient.StandardClient(),
		})
		if err != nil {
			return nil, fmt.Errorf("could not build prometheus client: %w", err)
		}

		api := v1.NewAPI(client)
		for _, query := range crit.Queries {
			if query.SecondsUntilHealthy < defaultSecondsUntilHealthy {
				query.SecondsUntilHealthy = defaultSecondsUntilHealthy
			}

			durationUntilHealthy := time.Second * time.Duration(query.SecondsUntilHealthy)
			condition, err := NewCondition(api, query.Name, query.Expression, durationUntilHealthy)
			if err != nil {
				return nil, fmt.Errorf("could not build criteria for vault %s: %w", crit.Address, err)
			}
			conditions = append(conditions, condition)
		}
	}

	return conditions, nil
}

func (c *Conf) GetRebootNeededChecker() (RebootNeededChecker, error) {
	switch strings.ToLower(c.RebootCheckStrategy) {
	case strings.ToLower(NeedrestartName):
		return &NeedrestartChecker{}, nil
	case strings.ToLower(AlwaysRebootName):
		return &AlwaysReboot{}, nil
	default:
		return nil, fmt.Errorf("unknown reboot checker: %s", c.RebootCheckStrategy)
	}
}

func read(filename string) (*Conf, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read config from file: %v", err)
	}

	conf := getDefaultConfig()
	err = json.Unmarshal(content, conf)
	return conf, err
}

func getDefaultConfig() *Conf {
	return &Conf{
		ConditionsTimeoutSeconds: defaultWaitOnConditionsTimeout,
		RebootCheckStrategy:      AlwaysRebootName,
	}
}
