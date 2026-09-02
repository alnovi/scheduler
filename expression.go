package scheduler

const (
	CronEveryMinute             = "* * * * *"    // Каждую минуту
	CronEvery5Minutes           = "*/5 * * * *"  // Каждые 5 минут
	CronEvery15Minutes          = "*/15 * * * *" // Каждые 15 минут
	CronEvery30Minutes          = "*/30 * * * *" // Каждые 30 минут
	CronHourly                  = "0 * * * *"    // Каждый час в начале часа
	CronHalfHourly              = "30 * * * *"   // Каждый час в 30 минут
	CronEvery2Hours             = "0 */2 * * *"  // Каждые 2 часа
	CronEvery4Hours             = "0 */4 * * *"  // Каждые 4 часа
	CronEvery6Hours             = "0 */6 * * *"  // Каждые 6 часов
	CronEvery12Hours            = "0 */12 * * *" // Каждые 12 часов
	CronDailyAtMidnight         = "0 0 * * *"    // Каждый день в полночь
	CronDailyWorkingAtMidnight  = "0 0 W * *"    // Каждый рабочий день в полночь
	CronWeeklySundayMidnight    = "0 0 * * 0"    // Каждое воскресенье в полночь
	CronWeeklySaturdayMidnight  = "0 0 * * 6"    // Каждую субботу в полночь
	CronMonthlyFirstDayMidnight = "0 0 1 * *"    // Каждый первый день месяца в полночь
	CronMonthlyLstDayMidnight   = "0 0 L * *"    // Каждый последний день месяца в полночь
	CronYearlyNewYearMidnight   = "0 0 1 1 *"    // Раз в год, 1 января в 00:00
)
