package routers

import "net/http"

func MainRouters() *http.ServeMux {
	aRouters := AdminRouters()
	qRouters := QuizRouter()

	aRouters.Handle("/", qRouters)
	return aRouters
}