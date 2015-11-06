package libocit

import (
	"fmt"
	"time"
)

type DBCollectName string

const (
	DBCase DBCollectName = "case"
	DBRepo               = "repo"
	DBTask               = "task"
)

type DBInterface interface {
	String() string
}

type DBQuery struct {
	Page     int
	PageSize int
	Params   map[string]string
}

var OCTDB map[DBCollectName](map[string]DBInterface)

func DBGenerateID(val string) string {
	return MD5(fmt.Sprintf("%d-%s", time.Now().Unix(), val))
}

func DBCollectExist(collect DBCollectName) bool {
	if OCTDB == nil {
		return false
	}
	_, ok := OCTDB[collect]
	return ok
}

func DBRegist(collect DBCollectName) bool {
	if OCTDB == nil {
		OCTDB = make(map[DBCollectName](map[string]DBInterface))
	}
	_, ok := OCTDB[collect]
	if ok {
		//fmt.Println("The collect is already registed, donnot do it twice")
		return false
	} else {
		OCTDB[collect] = make(map[string]DBInterface)
	}
	return true
}

func DBGet(collect DBCollectName, id string) (DBInterface, bool) {
	if !DBCollectExist(collect) {
		return nil, false
	}

	val, ok := OCTDB[collect][id]
	return val, ok
}

func DBAdd(collect DBCollectName, val DBInterface) (string, bool) {
	if !DBCollectExist(collect) {
		return "", false
	}

	id := DBGenerateID(val.String())

	switch collect {
	case DBCase:
		tc, _ := CaseFromString(val.String())
		tc.SetID(id)
		OCTDB[collect][id] = tc
	case DBRepo:
		repo, _ := RepoFromString(val.String())
		repo.SetID(id)
		OCTDB[collect][id] = repo
	case DBTask:
		task, _ := TestTaskFromString(val.String())
		task.SetID(id)
		OCTDB[collect][id] = task
	}

	return id, true
}

func DBModify(collect DBCollectName, id string, val DBInterface) bool {
	if !DBCollectExist(collect) {
		return false
	}
	_, ok := OCTDB[collect][id]
	if ok {
		OCTDB[collect][id] = val
	}

	return ok
}

func DBRemove(collect DBCollectName, id string) bool {
	if !DBCollectExist(collect) {
		return false
	}
	delete(OCTDB[collect], id)
	return true
}

func DBLookup(collect DBCollectName, query DBQuery) (ids []string) {
	switch collect {
	case DBCase:
	case DBRepo:
	case DBTask:
	}
	return ids
}
