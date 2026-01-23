package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/karinaSelezneva/minipub/internal/broker"
)

// 1. Создаем структуру для парсинга JSON
type PublishRequest struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

// Создай структуру Server, которая хранит ссылку на твой Broker.
type Server struct {
	Broker *broker.Broker
}

// 1. Напиши метод PublishHandler:
func (s *Server) PublishHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод POST (хороший тон для API)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PublishRequest

	// 2. Декодируем JSON напрямую из тела запроса
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация (чтобы не слали пустые поля)
	if req.Topic == "" || req.Message == "" {
		http.Error(w, "topic and message are required", http.StatusUnprocessableEntity)
		return
	}

	// 3. Вызываем метод брокера
	s.Broker.Publish(req.Topic, req.Message)

	// 4. Возвращаем статус 200 OK
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Success: topic %s", req.Topic)
}

// Напиши метод SubscribeHandler:
func (s *Server) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// 1. Получаем топик из параметров запроса
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "topic is required", http.StatusBadRequest)
		return
	}

	// 2. Подписываемся и получаем канал
	ch := s.Broker.Subscribe(topic)
	// Важно: когда клиент отключится, нужно отписаться!
	defer s.Broker.Unsubscribe(topic, ch)

	// 3. Устанавливаем заголовки, чтобы curl понимал, что это поток
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprintf(w, "---Subscribed to topic %s---\n\n", topic)

	// Принудительно отправляем текст клиенту (flushing)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// 4. БЕСКОНЕЧНЫЙ ЦИКЛ: ждем сообщения из канала и пишем в ответ
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "Message: %s\n", msg)

			// Снова пушим данные в сеть
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done(): // Клиент отключился
			fmt.Fprintf(w, "---Unsubscribed from topic %s---\n", topic)
			// Если клиент закрыл curl (Ctrl+C), выходим из цикла
			return
		}
	}
}
