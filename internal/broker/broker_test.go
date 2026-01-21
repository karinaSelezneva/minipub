package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	tests := []struct {
		name          string // Имя теста
		topic         string // Аргумент: топик
		expectedCount int    // Что ожидаем в итоге
	}{
		{
			name:          "Добавление подписки",
			topic:         "news",
			expectedCount: 1,
		},
		{
			name:          "Добавление второй подписки",
			topic:         "news",
			expectedCount: 2,
		},
		{
			name:          "Добавление подписки на другой топик",
			topic:         "beauty",
			expectedCount: 1,
		},
	}

	b := NewBroker()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := b.Subscribe(tt.topic)

			assert.NotNil(t, ch, "Метод должен возвращать канал")
			assert.Len(t, b.subsByTopic[tt.topic], tt.expectedCount) // Проверяем, что количество подписчиков увеличилось
		})
	}
}

func TestUnsubscribe(t *testing.T) {
	tests := []struct {
		name           string // Имя теста
		topic          string // Аргумент: топик
		numSubscribers int    // Сколько раз вызвать Subscribe сначала
		unsubIndex     int    // Индекс канала из созданных, который удаляем
		expectedCount  int    // Что ожидаем в итоге
	}{
		{
			name:           "Удаление единственного подписчика",
			topic:          "news",
			numSubscribers: 1,
			unsubIndex:     0,
			expectedCount:  0,
		},
		{
			name:           "Удаление одного из двух",
			topic:          "sport",
			numSubscribers: 2,
			unsubIndex:     0,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBroker()

			// 1. Создаем подписки и запоминаем каналы
			var channels []chan string
			for i := 0; i < tt.numSubscribers; i++ {
				ch := b.Subscribe(tt.topic)
				channels = append(channels, ch)
			}

			// 2. Вызываем удаление конкретного канала
			if len(channels) > 0 {
				b.Unsubscribe(tt.topic, channels[tt.unsubIndex])
			}

			// 3. Проверяем результат
			assert.Equal(t, tt.expectedCount, len(b.subsByTopic[tt.topic]))
		})
	}
}
