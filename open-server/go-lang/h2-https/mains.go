package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	//"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	port := flag.String("port", "8443", "HTTPS port")
	root := flag.String("root", ".", "root directory to serve")
	certDir := flag.String("certdir", ".certs", "directory to store/reuse certificate files")
	host := flag.String("host", "localhost", "primary host name for certificate (CN & SAN)")
	spa := flag.Bool("spa", true, "SPA fallback to index.html when 404")
	cors := flag.Bool("cors", true, "set Access-Control-Allow-Origin: *")
	regenerate := flag.Bool("regen", false, "force regenerate certificate even if exists")
	flag.Parse()

	_ = mime.AddExtensionType(".json", "application/json")

	// Đường dẫn cert/key cố định theo host
	if err := os.MkdirAll(*certDir, 0o755); err != nil {
		log.Fatalf("cannot create cert dir: %v", err)
	}
	certPath := filepath.Join(*certDir, fmt.Sprintf("%s.crt", *host))
	keyPath := filepath.Join(*certDir, fmt.Sprintf("%s.key", *host))

	// Tạo cert nếu chưa có (hoặc ép regen)
	if *regenerate || !fileExists(certPath) || !fileExists(keyPath) {
		log.Printf("Generating self-signed certificate for %q ...", *host)
		if err := generateSelfSignedCert(certPath, keyPath, *host); err != nil {
			log.Fatalf("generate cert error: %v", err)
		}
		log.Printf("Saved cert: %s , key: %s", certPath, keyPath)
	} else {
		log.Printf("Using existing cert: %s , key: %s", certPath, keyPath)
	}

	fs := http.FileServer(http.Dir(*root))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path

		// Cache dài hạn cho asset tĩnh
		if hasAnySuffix(p, ".css", ".js", ".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg", ".ico", ".woff", ".woff2") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		// CORS tối giản
		if *cors {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Nếu path có phần mở rộng -> serve trực tiếp
		if strings.Contains(filepath.Base(p), ".") {
			fs.ServeHTTP(w, r)
			return
		}

		// Thử serve, nếu 404 và bật SPA thì fallback index.html
		if *spa {
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			fs.ServeHTTP(rec, r)
			if rec.status == http.StatusNotFound {
				http.ServeFile(w, r, filepath.Join(*root, "index.html"))
			}
			return
		}

		fs.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:              ":" + *port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		// HTTP/2 sẽ tự bật với TLS của net/http khi cert hợp lệ
	}

	log.Printf("Serving HTTPS (h2) at https://0.0.0.0:%s  root=%s  certdir=%s  host=%s", *port, *root, *certDir, *host)
	log.Fatal(srv.ListenAndServeTLS(certPath, keyPath))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// --------- helpers ----------

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// generateSelfSignedCert tạo ECDSA P-256 key + cert tự ký, lưu ra file .crt/.key
// SAN gồm: host, "localhost", 127.0.0.1, ::1
func generateSelfSignedCert(certPath, keyPath, host string) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	// Subject & validity
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"Dev Local"},
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1 năm
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		BasicConstraintsValid: true,
	}

	// SAN
	tmpl.DNSNames = dedupStrings([]string{host, "localhost"})
	tmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Ghi cert
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		return err
	}

	// Ghi key (PKCS8 cho phổ biến)
	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return err
	}

	return nil
}

func dedupStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func hasAnySuffix(s string, exts ...string) bool {
	s = strings.ToLower(s)
	for _, e := range exts {
		if strings.HasSuffix(s, strings.ToLower(e)) {
			return true
		}
	}
	return false
}
