package routers

import (
	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router = SetStatusRoutes(router)
	router = SetHelloRoutes(router)
	router = SetInfobloxRoutes(router)
	router = SetAuthenticationRoutes(router)
	return router
}
