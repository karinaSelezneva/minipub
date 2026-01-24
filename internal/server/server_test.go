package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/karinaSelezneva/minipub/internal/broker"
)

func TestPublishHandler(t *testing.T) {
	// 1. Создаем зависимости
	b := broker.NewBroker()
	srv := &Server{
		Broker: b,
	}

	// 2. Готовим JSON-тело запроса
	jsonBody := []byte(`{"topic": "news", "message": "Hello, world!"}`)

	// 3. Создаем фейковый запрос и рекордер (куда запишется ответ)
	req := httptest.NewRequest(http.MethodPost, "/publish", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	// 4. Вызываем хендлер напрямую
	srv.PublishHandler(rr, req)

	// 5. Проверяем результат
	if rr.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d\n", rr.Code)
	}

	expected := "Success: topic news" // Проверь, совпадает ли с твоим выводом
	if rr.Body.String() != expected {
		t.Errorf("ожидали ответ %s, получили %s", expected, rr.Body.String())
	}
}

func TestIntegration_SubAndPub(t *testing.T) {
	b := broker.NewBroker()
	s := &Server{Broker: b}

	topic := "integration_test"
	msg := "test_message"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Создаем PIPE: pr - для чтения, pw - для записи
	pr, pw := io.Pipe()

	// 2. Создаем кастомный Recorder, который будет писать в нашу "трубу"
	subRR := httptest.NewRecorder()
	// Это магия: подменяем стандартный Body на наш конец трубы для записи
	subRR.Body = nil

	subReq := httptest.NewRequest(http.MethodGet, "/subscribe?topic="+topic, nil).WithContext(ctx)

	// 3. Запускаем хендлер, передав ему специальный ResponseWriter
	go func() {
		// Обертка, которая перенаправляет Write из хендлера в pw
		s.SubscribeHandler(fakeWriter{subRR, pw}, subReq)
	}()

	time.Sleep(50 * time.Millisecond)

	// 4. Публикуем сообщение
	pubBody := []byte(fmt.Sprintf(`{"topic": "%s", "message": "%s"}`, topic, msg))
	pubReq := httptest.NewRequest(http.MethodPost, "/publish", bytes.NewBuffer(pubBody))
	s.PublishHandler(httptest.NewRecorder(), pubReq)

	// 5. Читаем из трубы с тайм-аутом
	// 5. Читаем из трубы через Scanner (он будет ждать новых строк)
	resCh := make(chan string)
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := scanner.Text()
			// Нас интересует только строка с сообщением
			if strings.Contains(line, "Message:") {
				resCh <- line
				return // Нашли, что искали — выходим
			}
		}
	}()

	select {
	case result := <-resCh:
		if !strings.Contains(result, msg) {
			t.Errorf("Ожидали %s, получили %s", msg, result)
		}
	case <-time.After(1 * time.Second):
		t.Error("Тайм-аут: сообщение не дошло")
	}
}

// Вспомогательная структура для теста
type fakeWriter struct {
	*httptest.ResponseRecorder
	io.Writer
}

func (f fakeWriter) Write(b []byte) (int, error) { return f.Writer.Write(b) }
func (f fakeWriter) Flush()                      {} // Имитируем Flusher
