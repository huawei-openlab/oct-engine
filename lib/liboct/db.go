package liboct

import (
	"errors"
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

//TODO: change DBGenerateID to something like DBStore.Generate

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

func DBCollectExist(collect DBCollectName) error {
	if OCTDB == nil {
		return errors.New("The Database is not initialized.")
	}
	if _, ok := OCTDB[collect]; ok {
		return nil
	}

	return errors.New(fmt.Sprintf("The collect %s is not exist.", collect))
}

func DBRegist(collect DBCollectName) error {
	if OCTDB == nil {
		OCTDB = make(map[DBCollectName](map[string]DBInterface))
	}
	_, ok := OCTDB[collect]
	if ok {
		return errors.New("The collect is already registed, donnot do it twice.")
	} else {
		OCTDB[collect] = make(map[string]DBInterface)
	}
	return nil
}

func DBGet(collect DBCollectName, id string) (DBInterface, error) {
	if err := DBCollectExist(collect); err != nil {
		return nil, err
	}

	if val, ok := OCTDB[collect][id]; ok {
		return val, nil
	}
	return nil, errors.New(fmt.Sprintf("Cannot find %s in the collect %s", id, collect))
}

func DBAdd(collect DBCollectName, val DBInterface) (string, error) {
	if err := DBCollectExist(collect); err != nil {
		return "", err
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
	case DBScheduler:
		s, _ := SchedulerFromString(val.String())
		s.SetID(id)
		OCTDB[collect][id] = s
	}

	return id, nil
}

func DBUpdate(collect DBCollectName, id string, val DBInterface) error {
	if err := DBCollectExist(collect); err != nil {
		return err
	}
	if _, ok := OCTDB[collect][id]; ok {
		OCTDB[collect][id] = val
		return nil
	}

	return errors.New(fmt.Sprintf("Cannot find the %s in collect %s.", id, collect))
}

func DBRemove(collect DBCollectName, id string) error {
	if err := DBCollectExist(collect); err != nil {
		return err
	}
	delete(OCTDB[collect], id)
	return nil
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
