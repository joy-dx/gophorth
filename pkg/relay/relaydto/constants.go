package relaydto

type RelayLevel string

const (
	Debug RelayLevel = "debug"
	Info  RelayLevel = "info"
	Warn  RelayLevel = "warn"
	Error RelayLevel = "error"
	Fatal RelayLevel = "fatal"
)
