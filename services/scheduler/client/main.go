// Copyright 2015 Huawei Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/huawei-openlab/oct-engine/liboct"
)

const TestCache = "testCache"

func main() {
	app := cli.NewApp()
	app.Name = "oct"
	app.Version = "0.1.0"
	app.Usage = "OCT Engine CLient, Used to run the test from a single case/case.tar.gz"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path, p",
			Value: "",
			Usage: "path of the case, -p=case.tar.gz or -p=case.json, or -p=caseDir",
		},
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the running task",
		},
		cli.StringFlag{
			Name:  "scheduler-address, saddr",
			Value: "http://localhost:8001",
			Usage: "Scheduler Server address",
		},
	}

	app.Action = func(c *cli.Context) {
		casePath := c.String("path")
		sAddr := c.String("scheduler-address")
		ID := c.String("id")
		if casePath != "" {
			RunTest(casePath, sAddr)
		} else if ID != "" {
			QueryTest(ID, sAddr)
		} else {
			logrus.Fatal("Case path and the task ID cannot be empty at the same time")
		}

	}

	logrus.SetLevel(logrus.InfoLevel)

	if err := app.Run(os.Args); err != nil {
		logrus.Fatalf("Run App err %v\n", err)
	}
}

func RunTest(casePath string, sAddr string) {
	var caseBundle string
	var caseTar string

	if p, err := os.Stat(casePath); err != nil {
		logrus.Fatal(err)
	} else if p.IsDir() {
		caseBundle = casePath
	} else {
		if _, err := os.Stat(TestCache); err != nil {
			os.MkdirAll(TestCache, 0777)
		}
		caseBundle, _ = ioutil.TempDir(TestCache, "oct-")
		defer os.RemoveAll(caseBundle)
		if strings.HasSuffix(casePath, ".json") {
			copy(path.Join(caseBundle, "case.json"), casePath)
		} else if strings.HasSuffix(casePath, ".tar.gz") {
			caseTar = casePath
			liboct.UntarFile(casePath, caseBundle)
		} else {
			logrus.Fatalf("%s is not a valid test case", casePath)
		}
	}

	//Check if it is a valid test case
	if _, err := liboct.CaseFromBundle(caseBundle); err != nil {
		logrus.Fatal(err)
	}
	if len(caseTar) == 0 {
		caseTar = liboct.TarDir(caseBundle)
		defer os.Remove(caseTar)
	}

	logrus.Debugf("Bundle %s, tar %s, sending to %s", caseBundle, caseTar, sAddr)

	params := make(map[string]string)
	//params["id"] = liboct.MD5(fmt.Sprintf("%d", time.Now().Unix()))

	postURL := fmt.Sprintf("%s/task", sAddr)
	ret := liboct.SendFile(postURL, caseTar, params)
	if ret.Status != liboct.RetStatusOK {
		logrus.Warnf("Failed to apply run task %v", ret)
		return
	} else {
		logrus.Debugf("Success in apply run task %v", ret)
	}

	taskID := ret.Message
	postURL = fmt.Sprintf("%s/task/%s", sAddr, taskID)
	ret = liboct.SendCommand(postURL, []byte("deploy"))
	if ret.Status != liboct.RetStatusOK {
		logrus.Warnf("Failed to deploy task %v", ret)
		return
	} else {
		logrus.Debugf("Success in deploy task %v", ret)
	}

	ret = liboct.SendCommand(postURL, []byte("run"))
	if ret.Status != liboct.RetStatusOK {
		logrus.Warnf("Failed to run task %v", ret)
		return
	} else {
		logrus.Debugf("Success in run task %v", ret)
	}

	ret = liboct.SendCommand(postURL, []byte("collect"))
	if ret.Status != liboct.RetStatusOK {
		logrus.Warnf("Failed to run task %v", ret)
		return
	} else {
		logrus.Debugf("Success in run task %v", ret)
	}

	getURL := fmt.Sprintf("%s/task/%s/report", sAddr, taskID)
	resp, err := http.Get(getURL)
	if err != nil {
		logrus.Warnf("Failed to get report")
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Warnf("The report is empty")
		return
	}

	reportTar := fmt.Sprintf("%s/%s-report.tar.gz", TestCache, taskID)
	ioutil.WriteFile(reportTar, respBody, 0644)
	liboct.UntarFile(reportTar, fmt.Sprintf("%s/%s", TestCache, taskID))
	logrus.Infof("Success in run the test, the report generated here:\n%v", fmt.Sprintf("%s/%s", TestCache, taskID))
	os.Remove(reportTar)
}

func QueryTest(ID string, sAddr string) {
}

func copy(dst string, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}
