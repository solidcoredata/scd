package main

import (
	"fmt"
	"net/http"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

func main() {
	auth := (&scdhandler.AuthenticateMemory{
		KeyName: "app1",
		UserSetup: map[string]*scdhandler.MemoryUser{
			"user1": &scdhandler.MemoryUser{
				Identity:   "user1",
				GivenName:  "Myfirst",
				FamilyName: "Mylast",
				Password:   "password1",
			},
		},
	}).Init()
	h := (&scdhandler.RouteHandler{
		Router: scdhandler.URLRouter{
			"localhost:9786": scdhandler.LoginStateRouter{
				scdhandler.LoginNone:    (&scdhandler.LoginNoneHandler{Session: auth}).Init(),
				scdhandler.LoginGranted: (&scdhandler.LoginGrantedHandler{Session: auth}).Init(),
			},
		},
		Authenticator: auth,
		Loggerf: func(f string, a ...interface{}) {
			fmt.Printf(f, a...)
		},
	}).Init()

	for serveon := range h.Router {
		go func(serveon string) {
			http.ListenAndServe(serveon, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.URL.Host = serveon
				h.ServeHTTP(w, r)
			}))
		}(serveon)
	}
	fmt.Println("ready")
	select {}
}
