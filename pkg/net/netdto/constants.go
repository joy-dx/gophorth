package netdto

const NET_DEFAULT_CLIENT_REF = "default"

type TransferStatus string

const (
	IN_PROGRESS TransferStatus = "in_progress"
	ERROR       TransferStatus = "error"
	COMPLETE    TransferStatus = "complete"
	STOPPED     TransferStatus = "stopped"
)
