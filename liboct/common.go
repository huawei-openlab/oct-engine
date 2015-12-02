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
	"os"
	"os/exec"
	"path"

	"github.com/Sirupsen/logrus"
)

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
		logrus.Debugf("Tar file %v", sourceFile)
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

		h := new(tar.Header)
		h.Name = sourceFile
		h.Size = fi.Size()
		h.Mode = int64(fi.Mode())
		h.ModTime = fi.ModTime()
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
	logrus.Debugf("UntarFile %v %v", filename, cacheURL)
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
		dir := path.Dir(filename)
		if _, err := os.Stat(dir); err != nil {
			os.MkdirAll(dir, 0777)
		}
		fw, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.FileMode(h.Mode))
		if err != nil {
			//Dir for example
			continue
		} else {
			io.Copy(fw, tr)
			fw.Close()
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

func ExecSH(scripts string, dir string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", "-c", scripts)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	return cmd.CombinedOutput()
}
