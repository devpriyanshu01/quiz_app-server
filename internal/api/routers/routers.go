package routers

import "net/http"

func MainRouters() *http.ServeMux {
	aRouters := AdminRouters()
	return aRouters
}