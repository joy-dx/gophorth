package updaterdto

const (
	UPDATER_SETTING_LAST_TIME_CHECKED_UPDATE = "core_last_time_checked_update"
)

type UpdateStatus string

const (
	INITIAL          UpdateStatus = "initial"
	UPDATE_AVAILABLE UpdateStatus = "update_available"
	CHECKING         UpdateStatus = "checking"
	COMPLETE         UpdateStatus = "complete"
	DOWNLOADED       UpdateStatus = "downloaded"
	ERROR            UpdateStatus = "error"
	IN_PROGRESS      UpdateStatus = "in_progress"
	STOPPED          UpdateStatus = "stopped"
	UP_TO_DATE       UpdateStatus = "up_to_date"
)
