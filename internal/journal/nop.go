package journal

type NoopJournal struct{}

func (n *NoopJournal) Journal(string) error {
	return nil
}
