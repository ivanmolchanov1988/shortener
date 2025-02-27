package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ivanmolchanov1988/shortener/internal/memory"
)

func BenchmarkPostURL(b *testing.B) {
	// Инициализация памяти и обработчика
	memStore := memory.NewMemoryStorage()
	handler := NewHandler(memStore, cfg)

	// Сбрасываем таймер, чтобы подготовка не учитывалась
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Создаём новый запрос и Recorder для каждого прохода
		req := httptest.NewRequest(http.MethodPost, "http://localhost:8080", strings.NewReader("https://example.com"))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		// Вызываем обработчик
		handler.PostURL(w, req)

		// Закрываем тело запроса для освобождения ресурсов
		_ = req.Body.Close()
	}
}
