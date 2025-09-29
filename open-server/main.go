package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	// Cho phép nhập cổng qua flag, mặc định 9999
	port := flag.String("port", "9999", "port to serve on")
	root := flag.String("root", ".", "root directory to serve")
	flag.Parse()

	fs := http.FileServer(http.Dir(*root))

	// Middleware để set cache + CORS + SPA fallback
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, ".css") || strings.HasSuffix(p, ".js") ||
			strings.HasSuffix(p, ".png") || strings.HasSuffix(p, ".jpg") ||
			strings.HasSuffix(p, ".jpeg") || strings.HasSuffix(p, ".webp") ||
			strings.HasSuffix(p, ".svg") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Fallback index.html nếu là SPA
		fpath := filepath.Clean(*root + p)
		if strings.Contains(filepath.Base(fpath), ".") {
			fs.ServeHTTP(w, r)
			return
		}
		rw := &respWriter{ResponseWriter: w, status: 200}
		fs.ServeHTTP(rw, r)
		if rw.status == 404 {
			http.ServeFile(w, r, filepath.Join(*root, "index.html"))
		}
	})

	addr := ":" + *port
	log.Printf("Serving %s on http://0.0.0.0%s\n", *root, addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

type respWriter struct {
	http.ResponseWriter
	status int
}

func (w *respWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
