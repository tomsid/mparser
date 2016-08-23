package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tomsid/mparser/parser"
)

const (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = "9080"

	ENV_LISTEN_PORT   = "LISTEN_PORT"
	ENV_LISTEN_HOST   = "LISTEN_HOST"
	ENV_SSL_ON        = "SSL_ON"
	ENV_SSL_CERT_PATH = "SSL_CERT_PATH"
	ENV_SSL_KEY_PATH  = "SSL_KEY_PATH"
)

func messageHandler(w http.ResponseWriter, r *http.Request) {
	l := log.New(os.Stdout, "messageHandler: ", log.LstdFlags)

	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("An error occured"))
		l.Print("Unable to read body")
		return
	}
	defer r.Body.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	p := parser.NewParser(client, l)
	mInfo := p.Parse(string(response))

	response, err = mInfo.MarshalJSON()
	if err != nil {
		w.Write([]byte("An error occured"))
		l.Print("Unable to marshal JSON")
		return
	}

	w.Header().Set("Content-Type", " application/json; charset=utf-8")
	w.Write(response)
}

func main() {
	http.DefaultServeMux.HandleFunc("/", messageHandler)

	host := os.Getenv(ENV_LISTEN_HOST)
	port := os.Getenv(ENV_LISTEN_PORT)
	sslOn := os.Getenv(ENV_SSL_ON)

	if len(port) <= 0 {
		port = DEFAULT_PORT
	}

	if len(host) <= 0 {
		host = DEFAULT_HOST
	}

	server := http.Server{
		Addr:         host + ":" + port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if len(sslOn) > 0 && sslOn == "1" {
		certPath := os.Getenv(ENV_SSL_CERT_PATH)
		keyPath := os.Getenv(ENV_SSL_KEY_PATH)

		if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
			log.Fatalf("Unable to load key pair. Error: %s", err)
		}

		log.Printf("Listenting on https://%s:%s", host, port)
		server.ListenAndServeTLS(certPath, keyPath)
	} else {
		log.Printf("Listenting on http://%s:%s", host, port)
		server.ListenAndServe()
	}
}
