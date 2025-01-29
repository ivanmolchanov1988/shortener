package compress

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

var zlibPool = sync.Pool{
	New: func() interface{} {
		return zlib.NewWriter(nil)
	},
}

func getGzipWriter(w io.Writer) *gzip.Writer {
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(w)
	return gz
}

func putGzipWriter(gz *gzip.Writer) {
	gz.Close()
	gzipPool.Put(gz)
}

func getZlibWriter(w io.Writer) *zlib.Writer {
	zl := zlibPool.Get().(*zlib.Writer)
	zl.Reset(w)
	return zl
}

func putZlibWriter(zl *zlib.Writer) {
	zl.Close()
	zlibPool.Put(zl)
}

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

// Write сжимает данные в gzip.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

type zlibWriter struct {
	http.ResponseWriter
	writer *zlib.Writer
}

// Write сжимает данные в zlib.
func (w zlibWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func shouldCompress(contentLength int) bool {
	const minCompressSize = 1024 // Минимальный размер для сжатия (1 KB)
	return contentLength > minCompressSize
}

func isCompressible(contentType string) bool {
	supportedTypes := []string{"application/json", "text/html", "application/x-gzip"}
	for _, t := range supportedTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// NewCompressHandler сжимает данные.
func NewCompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		contentLength := r.ContentLength

		if !shouldCompress(int(contentLength)) || !isCompressible(contentType) {
			next.ServeHTTP(w, r)
			return
		}

		ae := r.Header.Get("Accept-Encoding")
		switch {
		case strings.Contains(ae, "gzip"):
			gz := getGzipWriter(w)
			defer putGzipWriter(gz)
			gzw := &gzipWriter{ResponseWriter: w, writer: gz}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding") // +Ответ может меняться
			next.ServeHTTP(gzw, r)
		case strings.Contains(ae, "deflate"):
			zl := getZlibWriter(w)
			defer putZlibWriter(zl)
			zlw := &zlibWriter{ResponseWriter: w, writer: zl}
			w.Header().Set("Content-Encoding", "deflate")
			w.Header().Set("Vary", "Accept-Encoding") // +Ответ может меняться
			next.ServeHTTP(zlw, r)
		default:
			next.ServeHTTP(w, r)
		}

	})
}

// DecompressHandler декодирует данные.
func DecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.ReadCloser
		switch r.Header.Get("Content-Encoding") {
		case "gzip":
			var err error
			reader, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress gzip body", http.StatusBadRequest)
				return
			}
			defer reader.Close()
			r.Body = reader
			r.Header.Del("Content-Encoding") // Уже декодировано
		case "deflate":
			reader = flate.NewReader(r.Body)
			defer reader.Close()
			r.Body = reader
			r.Header.Del("Content-Encoding") // Удалить
		}
		next.ServeHTTP(w, r)
	})
}
