# Scheduler

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/alnovi/scheduler)](https://go.dev/dl/)
[![GitHub License](https://img.shields.io/github/license/alnovi/scheduler)](https://github.com/alnovi/scheduler/blob/master/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/alnovi/scheduler)](https://github.com/alnovi/scheduler/releases)
![coverage](https://raw.githubusercontent.com/alnovi/scheduler/badges/.badges/master/coverage.svg)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/alnovi/scheduler/master.yml)

**Scheduler** — планировщик задач. Позволяет автоматически запускать команды, скрипты с заданной периодичностью.
Так же, доступна возможность реализации собственной логики запуска задач.

## Установка

```sh
go get github.com/alnovi/scheduler/v2
```

## Опции планировщика

| Option           | Default               | Description                                                   |
|------------------|-----------------------|---------------------------------------------------------------|
| **WithLogger**   | `slog.DiscardHandler` | Использовать преднастроенный логгер (slog.Logger)             |
| **WithLocation** | `time.UTC`            | Указать временную зону для планировщика                       |
| **WithLocker**   | `nil`                 | Использовать locker для блокировки паралельного запуска задач |
| **WithMetrics**  | `nil`                 | Собирать метрики по работе планировщика                       |

## Опции задач

| Option          | Default | Description                                                   |
|-----------------|---------|---------------------------------------------------------------|
| **WithEnabled** | `true`  | Активация/деактивация задачи                                  |
| **WithTimeout** | `0`     | Время для выполнения задачи (0 - не ограничено)               |
| **WithDelay**   | `0`     | Отложеный запуск задачи (только первый запуск)                |
| **WithLock**    | `0`     | Блокировать задачу для паралельного запуска на указаное время |

## Типы задач

| Type task        | Example                                 | Description                              |
|------------------|-----------------------------------------|------------------------------------------|
| **CronTask**     | `MustCronTask("* * * * *", handle)`     | Запуск испульзуя cron выражение          |
| **DurationTask** | `MustDurationTask(time.Minute, handle)` | Запуск через одинаковый интервал времени |

## Пример использования

```go
package main

import (
	"context"
	"fmt"
	"time"
	"github.com/alnovi/scheduler/v2"
	"github.com/alnovi/scheduler/v2/tasks"
)

func main() {
	sch := scheduler.New()

	handleFn := func(_ context.Context) error {
		fmt.Println("exec ok")
		return nil
	}

	task, _ := tasks.NewDurationTask(
		time.Minute,
		handleFn,
		tasks.WithTimeout(time.Second),
	)

	sch.AddTask("test", task)
	sch.Start(context.Background())
}
```

## Реализация собственного типа задачи

Для реализации собственного типа задачи, требуется имплементировать интерфейс `scheduler.Task`. Базовая функциональность
добавляется с помощью структуры `tasks.Base`.

### Пример собственного типа

```go
package example

import (
	"time"
	"github.com/alnovi/scheduler/v2/tasks"
)

type HourTask struct {
	*tasks.Base
	next time.Time
}

func NewHourTask(handler tasks.HandlerFn, opts ...tasks.Option) (*HourTask, error) {
	base, err := tasks.NewBase(handler, opts...)
	if err != nil {
		return nil, err
	}
	return &HourTask{Base: base}, nil
}

func (t *HourTask) Init(now time.Time) error {
	t.next = now.Add(t.Delay()).Round(time.Hour)
	return nil
}

func (t *HourTask) Compare(now time.Time) (bool, error) {
	for now.After(t.next) {
		t.next = now.Add(time.Hour).Round(time.Hour)
	}
	return t.next.Minute() == 0, nil
}
```

# Блокировка параллельного запуска (locker)

Библиотека содержит простой locker для работы в единственном экземпляре.
Для распределенной блокировки требуется собственная реализация блокировщика, на пример,
с использованием redis или других систем.  
Для реализации locker, достаточно имплементировать интерфейс `scheduler.Locker`.

```golang
type Locker interface {
	LockResource(ctx context.Context, resource string, ttl time.Duration) (bool, string, error)
}

```
