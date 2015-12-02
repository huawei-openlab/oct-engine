package liboct

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"

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
	//FIXME: Not found?
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

func SendFile(postURL string, fileURL string, params map[string]string) (ret HttpRet) {
	logrus.Debugf("SendFile %v %v ", postURL, fileURL)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	filename := path.Base(fileURL)
	//'tcfile': testcase file
	fileWriter, err := bodyWriter.CreateFormFile("tcfile", filename)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
		return ret
	}
	_, err = os.Stat(fileURL)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
		return ret
	}

	fh, err := os.Open(fileURL)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
		return ret
	}
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
		return ret
	}

	for key, val := range params {
		logrus.Debugf("SendFile key %v val %v", key, val)
		_ = bodyWriter.WriteField(key, val)
	}
	//	contentType := bodyWriter.FormDataContentType()

	bodyWriter.Close()
	request, err := http.NewRequest("POST", postURL, bodyBuf)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
		return ret
	}
	request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
		return ret
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ret.Status = RetStatusFailed
		ret.Message = err.Error()
	} else {
		json.Unmarshal([]byte(resp_body), &ret)
	}
	return ret

}

func SendCommand(apiurl string, b []byte) (ret HttpRet) {
	body := bytes.NewBuffer(b)
	resp, perr := http.Post(apiurl, "application/json;charset=utf-8", body)
	if perr != nil {
		logrus.Warn(perr)
		ret.Status = RetStatusFailed
		ret.Message = perr.Error()
	} else {
		result, berr := ioutil.ReadAll(resp.Body)
		if berr != nil {
			ret.Status = RetStatusFailed
			ret.Message = berr.Error()
		} else {
			json.Unmarshal([]byte(result), &ret)
		}
		resp.Body.Close()
	}
	return ret
}

func ReceiveFile(w http.ResponseWriter, r *http.Request, cacheURL string) (realURL string, params map[string]string) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("tcfile")

	logrus.Debugf("Receive file %v %v ", cacheURL, handler.Filename)
	params = make(map[string]string)

	if r.MultipartForm != nil {
		for key, val := range r.MultipartForm.Value {
			//Must use val[0]
			params[key] = val[0]
		}
	}

	if err != nil {
		logrus.Warn(err)
		return realURL, params
	}

	defer file.Close()
	var realDir string
	//Receive to the cache/task_id dir
	if val, ok := params["id"]; ok {
		realDir = PreparePath(path.Join(cacheURL, val), "")
	} else {
		_, err := os.Stat(cacheURL)
		if err != nil && !os.IsExist(err) {
			os.MkdirAll(cacheURL, 0777)
		}
		realDir, _ = ioutil.TempDir(cacheURL, "oct-received-file-")
	}
	realURL = fmt.Sprintf("%s/%s", realDir, handler.Filename)
	f, err := os.Create(realURL)
	if err != nil {
		logrus.Warn(err)
		//TODO: better system error
		http.Error(w, err.Error(), 500)
		return realURL, params
	}
	defer f.Close()
	io.Copy(f, file)

	return realURL, params
}
