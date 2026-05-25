package scheduler

const (
	EveryMinute             = "* * * * *"    // Каждую минуту
	Every5Minutes           = "*/5 * * * *"  // Каждые 5 минут
	Every15Minutes          = "*/15 * * * *" // Каждые 15 минут
	Every30Minutes          = "*/30 * * * *" // Каждые 30 минут
	Hourly                  = "0 * * * *"    // Каждый час в начале часа
	HalfHourly              = "30 * * * *"   // Каждый час в 30 минут
	Every2Hours             = "0 */2 * * *"  // Каждые 2 часа
	Every4Hours             = "0 */4 * * *"  // Каждые 4 часа
	Every6Hours             = "0 */6 * * *"  // Каждые 6 часов
	Every12Hours            = "0 */12 * * *" // Каждые 12 часов
	DailyAtMidnight         = "0 0 * * *"    // Каждый день в полночь
	DailyWorkingAtMidnight  = "0 0 W * *"    // Каждый рабочий день в полночь
	WeeklySundayMidnight    = "0 0 * * 0"    // Каждое воскресенье в полночь
	WeeklySaturdayMidnight  = "0 0 * * 6"    // Каждую субботу в полночь
	MonthlyFirstDayMidnight = "0 0 1 * *"    // Каждый первый день месяца в полночь
	MonthlyLstDayMidnight   = "0 0 L * *"    // Каждый последний день месяца в полночь
	YearlyNewYearMidnight   = "0 0 1 1 *"    // Раз в год, 1 января в 00:00
)
