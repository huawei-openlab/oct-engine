package proxy

import (
	"../liboct"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
)

type OCISetting struct {
	URL      string
	Port     int
	CacheDir string
	Runtime  string
}

type OCIContainer struct {
	ContainerCommon
	BundleDir string
	Setting   OCISetting
}

func OCISettingNew() OCISetting {
	var ociSetting OCISetting
	config := "oci.config"

	content := liboct.ReadFile(config)
	json.Unmarshal([]byte(content), &ociSetting)

	return ociSetting
}

func OCIContainerNew(cc ContainerCommon) OCIContainer {
	var container OCIContainer
	container.ContainerCommon = cc

	container.Setting = OCISettingNew()
	container.BundleDir = path.Join(container.Setting.CacheDir, container.Setting.Runtime, container.Name)

	container.Init()
	return container
}

func (oci OCIContainer) Init() bool {
	liboct.PreparePath(oci.BundleDir, "")

	//check if runtime (runC for example) was exist

	return true

}

func (oci *OCIContainer) Hook() bool {
	oci.Name = "Change this name"
	return true
}

func (oci OCIContainer) Build() string {
	oci.Name = "Change this name"
	oci.BundleDir = "bundle changed"

	fmt.Println("OCI build", oci.Distribution)
	fmt.Println("No avaiable build server supported yet!")
	return "Change the name"
}

func (oci OCIContainer) Pull() bool {
	//this is the discovery policy actually
	fmt.Println("for current usage, do a temp work, download from oci-server and untar to the certain directory!\n")
	getUrl := fmt.Sprintf("https://%s:%d?name=%s", oci.Setting.URL, oci.Setting.Port, oci.Name)
	fmt.Println("In get report: ", getUrl)
	resp, err := http.Get(getUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	//	w.Write([]byte(string(resp_body)))

	return true
}

func (oci OCIContainer) Deploy() bool {
	//NOTE: deploy dir is not the case cache dir
	//copy oci.DeployDir to bundleDir/rootfs/oct-test
	testDir := path.Join(oci.BundleDir, "rootfs", "oct-test")
	liboct.PreparePath(testDir, "")

	return true
}

func (oci OCIContainer) Run() bool {
	//oci here also means runc.
	//any way to remember the runC open id?
	os.Chdir(oci.BundleDir)
	cmd := fmt.Sprintf("%s start", oci.Setting.Runtime)
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Run()

	return true
}

func (oci OCIContainer) Collect() bool {
	return true
}

func (oci OCIContainer) Destroy() bool {
	return true
}

func (oci OCIContainer) Status() string {
	return "{\"status\": \"ok\"}"
}
