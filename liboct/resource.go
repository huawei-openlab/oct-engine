//This file is used for the 'Schedular'
package liboct

import (
	"encoding/json"
	"errors"

	"github.com/Sirupsen/logrus"
)

type ResourceStatus string

const (
	ResourceStatusFree   ResourceStatus = "free"
	ResourceStatusLocked                = "locked"
)

type ResourceClass string

const (
	ResourceOS        ResourceClass = "os"
	ResourceContainer               = "container"
)

//Common is the request from the user
type ResourceCommon struct {
	Class        ResourceClass
	Distribution string
	Version      string
	Arch         string
	CPU          int64
	Memory       int64
}

//This is record in the testing cluster
type Resource struct {
	ResourceCommon
	ID  string
	URL string

	//0 means no limit
	MaxJobs     int
	TestUnitIDs []string
	Status      ResourceStatus
}

func (res Resource) String() string {
	val, _ := json.Marshal(res)
	return string(val)
}

func ResourceFromString(val string) (res Resource, err error) {
	err = json.Unmarshal([]byte(val), &res)
	return res, err
}

func (res *Resource) IsValid() error {
	//TODO
	if res.Class == "" {
		return errors.New("'Class' required.")
	} else if res.Distribution == "" {
		return errors.New("'Distribution' required.")
	} else if res.Version == "" {
		return errors.New("'Version' required.")
	} else if res.Arch == "" {
		return errors.New("'Arch' required.")
	} else if res.URL == "" {
		return errors.New("'URL' required.")
	}
	return nil
}

func (res *Resource) SetID(id string) {
	if res.ID != id {
		res.ID = id
	}
}

func (res *Resource) GetID() string {
	return res.ID
}

func (res *Resource) IsQualify(req ResourceCommon) bool {
	if req.Distribution != res.Distribution {
		logrus.Debugf("Distribution not fit")
		return false
	}
	if req.Version != res.Version {
		logrus.Debugf("Version not fit")
		return false
	}
	if req.Arch != res.Arch {
		logrus.Debugf("Arch not fit")
		return false
	}

	//TODO: better calculation
	if req.CPU > res.CPU {
		logrus.Debugf("CPU not fit")
		return false
	}
	if req.Memory > res.Memory {
		logrus.Debugf("Memory not fit")
		return false
	}
	return true
}

func (res *Resource) Allocate(unit TestUnit) bool {
	if res.MaxJobs == 0 || res.MaxJobs > len(res.TestUnitIDs) {
		res.TestUnitIDs = append(res.TestUnitIDs, unit.GetID())
		return true
	}
	return false
}
