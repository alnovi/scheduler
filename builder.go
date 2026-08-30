package scheduler

import "time"

type Builder = func(now time.Time) (bool, *taskWrap, error)

func Cron(expression string, t Task, opts ...TaskOption) Builder {
	return CronIf(true, expression, t, opts...)
}

func Duration(d time.Duration, t Task, opts ...TaskOption) Builder {
	return DurationIf(true, d, t, opts...)
}

func DayAt(hour, minute int, t Task, opts ...TaskOption) Builder {
	return DayAtIf(true, hour, minute, t, opts...)
}

func CronIf(e bool, expression string, t Task, opts ...TaskOption) Builder {
	return func(now time.Time) (bool, *taskWrap, error) {
		if !e || t == nil {
			return false, nil, nil
		}

		task, err := newTaskWrapCron(expression, t, now, opts)

		return true, task, err
	}
}

func DurationIf(e bool, d time.Duration, t Task, opts ...TaskOption) Builder {
	return func(now time.Time) (bool, *taskWrap, error) {
		if !e || t == nil {
			return false, nil, nil
		}

		task, err := newTaskWrapDuration(d, t, now, opts)
		if err != nil {
			return true, nil, err
		}

		task.nextRun = task.nextRun.Truncate(time.Minute).Add(task.duration)

		return true, task, nil
	}
}

func DayAtIf(e bool, hour, minute int, t Task, opts ...TaskOption) Builder {
	return func(now time.Time) (bool, *taskWrap, error) {
		if !e || t == nil {
			return false, nil, nil
		}

		d, _ := time.ParseDuration("24h")

		task, err := newTaskWrapDuration(d, t, now, opts)
		if err != nil {
			return true, nil, err
		}

		task.nextRun = time.Date(task.nextRun.Year(), task.nextRun.Month(), task.nextRun.Day(), hour, minute, 0, 0, now.Location())

		return true, task, nil
	}
}
