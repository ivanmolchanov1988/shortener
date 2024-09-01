package compress

import (
	"compress/gzip"
	"compress/zlib"
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
