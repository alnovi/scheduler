package scheduler

const (
	EveryMinute             = "* * * * *"
	Every5Minutes           = "*/5 * * * *"
	Every15Minutes          = "*/15 * * * *"
	Every30Minutes          = "*/30 * * * *"
	Hourly                  = "0 * * * *"
	HalfHourly              = "30 * * * *"
	Every2Hours             = "0 */2 * * *"
	Every4Hours             = "0 */4 * * *"
	Every6Hours             = "0 */6 * * *"
	Every12Hours            = "0 */12 * * *"
	DailyAtMidnight         = "0 0 * * *"
	DailyWorkingAtMidnight  = "0 0 W * *"
	WeeklySundayMidnight    = "0 0 * * 0"
	WeeklySaturdayMidnight  = "0 0 * * 6"
	MonthlyFirstDayMidnight = "0 0 1 * *"
	MonthlyLstDayMidnight   = "0 0 L * *"
	YearlyNewYearMidnight   = "0 0 1 1 *"
)
