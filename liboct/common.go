package liboct

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
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

//When filename is null, we just want to prepare a pure directory
func PreparePath(cachename string, filename string) (dir string) {
	if filename == "" {
		dir = cachename
	} else {
		realurl := path.Join(cachename, filename)
		dir = path.Dir(realurl)
	}
	p, err := os.Stat(dir)
	if err != nil {
		if !os.IsExist(err) {
			os.MkdirAll(dir, 0777)
		}
	} else {
		if !p.IsDir() {
			os.Remove(dir)
			os.MkdirAll(dir, 0777)
		}
	}
	return dir
}

func SendFile(postURL string, fileURL string, params map[string]string) (ret HttpRet) {
	logrus.Debugf("SendFile ", postURL, fileURL)
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
		logrus.Debugf("SendFile key", key, "  val ", val)
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

//TODO: add err para?
func ReadFile(fileURL string) (content string) {
	_, err := os.Stat(fileURL)
	if err != nil {
		logrus.Warn(err)
		return content
	}
	file, err := os.Open(fileURL)
	defer file.Close()
	if err != nil {
		logrus.Warn(err)
		return content
	}
	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)
	content = buf.String()

	return content
}

func ReceiveFile(w http.ResponseWriter, r *http.Request, cacheURL string) (realURL string, params map[string]string) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("tcfile")

	logrus.Debugf("Receive file ", cacheURL, " ", handler.Filename)
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
		realDir = PreparePath(cacheURL, "")
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

// file name filelist is like this: './source/file'
func TarFileList(filelist []string, caseDir string, object_name string) (tarURL string) {
	tarURL = path.Join(caseDir, object_name) + ".tar.gz"
	fw, err := os.Create(tarURL)
	if err != nil {
		logrus.Warn(err)
		return tarURL
	}
	defer fw.Close()
	gw := gzip.NewWriter(fw)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for index := 0; index < len(filelist); index++ {
		sourceFile := filelist[index]
		logrus.Debugf("Tar file %s", sourceFile)
		fi, err := os.Stat(path.Join(caseDir, sourceFile))
		if err != nil {
			logrus.Warn(err)
			continue
		}
		fr, err := os.Open(path.Join(caseDir, sourceFile))
		if err != nil {
			logrus.Warn(err)
			continue
		}
		h, _ := tar.FileInfoHeader(fi, "")
		err = tw.WriteHeader(h)
		_, err = io.Copy(tw, fr)
	}
	return tarURL
}

func GetDirFiles(base_dir string, dir string) (files []string) {
	files_info, _ := ioutil.ReadDir(path.Join(base_dir, dir))
	for _, file := range files_info {
		if file.IsDir() {
			sub_files := GetDirFiles(base_dir, path.Join(dir, file.Name()))
			for _, sub_file := range sub_files {
				files = append(files, sub_file)
			}
		} else {
			files = append(files, path.Join(dir, file.Name()))
		}
	}
	return files

}

func TarDir(caseDir string) (tarURL string) {
	files := GetDirFiles(caseDir, "")
	case_name := path.Base(caseDir)
	tarURL = TarFileList(files, caseDir, case_name)
	return tarURL
}

func UntarFile(filename string, cacheURL string) {
	logrus.Debugf("UntarFile ", filename, cacheURL)
	_, err := os.Stat(filename)
	if err != nil {
		logrus.Warn(err)
		return
	}

	fr, err := os.Open(filename)
	if err != nil {
		logrus.Warn(err)
		return
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		logrus.Warn(err)
		return
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Fatal(err)
			panic(err)
		}
		filename := path.Join(cacheURL, h.Name)
		switch h.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(filename, os.FileMode(h.Mode))
		case tar.TypeReg:
			fw, err := os.Create(filename)
			if err != nil {
				continue
			} else {
				io.Copy(fw, tr)
				fw.Close()
			}
		case tar.TypeLink:
			os.Link(h.Linkname, filename)
		case tar.TypeSymlink:
			os.Symlink(h.Linkname, filename)
		}
		//TODO: set the time/own and the etc..
	}
}

func ReadCaseFromTar(tarURL string) (content string) {
	_, err := os.Stat(tarURL)
	if err != nil {
		logrus.Warn(err)
		return content
	}

	fr, err := os.Open(tarURL)
	if err != nil {
		logrus.Warn(err)
		return content
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		logrus.Warn(err)
		return content
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		fileSuffix := path.Ext(h.Name)
		if fileSuffix == ".json" {
			var tc TestCase
			buf := bytes.NewBufferString("")
			buf.ReadFrom(tr)
			fileContent := buf.String()
			json.Unmarshal([]byte(fileContent), &tc)
			if len(tc.Name) > 1 {
				content = fileContent
				break
			} else {
				continue
			}
		}
	}

	return content
}

//fileURL is the default file, suffix is the potential file
func ReadTar(tarURL string, fileURL string, suffix string) (content string) {
	_, err := os.Stat(tarURL)
	if err != nil {
		logrus.Warn(err)
		return content
	}

	fr, err := os.Open(tarURL)
	if err != nil {
		logrus.Warn(err)
		return content
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		logrus.Warn(err)
		return content
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Fatal(err)
			panic(err)
		}
		if len(suffix) > 0 {
			fileSuffix := path.Ext(h.Name)
			if fileSuffix == suffix {
				buf := bytes.NewBufferString("")
				buf.ReadFrom(tr)
				content = buf.String()
				break
			}
		}
		if len(fileURL) > 0 {
			if h.Name == fileURL {
				buf := bytes.NewBufferString("")
				buf.ReadFrom(tr)
				content = buf.String()
				break
			}
		}
	}

	return content
}

func MD5(data string) (val string) {
	t := md5.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))

}

func ExecSH(scripts string, dir string) bool {
	cmd := exec.Command("/bin/sh", "-c", scripts)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	if _, err := cmd.CombinedOutput(); err != nil {
		return false
	} else {
		return true
	}
}
