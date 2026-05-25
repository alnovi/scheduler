# Scheduler

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/alnovi/scheduler)](https://go.dev/dl/)
[![GitHub License](https://img.shields.io/github/license/alnovi/sso)](https://github.com/alnovi/scheduler/blob/master/LICENSE.md)

**Scheduler** — планировщик задач. Он позволяет автоматически запускать команды, скрипты в заранее определённое время
или с заданной периодичностью.

## Установка

```sh
go get github.com/alnovi/scheduler
```

## Опции планировщика

|       Опция       | Описание                                                                            |
|:-----------------:|-------------------------------------------------------------------------------------|
|  **WithLogger**   | Использовать свой логгер (slog.Logger)                                              |
|  **WithLocker**   | Использовать свою реализацию для распределенной блокировки повторного запуска задач |
|  **WithMetrics**  | Заполнять метрики                                                                   |
| **WithLocation**  | Указать временную зону для планировщика                                             |
| **WithContextFn** | Функция задающая context для задачи                                                 |

## Добавление задачи

| Пример                                       | Описание                                        |
|----------------------------------------------|-------------------------------------------------|
| `cron.AddDurationTask(time.Minute, &Task{})` | Запуск задачи через одинаковый интервал времени |
| `cron.AddDayAtTask(10, 30, &Task{})`         | Запуск задачи раз в день в 10:30                |
| `cron.AddCronTask("* * * * *", &Task{})`     | Запуск задачи используя cron выражение          |

## Типы задач

Все задачи должны имплементировать контракт `Task`:

```golang
type Task interface {
Name() string
Handle(ctx context.Context) error
}
```

Для дополнительных возможностей требуется реализовать следующие контракты:

- **TaskContext** - позволяет изменить context перед выполнением задачи
- **TaskTimeout** - ограничение времени работы задачи
- **TaskLocker** - распределенная блокировка, не запускать задачу параллельно на нескольких серверах

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
	cron.AddCronTask(scheduler.EveryMinute, &Task{})
	cron.Start()
}
```