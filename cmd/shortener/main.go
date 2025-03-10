package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ivanmolchanov1988/shortener/internal/auth"
	"github.com/ivanmolchanov1988/shortener/internal/compress"
	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/storage"

	//"github.com/ivanmolchanov1988/shortener/internal/auth"
	"github.com/ivanmolchanov1988/shortener/internal/logger"
	"github.com/ivanmolchanov1988/shortener/internal/server"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	_ "net/http/pprof"
)

// goimports - nothing to commit, working tree clean

func generateCert() {
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Company"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := os.Create("certs/cert.pem")
	if err != nil {
		log.Fatalf("Failed to open cert.pem for writing: %v", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certOut.Close()

	keyOut, err := os.Create("certs/key.pem")
	if err != nil {
		log.Fatalf("Failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	keyOut.Close()

	log.Println("Certificates generated and saved to certs/")
}

func main() {
	server.Usage()

	// Канал для перехвата системных сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Канал завершения
	done := make(chan struct{})

	cfg, store, err := server.InitConfigAndPrepareStorage()
	if err != nil {
		log.Fatalf("INIT is failed: %v\n", err)
	}
	fmt.Printf("Initialized config: %+v\n", cfg)
	fmt.Printf("Storage type: %T\n", store)

	// отправим secret
	auth.GetTimeForExpire(cfg.TimeToExpire)

	// Логгер
	if err := logger.Initialize(cfg.Logging); err != nil {
		log.Fatalf("Logger initialization failed: %v\n", err)
	}

	// Хендлеры
	r := setupHandlers(store, cfg)

	// Создаём HTTP-сервер
	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Старт
	fmt.Printf("Server start: => %s\n\r", cfg.Address)

	// Запуск сервера в рутине
	go func() {
		if cfg.EnableHTTPS {
			if _, err := os.Stat("certs/cert.pem"); os.IsNotExist(err) {
				log.Println("Сертификаты не найдены, генерируем новые...")
				generateCert()
			}
			// Запус с cert
			log.Println("Запуск HTTPS сервера...")
			if err := http.ListenAndServeTLS(cfg.Address, "certs/cert.pem", "certs/key.pem", r); err != nil {
				fmt.Printf("Start with error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := http.ListenAndServe(cfg.Address, r); err != nil {
				fmt.Printf("Start with error: %v\n", err)
				os.Exit(1)
			}
		}
	}()

	// Сервер профилирования в рутине
	// go func() {
	// 	log.Println("Starting pprof server on localhost:6060")
	// 	log.Println(http.ListenAndServe("localhost:6060", nil)) // Сервер профилирования
	// }()
	pprofSrv := &http.Server{Addr: "localhost:6060"}
	go func() {
		log.Println("Starting pprof server on localhost:6060")
		if err := pprofSrv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("pprof server error: %v", err)
		}
	}()

	// Горутина обработки сигналов
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v. Shutting down...\n", sig)

		// Создаём контекст с таймаутом для завершения работы
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Отключаем keep-alive, чтобы закрыть соединения
		srv.SetKeepAlivesEnabled(false)

		log.Println("Calling srv.Shutdown()...")
		// Завершаем HTTP-сервер
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		log.Println("Server shutdown complete.")

		// Закрываем pprof
		log.Println("Shutting down pprof server...")
		if err := pprofSrv.Shutdown(ctx); err != nil {
			log.Printf("pprof Shutdown error: %v", err)
		}

		// Закрываем хранилище (если оно поддерживает закрытие)
		if closer, ok := store.(storage.Closer); ok {
			log.Println("Closing storage...")
			if err := closer.Close(); err != nil {
				log.Printf("Failed to close storage: %v", err)
			}
		}

		log.Println("All services stopped.")
		close(done) // Сообщаем, что сервер завершил работу
	}()

	// Ожидаем завершения сервера перед выходом
	<-done
	log.Println("Server gracefully stopped")

}

func setupHandlers(store storage.Storage, cfg *server.Config) http.Handler {
	//хэндлеры
	handler := handlers.NewHandler(store, cfg)
	r := chi.NewRouter()
	// Добавляем middleware логирования к каждому запросу
	r.Use(logger.RequestLogger)
	// Применяем middleware сжатия
	r.Use(compress.NewCompressHandler)
	r.Use(compress.DecompressHandler)
	r.Post("/", handler.PostURL)
	r.Post("/api/shorten", handler.Shorten)
	r.Get("/{id}", handler.GetURL)
	r.Get("/ping", handler.GetPingDB)
	r.Post("/api/shorten/batch", handler.Batch)
	r.Get("/api/user/urls", handler.GetUserURLS)
	r.Delete("/api/user/urls", handler.DeleteURLS)
	r.Get("/api/internal/stats", handler.GetStats)

	return r
}
