package journal

type NoopAudit struct{}

func (n *NoopAudit) Journal(string) error {
	return nil
}
