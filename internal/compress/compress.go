package compress

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

type zlibWriter struct {
	http.ResponseWriter
	writer *zlib.Writer
}

func (w zlibWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func NewCompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" || contentType == "text/html" {
			ae := r.Header.Get("Accept-Encoding")
			switch {
			case strings.Contains(ae, "gzip"):
				gz := gzip.NewWriter(w)
				defer gz.Close()
				gzw := &gzipWriter{ResponseWriter: w, writer: gz}
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Vary", "Accept-Encoding") // +Ответ может меняться
				next.ServeHTTP(gzw, r)
			case strings.Contains(ae, "deflate"):
				zl := zlib.NewWriter(w)
				defer zl.Close()
				zlw := &zlibWriter{ResponseWriter: w, writer: zl}
				w.Header().Set("Content-Encoding", "deflate")
				w.Header().Set("Vary", "Accept-Encoding") // +Ответ может меняться
				next.ServeHTTP(zlw, r)
			default:
				next.ServeHTTP(w, r)
			}
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

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
