package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	defaultEvaluationPeriod   = 5 * time.Second
	defaultConfigFileLocation = "/etc/conditional-reboot.json"
	defaultMetricsFile        = "/var/lib/node_exporter/conditional_reboot.prom"
)

var (
	flagDebug      *bool
	flagDryRun     *bool
	flagConfigFile *string

	defaultHttpclient = retryablehttp.NewClient()
)

type ConditionalReboot struct {
	timeout    time.Time
	startTime  time.Time
	conditions []*Condition
	rebootImpl Reboot
	conf       *Conf

	evaluations int
}

func NewConditionalReboot(conditions []*Condition, rebootImpl Reboot, conf *Conf) (*ConditionalReboot, error) {
	if nil == conditions {
		return nil, errors.New("empty conditions provided")
	}

	if nil == rebootImpl {
		return nil, errors.New("empty rebootImpl implementation provided")
	}

	if nil == conf {
		return nil, errors.New("empty conf supplied")
	}

	timeout := time.Second * time.Duration(conf.ConditionsTimeoutSeconds)
	if conf.ConditionsTimeoutSeconds < 120 {
		log.Warn().Msgf("Ignoring supplied timeout of %ds, using default of %v", conf.ConditionsTimeoutSeconds, defaultWaitOnConditionsTimeout)
		timeout = time.Second * defaultWaitOnConditionsTimeout
	}
	MetricStartTime.SetToCurrentTime()

	return &ConditionalReboot{
		timeout:    time.Now().Add(timeout),
		startTime:  time.Now(),
		conditions: conditions,
		rebootImpl: rebootImpl,
		conf:       conf,
	}, nil
}

func (m *ConditionalReboot) evaluate() {
	stateDict := map[string]string{}
	for _, condition := range m.conditions {
		curState := condition.getState()
		stateDict[condition.name] = curState.Name()
	}

	healthyStateCnt := 0
	for _, state := range stateDict {
		if state == HealthyStateName {
			healthyStateCnt++
		}
	}
	if healthyStateCnt == len(m.conditions) {
		MetricSuccess.Set(1)
		HandleMetrics(m.conf)
		log.Info().Msgf("All %d conditions are healthy, rebooting", len(m.conditions))
		err := m.rebootImpl.Reboot()
		if err != nil {
			MetricSuccess.Set(0)
			HandleMetrics(m.conf)
			log.Fatal().Err(err).Msg("Reboot system failed")
			os.Exit(1)
		}

		// call os.Exit for reboot implementations that do not actually reboot the system
		os.Exit(0)
	}

	if m.evaluations%12 == 0 {
		HandleMetrics(m.conf)
		log.Info().Msgf("%d/%d conditions are healthy after %s", healthyStateCnt, len(m.conditions), time.Now().Sub(m.startTime))
	}
	if time.Now().After(m.timeout) {
		HandleMetrics(m.conf)
		log.Fatal().Msgf("Did not reach healthy state within %s", time.Now().Sub(m.startTime))
	}
	m.evaluations++
}

func (m *ConditionalReboot) Run() {
	wg := &sync.WaitGroup{}
	wg.Add(len(m.conditions))
	ctx, cancel := context.WithCancel(context.Background())
	for _, condition := range m.conditions {
		go condition.Run(ctx, wg)
	}

	ticker := time.NewTicker(defaultEvaluationPeriod)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	m.evaluate()
	for {
		select {
		case <-ticker.C:
			m.evaluate()
		case <-quit:
			cancel()
			log.Info().Msg("Caught signal from user, interrupting")
			wg.Wait()
			return
		}
	}
}

func parseFlags() {
	flagDebug = flag.Bool("debug", false, "sets log level to debug")
	flagConfigFile = flag.String("config", defaultConfigFileLocation, "config file to use")
	flagDryRun = flag.Bool("dry-run", false, "do not perform any action")

	version := flag.Bool("version", false, "Print version info and exit")

	flag.Parse()

	if *version {
		fmt.Println(BuildVersion)
		os.Exit(0)
	}

}

type MetricsHook struct {
	conf *Conf
}

func (h MetricsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.FatalLevel {
		HandleMetrics(h.conf)
		MetricSuccess.Set(0)
	}
}

func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	defaultHttpclient.Logger = &log.Logger

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *flagDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func main() {
	parseFlags()
	initLogging()
	log.Info().Msgf("conditional-reboot version %s", BuildVersion)

	conf, err := read(*flagConfigFile)
	if err != nil {
		WriteMetrics(defaultMetricsFile)
		log.Fatal().Err(err).Msg("Reading config file failed")
	}

	defer HandleMetrics(conf)
	log := log.Hook(MetricsHook{conf: conf})

	restartChecker, err := conf.GetRebootNeededChecker()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not build restart checker")
	}
	log.Info().Msgf("Using restart checker '%s'", restartChecker.Name())

	conditions, err := conf.BuildConditions()
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing conditions")
	}
	MetricConditions.Set(float64(len(conditions)))

	var rebootImpl Reboot = &DefaultRebootImpl{}
	if *flagDryRun {
		log.Info().Msg("Dry-run mode active, not going to reboot system")
		rebootImpl = &NoReboot{}
	}
	m, err := NewConditionalReboot(conditions, rebootImpl, conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not build command center")
	}

	restartNecessary, err := restartChecker.NeedsReboot()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not reliably detect whether restart is needed")
	}
	if restartNecessary {
		MetricRebootNeeded.Set(1)
		log.Info().Msg("Reboot is needed, checking if restart conditions are met")
		m.Run()
	} else {
		MetricRebootNeeded.Set(0)
		log.Info().Msg("No restart needed, quitting")
	}
	MetricSuccess.Set(1)
}
