package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/conditional-reboot/internal/group"
	"github.com/soerenschneider/conditional-reboot/internal/journal"
	"github.com/soerenschneider/conditional-reboot/pkg/reboot"
	"os"
	"os/signal"
	"time"
)

type ConditionalReboot struct {
	groups        []*group.Group
	rebootImpl    reboot.Reboot
	audit         journal.Journal
	rebootRequest chan *group.Group
}

func NewConditionalReboot(groups []*group.Group, rebootImpl reboot.Reboot, audit journal.Journal, rebootReq chan *group.Group) (*ConditionalReboot, error) {
	if len(groups) == 0 {
		return nil, errors.New("no groups provided")
	}

	if rebootImpl == nil {
		return nil, errors.New("no reboot impl provided")
	}

	if audit == nil {
		return nil, errors.New("empty journal provided")
	}

	if rebootReq == nil {
		return nil, errors.New("no channel provided")
	}

	return &ConditionalReboot{
		groups:        groups,
		rebootImpl:    rebootImpl,
		audit:         audit,
		rebootRequest: rebootReq,
	}, nil
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
			log.Info().Msgf("Group '%s' requests reboot", group.GetName())

			if err := app.audit.Journal(actionToText(group)); err != nil {
				log.Err(err).Msg("could not write journal")
			}

			log.Info().Msg("Trying to reboot...")
			if err := app.rebootImpl.Reboot(); err != nil {
				RebootErrors.Set(1)
				log.Error().Err(err).Msg("reboot failed")
			} else {
				log.Info().Msgf("Cancelling")
				cancel()
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func actionToText(g *group.Group) string {
	now := time.Now()
	formattedTime := now.Format("2006-01-02T15:04:05-07:00")
	return fmt.Sprintf("%s Group '%s' requested reboot", formattedTime, g.GetName())
}
