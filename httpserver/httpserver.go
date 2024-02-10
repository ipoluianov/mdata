package httpserver

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ipoluianov/mdata/bybit"
	"github.com/ipoluianov/mdata/logger"
)

type Host struct {
	Name string
}

type HttpServer struct {
	srvTLS *http.Server
	rTLS   *mux.Router
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func NewHttpServer() *HttpServer {
	var c HttpServer
	return &c
}

func (c *HttpServer) Start() {
	logger.Println("HttpServer start")
	go c.thListenTLS()
}

func (c *HttpServer) thListenTLS() {
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = make([]tls.Certificate, 0)

	cert, err := tls.LoadX509KeyPair(CurrentExePath()+"/bundle.crt", CurrentExePath()+"/private.key")
	if err == nil {
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	} else {
		logger.Println("loading certificates error:", err.Error())
	}

	c.srvTLS = &http.Server{
		Addr:      ":8401",
		TLSConfig: tlsConfig,
	}

	c.rTLS = mux.NewRouter()
	c.rTLS.HandleFunc("/api/bybit/items", bybit.Items)
	c.rTLS.NotFoundHandler = http.HandlerFunc(c.processFile)
	c.srvTLS.Handler = c

	logger.Println("HttpServerTLS thListen begin")
	listener, err := tls.Listen("tcp", ":8401", tlsConfig)
	if err != nil {
		logger.Println("TLS Listener error:", err)
		return
	}

	err = c.srvTLS.Serve(listener)
	if err != nil {
		logger.Println("HttpServerTLS thListen error: ", err)
	}
	logger.Println("HttpServerTLS thListen end")
}

func (s *HttpServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.rTLS.ServeHTTP(rw, req)
}

func (c *HttpServer) Stop() error {
	var err error

	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = c.srvTLS.Shutdown(ctx); err != nil {
			logger.Println(err)
		}
	}
	return err
}

func SplitRequest(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/'
	})
}

func (c *HttpServer) processHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println("ProcessHTTP host: ", r.Host)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	c.processFile(w, r)
}

func (c *HttpServer) processFile(w http.ResponseWriter, r *http.Request) {
	c.file(w, r, r.URL.Path)
}

func (c *HttpServer) fullpath(url string, host string) (string, error) {
	result := ""

	result = CurrentExePath() + "/data/" + url

	fi, err := os.Stat(result)
	if err == nil {
		if fi.IsDir() {
			result += "/index.html"
		}
	}

	return result, err
}

func (c *HttpServer) file(w http.ResponseWriter, r *http.Request, urlPath string) {
	var err error
	var fileContent []byte
	var writtenBytes int

	realIP := getRealAddr(r)

	logger.Println("Real IP: ", realIP)
	logger.Println("HttpServer processFile: ", r.URL.String())

	var urlUnescaped string
	urlUnescaped, err = url.QueryUnescape(urlPath)
	if err == nil {
		urlPath = urlUnescaped
	}

	if urlPath == "/" || urlPath == "" {
		urlPath = "/index.html"
	}

	url, err := c.fullpath(urlPath, r.Host)

	logger.Println("FullPath: " + url)

	if strings.Contains(url, "..") {
		logger.Println("Wrong FullPath")
		w.WriteHeader(404)
		return
	}

	if err != nil {
		w.WriteHeader(404)
		return
	}

	fileContent, err = ioutil.ReadFile(url)

	if err == nil {
		w.Header().Set("Content-Type", c.contentTypeByExt(filepath.Ext(url)))
		writtenBytes, err = w.Write(fileContent)
		if err != nil {
			logger.Println("HttpServer sendError w.Write error:", err)
		}
		if writtenBytes != len(fileContent) {
			logger.Println("HttpServer sendError w.Write data size mismatch. (", writtenBytes, " / ", len(fileContent))
		}
	} else {
		logger.Println("HttpServer processFile error: ", err)
		w.WriteHeader(404)
	}
}

func (c *HttpServer) contentTypeByExt(ext string) string {
	var builtinTypesLower = map[string]string{
		".css":  "text/css; charset=utf-8",
		".gif":  "image/gif",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".js":   "text/javascript; charset=utf-8",
		".mjs":  "text/javascript; charset=utf-8",
		".pdf":  "application/pdf",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".wasm": "application/wasm",
		".webp": "image/webp",
		".xml":  "text/xml; charset=utf-8",
	}

	logger.Println("Ext: ", ext)

	if ct, ok := builtinTypesLower[ext]; ok {
		return ct
	}
	return "text/plain"
}

func getRealAddr(r *http.Request) string {

	remoteIP := ""
	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		remoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			remoteIP = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			remoteIP = ip.String()
		}
	}

	return remoteIP

}
