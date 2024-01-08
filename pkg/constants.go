package greed

const DATETIME_DB_LAYOUT = "2006-01-02T15:04:05-07:00"
const DATE_INPUT_LAYOUT = "2006-01-02"
const TIME_INPUT_LAYOUT = "15:04"
const DATETIME_INPUT_LAYOUT = "2006-01-02 15:04"

// enum for date range filter
const (
	None       string = "none"
	Today      string = "today"
	Last7Days  string = "last_7_days"
	LastWeek   string = "last_week"
	Last30Days string = "last_30_days"
	LastMonth  string = "last_month"
	LastYear   string = "last_year"
	Custom     string = "custom"
)
