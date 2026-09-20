# ktan — Kafka Topic ANalyzer

TUI-инструмент для просмотра Kafka: список топиков и чтение содержимого.
Терминальное приложение на Go, локальный dev-стек Kafka в docker-compose.

## Стек

- Go 1.26 (модуль `github.com/satmaelstorm/ktan`)
- TUI: `charmbracelet/bubbletea` + `bubbles` (spinner, viewport) + `lipgloss`
- Kafka: `twmb/franz-go` (kgo + kmsg, низкоуровневые запросы через `RequestCachedMetadata`)
- Kafka-сервер: `apache/kafka:4.1.2` в KRaft-режиме (одна нода, 3 партиции по умолчанию)

## Команды

```bash
# поднять локальный Kafka (advertised listener: localhost:9092)
docker compose up -d

# собрать и запустить (дефолт -brokers localhost:9092)
go run . 
go run . -brokers localhost:9092,other:9092
go run . -b localhost:9092 -l en

# проверки
go build ./...
go vet ./...
gofmt -l .
```

Тестов пока нет; при добавлении — стандартный `go test ./...`.

## Структура

```
main.go                      # тонкая обёртка: cmd.Execute()
cmd/
  root.go                    # корневая cobra-команда: флаги -brokers/-lang, запуск TUI
internal/kafka/              # слой работы с Kafka, без зависимостей от UI
  client.go                  #   обёртка kgo.Client + mutex на переключение партиций
  topics.go                  #   ListTopics: metadata-запрос, фильтр internal/__*-топиков
  messages.go                #   TailMessages: чтение хвоста всех партиций, сортировка по времени
internal/ui/                 # bubbletea-модели
  app.go                     #   корневая модель, роутинг экранов (topics <-> messages)
  i18n.go                    #   локаль интерфейса: localeGothic (дефолт) / localeEn (-lang=en)
  topics_view.go             #   экран списка топиков
  messages_view.go           #   экран сообщений (tail 10, viewport, refresh по 'r')
  styles.go                  #   lipgloss-стили + truncate()
docker-compose.yml           # локальный Kafka (KRaft, 1 брокер, healthcheck)
```

## Архитектурные решения и соглашения

- **Слои:** `internal/kafka` не знает про UI; UI вызывает методы клиента из tea.Cmd
  (горутины bubbletea), результаты приходят в модель через msg-типы
  (`topicsLoadedMsg`, `messagesLoadedMsg`, `topicSelectedMsg`, `backMsg`).
- **Direct-консьюмер без группы:** клиент создаётся без consumer group;
  чтение хвоста — через `AddConsumePartitions` с offset `AtEnd().Relative(-limit)`.
  Переключение назначенных партиций защищено `Client.mu` (команды bubbletea
  выполняются в разных горутинах и могут пересекаться).
- **TailMessages:** читает до `limit * len(partitions)` записей с таймаутом 3s,
  сортирует по (timestamp, partition, offset), обрезает до `limit` самых свежих.
- **Контекст:** таймауты задаются внутри tea.Cmd (`context.WithTimeout` 5s);
  `context.WithValue` не используем — нужные данные передаются явно аргументами.
- **Стиль UI:** статусы/подсказки — `statusStyle`, ошибки — `errStyle`, заголовки —
  `titleStyle`, курсор — `selectedStyle`. Управление: vim-клавиши (j/k) + стрелки,
  `r` — refresh, `esc` — назад, `ctrl+c` — выход.
- **Интернационализация:** все строки UI — через структуру `locale` (internal/ui/i18n.go).
  Дефолт — `localeGothic` (стиль техножрецов: "machine-spirit", "data-canticles" и т.п.),
  `-lang=en` или `KTAN_LANG=en` переключают на `localeEn`. Приоритет: флаг `-lang`
  перебивает `KTAN_LANG`. Локаль создаётся в `ui.New` и передаётся
  в модели явно (не через глобальные переменные). Новые строки UI добавлять в `locale`,
  а не хардкодить в моделях.

## Что уже сделано

- Список топиков (имя + число партиций), навигация, refresh.
- Просмотр топика: 10 самых свежих сообщений (timestamp, partition, offset, key, value),
  скролл viewport, refresh, возврат по esc.
- Интернационализация: дефолт — готический стиль техножрецов, `-lang=en` / `KTAN_LANG=en` — английский (флаг приоритетнее env).
- Локальный Kafka в docker-compose для разработки.

## Идеи дальнейшей разработки

- Пагинация/скролл по истории сообщений (не только tail 10), переход к произвольному offset.
- Детальный просмотр сообщения (полный key/value, headers) по enter.
- Фильтр/поиск по сообщениям, JSON pretty-print для value.
- Инфо о топике: retention, in-sync replicas, offsets per partition, lag.
- Consumer groups: список, lag, сброс offset.
- Продюсер: отправка тестового сообщения в топик.
- Флаги/конфиг: TLS/SASL, размер tail, формат вывода.
- Тесты: unit на сортировку/обрезку TailMessages (с моком kgo сложнее — можно вынести чистые функции).
