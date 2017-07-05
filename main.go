package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/solidcoredata/scdhttp/scdhandler"
)

func main() {
	authSession := (&scdhandler.AuthenticateMemory{
		UserSetup: map[string]*scdhandler.MemoryUser{
			"user1": &scdhandler.MemoryUser{
				Identity:   "user1",
				GivenName:  "Myfirst",
				FamilyName: "Mylast",
				Password:   "password1",
			},
		},
	}).Init()
	// The authenticator needs to be per URL, NOT per RouteHandler or app.
	// On second thought, the authentication system should be tied to a single cookie name.
	// The reason I wanted per URL cookie names is to prevent programs running on different ports
	// from overwriting each other's cookies. One way to prevent this and to work around this
	// at least for now is to use a consistent hash scheme of the Hostname.
	//
	// The RouteHandler should not have Authenticator, it should be on the LoginStateRouter.
	// Additionally the LoginStateRouter should be able to pass along the authenticator to
	// all applications.
	//
	// I'd like to be able to support multiple QA environments on the same Host.
	// I can probably do this by configuring a special LoginGranted handler that
	// maps paths to nested-applications.
	h := (&scdhandler.RouteHandler{
		Router: scdhandler.HostRouter{
			"localhost:9786": scdhandler.LoginStateRouter{
				scdhandler.LoginNone:    (&scdhandler.LoginNoneHandler{Session: authSession}).Init(),
				scdhandler.LoginGranted: (&scdhandler.LoginGrantedHandler{Session: authSession}).Init(),
			},
			"localhost:9787": scdhandler.LoginStateRouter{
				scdhandler.LoginNone:    (&scdhandler.LoginNoneHandler{Session: authSession}).Init(),
				scdhandler.LoginGranted: (&scdhandler.LoginGrantedHandler{Session: authSession}).Init(),
			},
		},
		Authenticator: authSession,
		Loggerf: func(f string, a ...interface{}) {
			fmt.Printf(f, a...)
		},
	}).Init()

	for serveon := range h.Router {
		go func(serveon string) {
			err := http.ListenAndServe(serveon, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.URL.Host = serveon
				h.ServeHTTP(w, r)
			}))
			if err != nil {
				fmt.Fprintf(os.Stderr, "main: unable to listen and serve on %q: %v\n", serveon, err)
			}
		}(serveon)
	}
	fmt.Println("ready")
	select {}
}
