package journal

type Journal interface {
	Journal(action string) error
}
