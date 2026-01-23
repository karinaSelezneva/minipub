package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	server "github.com/karinaSelezneva/minipub/internal/api"
	"github.com/karinaSelezneva/minipub/internal/broker"
)

func main() {
	b := broker.NewBroker()
	// Инициализируй Server с этим брокером.
	srv := &server.Server{
		Broker: b,
	}
	// Регистрируем хендлеры
	http.HandleFunc("/publish", srv.PublishHandler)

	http.HandleFunc("/subscribe", srv.SubscribeHandler)
	// Запусти http.ListenAndServe(":8080", nil).
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
	log.Println("🚀 Сервер запущен на :8080")

	var wg sync.WaitGroup

	topic := "sport"
	ch := b.Subscribe(topic)

	wg.Add(1)

	go func() {
		defer wg.Done()

		// Читаем только 1 сообщение, чтобы тест завершился
		// Если нужен цикл for range, то брокер должен уметь закрывать каналы
		msg, ok := <-ch
		if ok {
			fmt.Printf("✅ Получено в main: %s\n", msg)
		}
	}()

	// Публикуем сообщение
	b.Publish(topic, "Привет из main!")

	// Ждем завершения горутины
	wg.Wait()

	fmt.Println("🚀 Все сообщения обработаны, выходим.")
}

// Как тестировать:
// Открой терминал и сделай: curl "http://localhost:8080/subscribe?topic=NikePro" (он зависнет в ожидании — это нормально).
// В другом терминале: curl -X POST -d '{"topic": "go", "message": "Rocks!"}' http://localhost:8080/publish.
// В первом терминале должна появиться строка "Rocks!"

// curl -X POST -d '{"topic": "NikePro", "message": "Just Do It!"}' http://localhost:8080/publish
