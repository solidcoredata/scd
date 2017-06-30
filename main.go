package main

import (
	"flag"
	"net/http"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

func main() {
	httpf := flag.String("http", ":9786", "listener address")
	flag.Parse()

	http.ListenAndServe(*httpf, &scdhandler.Handler{})
}
