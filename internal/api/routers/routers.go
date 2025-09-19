package routers

import "net/http"

func MainRouters() *http.ServeMux {
	aRouters := AdminRouters()
	qRouters := QuizRouter()
	pRouters := PlayersRouters()

	aRouters.Handle("/", qRouters)
	qRouters.Handle("/", pRouters)
	return aRouters
}