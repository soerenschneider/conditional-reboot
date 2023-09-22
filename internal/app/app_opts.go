package app

import (
	"errors"
	"time"

	"github.com/soerenschneider/conditional-reboot/internal/journal"
)

func UseJournal(journal journal.Journal) ConditionalRebootOpts {
	return func(c *ConditionalReboot) error {
		if journal == nil {
			return errors.New("nil journal provided")
		}

		c.audit = journal
		return nil
	}
}

func SafeMinSystemUptime(duration time.Duration) ConditionalRebootOpts {
	return func(c *ConditionalReboot) error {
		if duration.Hours() <= 1 {
			return errors.New("duration should not be less than 1 hour")
		}

		c.safeMinSystemUptime = duration
		return nil
	}
}
