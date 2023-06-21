package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/cmd/deps"
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/group"
	"github.com/soerenschneider/conditional-reboot/pkg/reboot"
	"net/http"
	"os"
	"time"
)

const defaultConfigFile = "/etc/conditional-reboot.json"

var (
	debug           bool
	dryRun          bool
	cmdVerifyReboot bool
	cmdPrintVersion bool
	configFile      string
	rebootImpl      reboot.Reboot
)

func main() {
	parseFlags()

	if cmdPrintVersion {
		printVersion()
	}

	var err error
	rebootImpl, err = deps.BuildRebootImpl(dryRun)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build reboot impl")
	}

	if cmdVerifyReboot {
		verifyReboot()
	}

	runApp()
}

func parseFlags() {
	flag.StringVar(&configFile, "config", defaultConfigFile, "read configuration from specified config file")
	flag.BoolVar(&debug, "debug", false, "print debug statements")
	flag.BoolVar(&dryRun, "dry-run", false, "don't actually reboot the system, only test configuration")
	flag.BoolVar(&cmdVerifyReboot, "verify-reboot", false, "test reboot implementation. CAUTION: this will try to reboot your system")
	flag.BoolVar(&cmdPrintVersion, "version", false, "print version and exit")

	flag.Parse()

	if dryRun && cmdVerifyReboot {
		log.Fatal().Msg("can not specify both 'dry-run' and 'verify-reboot'")
	}

	if cmdPrintVersion && cmdVerifyReboot {
		log.Fatal().Msg("can not specify both 'verify-reboot' and 'version' commands")
	}
}

func printVersion() {
	fmt.Println(internal.BuildVersion)
	os.Exit(0)
}

func verifyReboot() {
	sleepTime := 15 * time.Second
	log.Warn().Msg("Verifying whether reboot works. This will (most-likely) reboot your machine.")
	log.Warn().Msgf("You have %s to cancel this.", sleepTime)
	time.Sleep(sleepTime)
	if err := rebootImpl.Reboot(); err != nil {
		log.Fatal().Err(err).Msg("Reboot returned error")
	}
	log.Fatal().Msg("Reboot apparently did not work as expected")
}

func runApp() {
	initLogging()
	log.Info().Msgf("conditional-reboot %s", internal.BuildVersion)

	appConfig, err := readConfig(configFile)
	if err != nil {
		log.Fatal().Err(err).Msgf("could not read config file '%s'", configFile)
	}
	appConfig.Print()
	if err := internal.ValidateConfig(appConfig); err != nil {
		log.Fatal().Err(err).Msg("config could not be validated")
	}
	groupUpdates := make(chan *group.Group, 1)

	groups, err := deps.BuildGroups(groupUpdates, appConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build groups")
	}

	audit, err := deps.BuildAudit(appConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build journal impl")
	}

	app, err := internal.NewConditionalReboot(groups, rebootImpl, audit, groupUpdates)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build conditional-reboot app")
	}

	if len(appConfig.MetricsListenAddr) > 0 {
		go func() {
			log.Info().Msgf("Starting metrics server at '%s'", appConfig.MetricsListenAddr)
			err := internal.StartMetricsServer(appConfig.MetricsListenAddr)
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("could not start metrics server")
			}
		}()
	}

	log.Info().Msg("Starting agents...")
	app.Start()
}

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func readConfig(file string) (*internal.ConditionalRebootConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var ret internal.ConditionalRebootConfig
	if err = json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}
