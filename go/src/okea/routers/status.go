package routers

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"okea/controllers"
	//	"okea/core/authentication"
)

func SetStatusRoutes(router *mux.Router) *mux.Router {
	router.Handle("/status",
		negroni.New(
			// Don't require login for status info
			// negroni.HandlerFunc(authentication.RequireTokenAuthentication),
			negroni.HandlerFunc(controllers.StatusController),
		)).Methods("GET")

	return router
}
