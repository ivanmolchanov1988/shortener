package handlers

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ivanmolchanov1988/shortener/internal/memory"
	"github.com/ivanmolchanov1988/shortener/internal/server"
	"github.com/stretchr/testify/require"
)

// для теста конфига
var cfg = &server.Config{
	Address: "localhost:8080",
	BaseURL: "http://localhost:8080",
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
			name:        "valid_url_shgould_return_201_created",
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
			name:        "invalid_url_return_400_bad_request",
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
			name:        "invalid_Content-Type",
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

			handler.PostURL(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)

			//require.NoError(t, err)
			//assert.Equal(t, tt.want.code, resp.StatusCode)
			//assert.True(t, strings.HasPrefix(string(respBody), tt.want.response))
			//assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			if err != nil {
				t.Fatalf("Unable to read resp body: %v", err)
			}

			if resp.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, resp.StatusCode)
			}

			if tt.name == "valid_url_shgould_return_201_created" {
				if !strings.HasPrefix(string(respBody), tt.want.response) {
					t.Errorf("Expected response body to start with %q, got %q", tt.want.response, string(respBody))
				}
			} else {
				if string(respBody) != tt.want.response {
					t.Errorf("Expected response body %q, got %q", tt.want.response, string(respBody))
				}
			}

			if resp.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected content type %q, got %q", tt.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}

}

func TestShorten(t *testing.T) {
	memStore := memory.NewMemoryStorage()
	handler := NewHandler(memStore, cfg)

	urlToSend := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(urlToSend))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.Shorten(w, req)

	res := w.Result()
	defer res.Body.Close()

	// Тест для компресии
	var reader io.Reader
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(res.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gz.Close()
		reader = gz
	case "deflate":
		zlb, err := zlib.NewReader(res.Body)
		if err != nil {
			t.Fatalf("Failed to create zlib reader: %v", err)
		}
		defer zlb.Close()
		reader = zlb
	default:
		reader = res.Body
	}

	//body, err := io.ReadAll(res.Body)
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var responseData map[string]string
	if err := json.Unmarshal(body, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	shortURL, ok := responseData["result"]
	if !ok {
		t.Errorf("Response body does not contain 'result'")
	}

	expectedResponse := `{"result":"` + shortURL + `"}` + "\n"
	if string(body) != expectedResponse {
		t.Errorf("Expected response body %q, got %q", expectedResponse, string(body))
	}

	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", res.Header.Get("Content-Type"))
	}
}

func TestGetUrl(t *testing.T) {

	//запись для тестов
	testShortURL := "testURL"
	invalidShortURL := "123321"

	memStore := memory.NewMemoryStorage()
	handler := NewHandler(memStore, cfg)

	err := memStore.SaveURL(testShortURL, "https://testURL123.ru")
	require.NoError(t, err)

	tests := []struct { // мне надо передать: URL. Жду: код, original
		name     string
		shortURL string
		want     struct {
			code        int
			originalURL string
		}
	}{
		{
			name:     "valid shortURL",
			shortURL: "/" + testShortURL,
			want: struct {
				code        int
				originalURL string
			}{
				code:        http.StatusTemporaryRedirect,
				originalURL: "https://testURL123.ru",
			},
		},
		{
			name:     "invalid shortURL",
			shortURL: "/" + invalidShortURL,
			want: struct {
				code        int
				originalURL string
			}{
				code:        http.StatusNotFound,
				originalURL: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080"+tt.shortURL, nil)
			w := httptest.NewRecorder()

			handler.GetURL(w, req)

			res := w.Result()
			defer res.Body.Close()

			// assert.Equal(t, tt.want.code, res.StatusCode)
			// if tt.want.originalURL != "" {
			// 	assert.Equal(t, tt.want.originalURL, res.Header.Get("Location"))
			// }

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)
			}

			if res.Header.Get("Location") != tt.want.originalURL {
				t.Errorf("Expected location %q, got %q", tt.want.originalURL, res.Header.Get("Location"))
			}
		})
	}

}
