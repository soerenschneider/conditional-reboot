package deps

import (
	"github.com/soerenschneider/conditional-reboot/internal"
	"github.com/soerenschneider/conditional-reboot/internal/journal"
)

func BuildAudit(config *internal.ConditionalRebootConfig) (journal.Journal, error) {
	if len(config.JournalFile) > 0 {
		return journal.NewJournalFile(config.JournalFile)
	}

	return &journal.NoopAudit{}, nil
}
