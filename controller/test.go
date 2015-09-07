package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type TestController interface {
	RegenerateDefault(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

func (c *controller) RegenerateDefault(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}
