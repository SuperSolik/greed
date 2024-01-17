package greed

const DATETIME_DB_LAYOUT = "2006-01-02T15:04:05-07:00"
const DATE_INPUT_LAYOUT = "2006-01-02"
const DATE_NICE_LAYOUT = "02-01-2006"
const TIME_INPUT_LAYOUT = "15:04"
const DATETIME_INPUT_LAYOUT = "2006-01-02 15:04"

type DateRangePickerOption = Pair[string, string]

// enum for date range filter
var (
	NotSelected = DateRangePickerOption{First: "placeholder", Second: "date filter..."}
	None        = DateRangePickerOption{First: "none", Second: "-"}
	Today       = DateRangePickerOption{First: "today", Second: "today"}
	ThisWeek    = DateRangePickerOption{First: "this_week", Second: "this week"}
	ThisMonth   = DateRangePickerOption{First: "this_month", Second: "this month"}
	ThisYear    = DateRangePickerOption{First: "this_year", Second: "this year"}
	Last7Days   = DateRangePickerOption{First: "last_7_days", Second: "last 7 days"}
	Last30Days  = DateRangePickerOption{First: "last_30_days", Second: "last 30 days"}
	Custom      = DateRangePickerOption{First: "custom", Second: "custom"}
)

var DateRangePickerOptions = []DateRangePickerOption{
	NotSelected,
	None,
	Today,
	ThisWeek,
	ThisMonth,
	ThisYear,
	Last7Days,
	Last30Days,
	Custom,
}
