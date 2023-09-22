package deps

import (
	"github.com/soerenschneider/conditional-reboot/internal/config"
	"github.com/soerenschneider/conditional-reboot/internal/journal"
)

func BuildAudit(config *config.ConditionalRebootConfig) (journal.Journal, error) {
	if len(config.JournalFile) > 0 {
		return journal.NewFileJournal(config.JournalFile)
	}

	return &journal.NoopJournal{}, nil
}
