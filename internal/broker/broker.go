package broker

import (
	"log/slog"
	"os"
	"sync"
)

// Определи структуру Broker:
// 1. Внутри должно быть поле для мапы (ключ — строка, значение — слайс каналов).
// 2. Поле для мьютекса (sync.RWMutex)

type Broker struct {
	mu          sync.RWMutex
	subsByTopic map[string][]chan string
	logger      *slog.Logger
}

// Напиши конструктор NewBroker():
// Функция, которая возвращает инициализированную структуру (не забудь сделать make для мапы!).
func NewBroker() *Broker {
	// Создаем логгер, который пишет только важную инфу в одну строку
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Опционально: убираем время для локальной разработки, чтобы не мусорило
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return &Broker{
		subsByTopic: make(map[string][]chan string),
		logger:      slog.New(handler),
	}
}

// Реализуй метод Subscribe(topic string) chan string:
// 1. Он должен создать новый канал.
// 2. Заблокировать мьютекс на запись.
// 3. Добавить канал в слайс по ключу topic.
// 4. Вернуть этот канал вызывающему коду.

func (b *Broker) Subscribe(topic string) chan string {
	ch := make(chan string, 1)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subsByTopic[topic] = append(b.subsByTopic[topic], ch)

	b.logger.Info("+New Subscribe", "topic", topic, "count", len(b.subsByTopic[topic]))

	return ch
}

// Реализуй метод Unsubscribe(topic string, ch chan string):
// (Это со звездочкой): Нужно найти канал в слайсе и удалить его.
// Пригодится знание того, как удалять элемент из слайса (через append или сдвиг). Не забудь про мьютекс!
func (b *Broker) Unsubscribe(topic string, ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.subsByTopic[topic]
	if !ok {
		return // Топика нет, делать нечего
	}

	for i, val := range subs {
		if val == ch {
			b.subsByTopic[topic] = append(subs[:i], subs[i+1:]...)

			// Если подписчиков больше нет — удаляем топик целиком
			if len(b.subsByTopic[topic]) == 0 {
				delete(b.subsByTopic, topic)
			}

			b.logger.Info("-Unsubscribe", "topic", topic, "count", len(b.subsByTopic[topic]))

			return
		}
	}
}
