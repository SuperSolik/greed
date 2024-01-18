package greed

const DATETIME_DB_LAYOUT = "2006-01-02T15:04:05-07:00"
const DATE_INPUT_LAYOUT = "2006-01-02"
const DATE_NICE_LAYOUT = "02-01-2006"
const TIME_INPUT_LAYOUT = "15:04"
const DATETIME_INPUT_LAYOUT = "2006-01-02 15:04"

type DateRangeType string
type DateRangePickerOption = Pair[DateRangeType, string]

const (
	NotSelected DateRangeType = "placeholder"
	None        DateRangeType = "none"
	Today       DateRangeType = "today"
	ThisWeek    DateRangeType = "this_week"
	ThisMonth   DateRangeType = "this_month"
	ThisYear    DateRangeType = "this_year"
	Last7Days   DateRangeType = "last_7_days"
	Last30Days  DateRangeType = "last_30_days"
	Custom      DateRangeType = "custom"
)

var DateRangePickerOptions = []DateRangePickerOption{
	{First: NotSelected, Second: "date filter..."},
	{First: None, Second: "-"},
	{First: Today, Second: "today"},
	{First: ThisWeek, Second: "this week"},
	{First: ThisMonth, Second: "this month"},
	{First: ThisYear, Second: "this year"},
	{First: Last7Days, Second: "last 7 days"},
	{First: Last30Days, Second: "last 30 days"},
	{First: Custom, Second: "custom"},
}
