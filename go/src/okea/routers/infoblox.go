package routers

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"okea/controllers"
	"okea/core/authentication"
)

func SetInfobloxRoutes(router *mux.Router) *mux.Router {
	router.Handle("/cnames/{project}",
		negroni.New(
			negroni.HandlerFunc(authentication.RequireTokenAuthentication),
			negroni.HandlerFunc(controllers.ListCNamesForProject),
		)).Methods("GET")

	return router
}
