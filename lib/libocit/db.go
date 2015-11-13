package libocit

import (
	"fmt"
	"time"
)

type DBCollectName string

const (
	DBCase     DBCollectName = "case"
	DBRepo                   = "repo"
	DBResource               = "resource"
	//testing task: for frontend users
	DBTask = "task"
	//scheduler: for backend servers
	DBScheduler = "scheduler"
	//scheduler unit
	DBUnit = "unit"
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

//The case, repo, resource should be consistent
//The task could always be different
func DBGenerateID(collect DBCollectName, val string) string {
	switch collect {
	case DBCase:
		tc, _ := CaseFromString(val)
		return MD5(tc.GetBundleContent())
	case DBRepo:
		repo, _ := RepoFromString(val)
		return MD5(fmt.Sprintf("%s%s", repo.URL, repo.CaseFolder))
	case DBTask:
	}
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

	id := DBGenerateID(collect, val.String())

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
		task, _ := TaskFromString(val.String())
		task.SetID(id)
		OCTDB[collect][id] = task
	case DBUnit:
		unit, _ := UnitFromString(val.String())
		unit.SetID(id)
		OCTDB[collect][id] = unit
	case DBResource:
		res, _ := ResourceFromString(val.String())
		res.SetID(id)
		OCTDB[collect][id] = res
	}

	return id, true
}

func DBUpdate(collect DBCollectName, id string, val DBInterface) bool {
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
	noLimit := false
	if query.Page < 0 {
		query.Page = 0
	}
	if query.PageSize < 0 {
		query.PageSize = 30
	} else if query.PageSize == 0 {
		noLimit = true
	}
	i := 0
	for id, _ := range OCTDB[collect] {
		if noLimit == true || (i >= query.Page*query.PageSize && i < (query.Page+1)*query.PageSize) {
			//TODO: check by 'reflect'
			switch collect {
			case DBCase:
			case DBRepo:
			case DBTask:
			}
			ids = append(ids, id)
		}
		i++
	}
	return ids
}
