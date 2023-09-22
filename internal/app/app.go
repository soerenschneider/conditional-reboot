package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/group"
	"github.com/soerenschneider/conditional-reboot/internal/journal"
	"github.com/soerenschneider/conditional-reboot/internal/uptime"
	"github.com/soerenschneider/conditional-reboot/pkg/reboot"
	"go.uber.org/multierr"
)

const defaultSafeMinimumSystemUptime = 4 * time.Hour

type ConditionalReboot struct {
	groups        []*group.Group
	rebootImpl    reboot.Reboot
	audit         journal.Journal
	rebootRequest chan *group.Group

	safeMinSystemUptime time.Duration
}

type ConditionalRebootOpts func(c *ConditionalReboot) error

func NewConditionalReboot(groups []*group.Group, rebootImpl reboot.Reboot, rebootReq chan *group.Group, opts ...ConditionalRebootOpts) (*ConditionalReboot, error) {
	if len(groups) == 0 {
		return nil, errors.New("no groups provided")
	}

	if rebootImpl == nil {
		return nil, errors.New("no reboot impl provided")
	}

	if rebootReq == nil {
		return nil, errors.New("no channel provided")
	}

	c := &ConditionalReboot{
		groups:              groups,
		rebootImpl:          rebootImpl,
		rebootRequest:       rebootReq,
		audit:               &journal.NoopJournal{},
		safeMinSystemUptime: defaultSafeMinimumSystemUptime,
	}

	var errs error
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return c, errs
}

// IsSafeSystemBootUptimeReached returns whether minimum limit of system uptime has been reached or not. This is used
// to prevent reboot loops.
func (app *ConditionalReboot) IsSafeSystemBootUptimeReached() bool {
	systemUptime, err := uptime.Uptime()
	if err != nil {
		log.Error().Err(err).Msgf("could not determine system uptime, rebooting anyway: %v", err)
		return true
	}

	if systemUptime >= app.safeMinSystemUptime {
		return true
	}
	return false
}

func (app *ConditionalReboot) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	for _, group := range app.groups {
		group.Start(ctx)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-sig:
			cancel()
			log.Info().Msgf("Received signal, cancelling..")
			return

		case group := <-app.rebootRequest:
			log.Info().Msgf("Reboot request from group '%s'", group.GetName())
			err := app.tryReboot(group)
			if err != nil {
				internal.RebootErrors.Set(1)
				log.Error().Err(err).Msg("Reboot failed")
			} else {
				log.Info().Msgf("Cancelling all checkers...")
				cancel()
				// TODO: Get rid of lazy way, use waitgroups?!
				time.Sleep(5 * time.Second)
			}
		}
	}
}

var printUptimeWarning = true // prevent repetitive logs
func (app *ConditionalReboot) tryReboot(group *group.Group) error {
	if !app.IsSafeSystemBootUptimeReached() {
		if printUptimeWarning {
			printUptimeWarning = false
			log.Warn().Msgf("Refusing to reboot, safe minimum system uptime (%s) not reached yet", defaultSafeMinimumSystemUptime)
		}
		return nil
	}

	log.Info().Msg("Trying to reboot...")
	if err := app.audit.Journal(actionToText(group)); err != nil {
		log.Err(err).Msg("could not write journal")
	}

	return app.rebootImpl.Reboot()
}

func actionToText(g *group.Group) string {
	now := time.Now()
	formattedTime := now.Format("2006-01-02T15:04:05-07:00")
	return fmt.Sprintf("%s Group '%s' requested reboot", formattedTime, g.GetName())
}
