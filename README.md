# Scheduler

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/alnovi/scheduler)](https://go.dev/dl/)
[![GitHub License](https://img.shields.io/github/license/alnovi/sso)](https://github.com/alnovi/scheduler/blob/master/LICENSE.md)

**Scheduler** — планировщик задач. Он позволяет автоматически запускать команды, скрипты в заранее определённое время или с заданной периодичностью.

## Установка

```sh
go get -u github.com/alnovi/scheduler
```

## Использование

```go
package main

import (
	"context"
	"fmt"
	"time"
	"github.com/alnovi/scheduler"
)

type Task struct{}

func (t *Task) Name() string {
	return "task"
}

func (t *Task) Timeout() time.Duration {
	return time.Minute
}

func (t *Task) Handle(_ context.Context) error {
	fmt.Println(time.Now())
	return nil
}

func main() {
	cron := scheduler.New()
	cron.AddDurationTask(time.Second, &Task{})
	cron.Start()
}
```