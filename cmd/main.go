package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/cmd/deps"
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/app"
	"github.com/soerenschneider/conditional-reboot/internal/config"
	"github.com/soerenschneider/conditional-reboot/internal/group"
	"github.com/soerenschneider/conditional-reboot/pkg/reboot"
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
	considerationTime := 15 * time.Second
	log.Warn().Msg("Verifying whether reboot works. This will (most-likely) reboot your machine.")
	log.Warn().Msgf("You have %s to abort by pressing CTRL+C.", considerationTime)
	time.Sleep(considerationTime)
	if err := rebootImpl.Reboot(); err != nil {
		log.Fatal().Err(err).Msg("Reboot returned error")
	}
	log.Fatal().Msg("Reboot apparently did not work as expected")
}

func runApp() {
	internal.ProcessStartTime.SetToCurrentTime()
	internal.Version.WithLabelValues(internal.BuildVersion).Set(1)

	initLogging()
	log.Info().Msgf("conditional-reboot %s", internal.BuildVersion)

	appConfig, err := config.ReadConfig(configFile)
	if err != nil {
		log.Fatal().Err(err).Msgf("could not read config file '%s'", configFile)
	}
	appConfig.Print()
	if err := config.Validate(appConfig); err != nil {
		log.Fatal().Err(err).Msg("config could not be validated")
	}
	groupUpdates := make(chan *group.Group, 1)

	groups, err := deps.BuildGroups(groupUpdates, appConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build groups")
	}

	journal, err := deps.BuildAudit(appConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build journal impl")
	}

	opts := []app.ConditionalRebootOpts{app.UseJournal(journal)}
	app, err := app.NewConditionalReboot(groups, rebootImpl, groupUpdates, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build conditional-reboot app")
	}

	go internal.StartHeartbeat(context.Background())

	if len(appConfig.MetricsListenAddr) > 0 {
		go func() {
			log.Info().Msgf("Starting metrics server at '%s'", appConfig.MetricsListenAddr)
			err := internal.StartMetricsServer(appConfig.MetricsListenAddr)
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("could not start metrics server")
			}
		}()
	} else if len(appConfig.MetricsDir) > 0 {
		go internal.StartMetricsDumper(context.Background(), appConfig.MetricsDir)
	} else {
		log.Warn().Msg("Neither metrics server nor metrics dumping configured")
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
