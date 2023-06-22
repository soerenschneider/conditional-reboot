package internal

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/internal/agent/state"
)

const (
	defaultStreakUntilOk      = 3
	defaultStreakUntilReboot  = 1
	defaultStateEvaluatorName = "or"
)

var (
	defaultStateEvaluatorArgs = map[string]string{
		string(state.RebootStateName): "0s",
	}
)

type ConditionalRebootConfig struct {
	Groups            []*GroupConf `json:"groups" validate:"dive,required"`
	JournalFile       string       `json:"journal_file" validate:"omitempty,filepath"`
	MetricsListenAddr string       `json:"metrics_listen_addr" validate:"excluded_with=MetricsDir"`
	MetricsDir        string       `json:"metrics_dir" validate:"excluded_with=MetricsListenAddr,omitempty,dirpath"`
}

func (conf *ConditionalRebootConfig) Print() {
	log.Info().Msg("Active config values:")
	for _, group := range conf.Groups {
		log.Info().Msgf("Group '%s', stateEvaluator='%s', stateEvaluatorArgs=%v", group.Name, group.StateEvaluatorName, group.StateEvaluatorArgs)
		for _, agent := range group.Agents {
			log.Info().Msgf("--> Agent '%s', checkerArgs=%v, precondition='%s', preconditionArgs=%v, streakUntilRecovered=%d, streakUntilUnhealthy=%d", agent.CheckerName, agent.CheckerArgs, agent.PreconditionName, agent.PreconditionArgs, agent.StreakUntilOk, agent.StreakUntilReboot)
		}
	}
}

type GroupConf struct {
	Agents []*AgentConf `json:"agents" validate:"dive"`
	Name   string       `json:"name" validate:"required"`

	StateEvaluatorName string            `json:"state_evaluator_name"`
	StateEvaluatorArgs map[string]string `json:"state_evaluator_args" validate:"required"`
}

func (conf *GroupConf) UnmarshalJSON(data []byte) error {
	type Alias GroupConf // Create an alias to avoid recursion during unmarshalling

	// Define conf temporary struct with default values
	tmp := &Alias{
		StateEvaluatorName: defaultStateEvaluatorName,
		StateEvaluatorArgs: defaultStateEvaluatorArgs,
	}

	// Unmarshal the JSON data into the temporary struct
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// Assign the values from the temporary struct to the original struct
	*conf = GroupConf(*tmp)
	return nil
}

type AgentConf struct {
	CheckInterval     string `json:"check_interval" validate:"required"`
	StreakUntilOk     int    `json:"streak_until_ok" validate:"required,gte=1"`
	StreakUntilReboot int    `json:"streak_until_reboot" validate:"gte=1"`

	CheckerName string         `json:"checker_name" validate:"required"`
	CheckerArgs map[string]any `json:"checker_args"`

	PreconditionName string         `json:"precondition_name"`
	PreconditionArgs map[string]any `json:"precondition_args"`
}

func (conf *AgentConf) UnmarshalJSON(data []byte) error {
	type Alias AgentConf // Create an alias to avoid recursion during unmarshalling

	// Define a temporary struct with default values
	tmp := &Alias{
		StreakUntilReboot: defaultStreakUntilReboot,
		StreakUntilOk:     defaultStreakUntilOk,
	}

	// Unmarshal the JSON data into the temporary struct
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// Assign the values from the temporary struct to the original struct
	*conf = AgentConf(*tmp)
	return nil
}
