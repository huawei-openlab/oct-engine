package main

import (
	"../../lib/libocit"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

/* Just for refer:
type Resource struct {
	Class        TUClass
	ID           string
	Distribution string
	Version      string
	Arch         string
	CPU          int64
	Memory       int64
}
*/
type TSResource struct {
	libocit.Resource

	ID string
	//0 means no limit
	MaxJobs     int
	TestUnitIDs []string
}

var ResourceStore map[string]TSResource

func (res *TSResource) Valid() error {
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

func (res *TSResource) GenerateID() string {
	if val, err := json.Marshal(res); err == nil {
		res.ID = libocit.MD5(string(val))
		fmt.Println("generated id ", res.ID)
	} else {
		fmt.Println("Fatal error in generate id ", err)
	}
	return res.ID
}

func (res *TSResource) Match(unit libocit.TestUnit) bool {
	if unit.Distribution != res.Distribution {
		return false
	}
	if unit.Version != res.Version {
		return false
	}
	if unit.Arch != res.Arch {
		return false
	}

	//TODO: better calculation
	if unit.CPU > res.CPU {
		return false
	}
	if unit.Memory > res.Memory {
		return false
	}
	return true
}

func (res *TSResource) Allocate(unit libocit.TestUnit) bool {
	if res.MaxJobs == 0 || res.MaxJobs > len(res.TestUnitIDs) {
		res.TestUnitIDs = append(res.TestUnitIDs, unit.GetID())
		return true
	}
	return false
}

func TSRApply(unit libocit.TestUnit) string {
	for id, res := range ResourceStore {
		if res.Match(unit) {
			res.Allocate(unit)
			return id
		}
	}
	return ""
}

func TSRQueryList(resQuery TSResource) (ids []string) {
	fmt.Println(resQuery)
	for id, res := range ResourceStore {
		if len(resQuery.Class) > 1 {
			if resQuery.Class != res.Class {
				continue
			}
		}
		if len(resQuery.Distribution) > 1 {
			if resQuery.Distribution != res.Distribution {
				continue
			}
		}
		if len(resQuery.Version) > 1 {
			if resQuery.Version != res.Version {
				continue
			}
		}
		if len(resQuery.Arch) > 1 {
			if resQuery.Arch != res.Arch {
				continue
			}
		}
		if resQuery.CPU > res.CPU {
			log.Println("not enough CPU")
			continue
		}
		if resQuery.Memory > res.Memory {
			log.Println("not enough Memory")
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func TSRQuery(id string) (TSResource, bool) {
	fmt.Println("TSRQuery", id)
	fmt.Println(ResourceStore)
	val, ok := ResourceStore[id]
	return val, ok
}

func TSRAdd(res TSResource) bool {
	id := res.GenerateID()
	if _, ok := ResourceStore[id]; ok {
		return false
	}
	ResourceStore[id] = res
	return true
}

func TSRDelete(id string) bool {
	if _, ok := ResourceStore[id]; ok {
		delete(ResourceStore, id)
		return true
	}
	return false
}

func TSRInit() {
	ResourceStore = make(map[string]TSResource)
}

func TSRInitFromFile(url string) error {
	TSRInit()

	f, err := os.Open(url)
	if err != nil {
		return err
	}
	defer f.Close()

	var rs []TSResource
	if err = json.NewDecoder(f).Decode(&rs); err != nil {
		return err
	}

	for index := 0; index < len(rs); index++ {
		if TSRAdd(rs[index]) {
			fmt.Println(rs[index])
		}
	}
	return nil
}
