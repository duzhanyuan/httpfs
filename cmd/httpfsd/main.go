package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/namsral/flag"

	"github.com/prologic/httpfs/webapi"
)

func cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}

// Log ...
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	var (
		config  string
		tls     bool
		tlscert string
		tlskey  string
		debug   bool
		bind    string
		root    string
	)

	flag.StringVar(&config, "config", "", "config file")
	flag.BoolVar(&tls, "tls", false, "Use TLS")
	flag.BoolVar(&debug, "debug", false, "set debug logging")
	flag.StringVar(&tlscert, "tlscert", "server.crt", "server certificate")
	flag.StringVar(&tlskey, "tlskey", "server.key", "server key")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")
	flag.StringVar(&root, "root", cwd(), "path to serve")
	flag.Parse()

	if !debug {
		log.SetOutput(ioutil.Discard)
	}

	http.Handle("/", webapi.FileServer(root))

	var handler http.Handler

	if debug {
		handler = Log(http.DefaultServeMux)
	} else {
		handler = http.DefaultServeMux
	}

	if tls {
		log.Fatal(
			http.ListenAndServeTLS(
				bind,
				tlscert,
				tlskey,
				handler,
			),
		)
	} else {
		log.Fatal(http.ListenAndServe(bind, handler))
	}
}
