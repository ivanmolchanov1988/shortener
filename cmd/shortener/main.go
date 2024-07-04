package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Временное хранилище
var tempUrlStore = make(map[string]string)

// Рандомные строки для короткой ссылки
func randStr(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}

// Инкремент_1
// #0 путь = /
// #1 POST проверка
// #2 Header Content-Type = text/plain
// #3 возвращает ответ с кодом 201
// #4 Сервер принимает в теле запроса строку URL как text/plain
// #5 возвращает ответ с сокращённым URL
// #6 Эндпоинт с методом GET
// #7 id — идентификатор сокращённого URL
// #8 В случае успешной обработки запроса сервер возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location

// //////// POST //////////
func postUrl(res http.ResponseWriter, req *http.Request) {
	// #1 проверка на POST
	if req.Method != http.MethodPost {
		http.Error(res, "Only the POST method is available", http.StatusBadRequest)
		return
	}
	// #4.1 проверка URL как text/plain
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}
	// #4.2 Сервер принимает в теле запроса строку URL
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Unable to read body", http.StatusBadRequest)
		return
	}
	urlStr := string(body)
	_, err = url.ParseRequestURI(urlStr)
	if err != nil {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Забираем рандомную строку для ссылки
	shortUrl, err := randStr(8)
	if err != nil {
		http.Error(res, "Unable to generate short URL", http.StatusBadRequest)
		return
	}
	// Сохраним URL's в map tempUrlStore
	tempUrlStore[shortUrl] = urlStr

	// #2 Header Content-Type = text/plain
	res.Header().Set("Content-Type", "text/plain")
	// #3 res code = 201
	res.WriteHeader(http.StatusCreated)
	// #5 возвращает ответ с сокращённым URL
	fullShortUrl := fmt.Sprintf("http://%s/%s", req.Host, shortUrl)
	res.Write([]byte(fullShortUrl))

	// res.Write([]byte("Тут будет POST '/' URL = text/plain и ответ = 201 + сокращённый URL как text/plain"))
}

// //////// GET //////////
func getUrl(res http.ResponseWriter, req *http.Request) {
	// #6 проверка на GET
	if req.Method != http.MethodGet {
		http.Error(res, "Only the GET method is available", http.StatusBadRequest)
		return
	}

	// #7 парсинг ссылки
	idLink := strings.TrimPrefix(req.URL.Path, "/")
	if idLink == "" {
		http.Error(res, "This ID does not exist", http.StatusBadRequest)
		// res.Header().Set("Content-Type", "text/plain")
		// res.Write([]byte("Shortener"))
		return
	}
	// #8 возвращение исходной ссылки и 307 в HTTP-заголовке Location
	originUrl, ok := tempUrlStore[idLink]
	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)

	// res.Write([]byte("Тут будет GET '/{id}/' id = идентификатор сокращённого URL и ответ = 307 + оригинальным URL в HTTP-заголовке Location"))
}

func main() {
	// #0 путь = /
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost && req.URL.Path == "/" {
			postUrl(res, req)
		} else if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/") {
			getUrl(res, req)
		} else {
			http.Error(res, "Not Found", http.StatusBadRequest)
		}
	})

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
