package routers

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"okea/controllers"
	"okea/core/authentication"
)

func SetHelloRoutes(router *mux.Router) *mux.Router {
	router.Handle("/test/hello",
		negroni.New(
			negroni.HandlerFunc(authentication.RequireTokenAuthentication),
			negroni.HandlerFunc(controllers.HelloController),
		)).Methods("GET")

	return router
}
