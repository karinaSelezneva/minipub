package broker

import "sync"

// Определи структуру Broker:
// 1. Внутри должно быть поле для мапы (ключ — строка, значение — слайс каналов).
// 2. Поле для мьютекса (sync.RWMutex)

type Broker struct {
	mu          sync.RWMutex
	subsByTopic map[string][]chan string
}

// Напиши конструктор NewBroker():
// Функция, которая возвращает инициализированную структуру (не забудь сделать make для мапы!).
func NewBroker() *Broker {
	return &Broker{
		subsByTopic: make(map[string][]chan string),
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

			return
		}
	}
}
