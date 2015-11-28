package liboct

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type RetStatus string

const (
	RetStatusOK     RetStatus = "ok"
	RetStatusFailed RetStatus = "failed"
)

type HttpRet struct {
	Status  RetStatus
	Message string
	Data    interface{}
}

func Render(w http.ResponseWriter, httpStatus int, ret HttpRet) {
	w.WriteHeader(httpStatus)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result, _ := json.MarshalIndent(ret, "", "\t")
	w.Write(result)
}

func RenderErrorf(w http.ResponseWriter, msg string) {
	var ret HttpRet
	ret.Status = RetStatusFailed
	ret.Message = msg

	logrus.Warnf(msg)

	httpStatus := http.StatusServiceUnavailable
	Render(w, httpStatus, ret)
}

func RenderError(w http.ResponseWriter, err error) {
	var ret HttpRet
	ret.Status = RetStatusFailed
	ret.Message = err.Error()

	logrus.Warn(err)

	//TODO: Check error for better result?
	httpStatus := http.StatusServiceUnavailable
	Render(w, httpStatus, ret)
}

func RenderOK(w http.ResponseWriter, msg string, data interface{}) {
	var ret HttpRet
	ret.Status = RetStatusOK
	ret.Message = msg
	ret.Data = data

	httpStatus := http.StatusOK
	Render(w, httpStatus, ret)
}
