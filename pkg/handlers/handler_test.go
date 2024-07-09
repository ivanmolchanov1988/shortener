package handlers

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ivanmolchanov1988/shortener/config"
	"github.com/ivanmolchanov1988/shortener/pkg/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// для теста конфига
var cfg = &config.Config{
	Address: "localhost:8080",
	B_URL:   "http://localhost:8080",
}

func init() {
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
}

func TestPostUrl(t *testing.T) {

	tests := []struct { // мне надо передать: контент, тело. Жду: код, ответ, контент
		name        string
		contentType string
		body        string
		want        struct {
			code        int
			response    string
			contentType string
		}
	}{
		{
			name:        "valid URL",
			contentType: "text/plain",
			body:        "https://practicum.yandex.ru/learn/go-advanced/courses/4059e8ec-b819-4c6c-801e-5307db3ff750/sprints/256128/topics/dbfad219-91f2-4f71-948d-953d4c449ad1/lessons/1e1e02c5-f0b0-4f61-97d2-3a7c8d8e9239/",
			want: struct {
				code        int
				response    string
				contentType string
			}{
				code:        http.StatusCreated,
				response:    "http://localhost:8080/",
				contentType: "text/plain",
			},
		},
		{
			name:        "invalid URL",
			contentType: "text/plain",
			body:        "invalid_url",
			want: struct {
				code        int
				response    string
				contentType string
			}{
				code:        http.StatusBadRequest,
				response:    "Invalid URL\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "invalid Content-Type",
			contentType: "application/json",
			body:        "https://practicum.yandex.ru/learn/go-advanced/courses/4059e8ec-b819-4c6c-801e-5307db3ff750/sprints/256128/topics/dbfad219-91f2-4f71-948d-953d4c449ad1/lessons/1e1e02c5-f0b0-4f61-97d2-3a7c8d8e9239/",
			want: struct {
				code        int
				response    string
				contentType string
			}{
				code:        http.StatusBadRequest,
				response:    "Content-Type must be text/plain\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	memStore := memory.NewMemoryStorage()
	handler := NewHandler(memStore, cfg)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.PostUrl(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.True(t, strings.HasPrefix(string(respBody), tt.want.response))
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

		})
	}

}

func TestGetUrl(t *testing.T) {

	//запись для тестов
	testShortURL := "testURL"
	invalidShortUrl := "123321"

	memStore := memory.NewMemoryStorage()
	handler := NewHandler(memStore, cfg)

	err := memStore.SaveURL(testShortURL, "https://testURL123.ru")
	require.NoError(t, err)

	tests := []struct { // мне надо передать: URL. Жду: код, original
		name     string
		shortUrl string
		want     struct {
			code        int
			originalUrl string
		}
	}{
		{
			name:     "valid shortURL",
			shortUrl: "/" + testShortURL,
			want: struct {
				code        int
				originalUrl string
			}{
				code:        http.StatusTemporaryRedirect,
				originalUrl: "https://testURL123.ru",
			},
		},
		{
			name:     "invalid shortURL",
			shortUrl: "/" + invalidShortUrl,
			want: struct {
				code        int
				originalUrl string
			}{
				code:        http.StatusBadRequest,
				originalUrl: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080"+tt.shortUrl, nil)
			w := httptest.NewRecorder()

			handler.GetUrl(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.originalUrl != "" {
				assert.Equal(t, tt.want.originalUrl, res.Header.Get("Location"))
			}
		})
	}

}
