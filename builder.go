package scheduler

import "time"

type Builder = func(now time.Time) (bool, *taskWrap, error)

func Cron(expression string, t Task) Builder {
	return CronIf(true, expression, t)
}

func Duration(d time.Duration, t Task) Builder {
	return DurationIf(true, d, t)
}

func DayAt(hour, minute int, t Task) Builder {
	return DayAtIf(true, hour, minute, t)
}

func CronIf(e bool, expression string, t Task) Builder {
	return func(now time.Time) (bool, *taskWrap, error) {
		if !e || t == nil {
			return false, nil, nil
		}

		task, err := newTaskWrapCron(expression, t, now)

		return true, task, err
	}
}

func DurationIf(e bool, d time.Duration, t Task) Builder {
	return func(now time.Time) (bool, *taskWrap, error) {
		if !e || t == nil {
			return false, nil, nil
		}

		task, err := newTaskWrapDuration(d, t)
		if err != nil {
			return true, nil, err
		}

		task.nextRun = now.Truncate(time.Minute).Add(task.duration)

		return true, task, nil
	}
}

func DayAtIf(e bool, hour, minute int, t Task) Builder {
	return func(now time.Time) (bool, *taskWrap, error) {
		if !e || t == nil {
			return false, nil, nil
		}

		d, _ := time.ParseDuration("24h")

		task, err := newTaskWrapDuration(d, t)
		if err != nil {
			return true, nil, err
		}

		task.nextRun = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		return true, task, nil
	}
}
