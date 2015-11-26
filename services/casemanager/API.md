#Test Case Manager
The 'Test Case Manager' provides all the test cases which stored at github.

- The [configuration](#configs "configuration") file is used for the OCT users to set his/her own configuration.
- The [openAPIs](#apis "APIs") are used for the developers to integrate the 'Test Case Manager' with or their own tools(web for example).
- The [attributes](#attributes "attributes") are listed at the end of this document.

##Configs
|Key|Type|Description|Example|
|------|----|------| ----- |
| Repos | []Repo | The repos of the case, see details about [Repo](#repo) ||
| SchedulerURL | string| The url of the `Scheduler` server | "http://localhost:8011"|
| Port | int | Port of this casemanager server| default to 8011|
| Debug | bool | Print the debug information on the screen| true, default to false |

###Repo
|Key|Type|Description|
|------|----|------|
|Name| string| Name of the repo, should be unique|
|Url| string| Repo url, could be private url|
|CaseFolder| string | The 'case folder' of the repo|
|Enable| bool ||
|Groups| []string] ||

```
###casemanager.conf
{
  "Repos": [
                {
                "Name":  "huawei",
                "Url":   "github.com/huawei-openlab/oct-engine",
                "CaseFolder":  "cases/case-db",
                "Enable": false,
                "Groups": [
                        "specstest", "benchmark/monitor", "network"
                        ]
                }
        ],
  "SchedulerURL": "http://localhost:8001",
  "Port":  8011,
  "Debug": true
}
```

##APIs

|Method|Path|Summary|Description|
|------|----|------|-----------|

| GET | `/repo` | [List Repos](#list-repos) | List repos|
| GET | `/repo/:ID` | [List Repo](#list-repo) | List repo by ID|
| POST | `/repo` | [Add a repo, or Refresh repos](#add-a-repo-or-refresh-repos) | Add a repo or refresh all the enabled repos|
| POST | `/repo/:ID` | [Modify or refresh a repo](#modify-or-refresh-a-repo) | Modify or refresh a repo by ID|
| GET | `/case` | [List cases](#list-cases) | List the cases|
| GET | `/case/:ID` | [Get Case](#get-a-case) | Fetch the case files. %caseid.tar.gz |
| GET | `/case/:ID/report` | [Case Report](#case-report) | Get the case report|
| GET | `/task` | [List Tasks](#list-tasks) | List tasks|
| GET | `/task/:TaskID` | [List Task info](#list-task) | List task|
| GET | `/task/:TaskID/report` | [Get task report](#get-report) | Get task report|
| POST | `/task` | [Add Task](#add-task) | Add task based on a test case|
| POST | `/task` | [Run Task](#run-task) | Send a apply/deploy/run/collect/destroy action to run a task|

###List Repos
```
GET /repo
```
List all the repos recorded in the server.

**Parameters**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| Page | int | The page number of the test repos, sort by time. Default: 0 |
| Pagesize | int | The pagesize of the test repos. Default: 10, no more than 100 |

**Response**

```
{
        "Status": "OK",
        "Message": "2 repos founded",
        "Data": [
                {
                        "ID": "2eae7c8a4789c6b174b936e886c1753f",
                        "Name": "github.com/huawei-openlab/oct-engine",
                        "CaseFolder": "cases/case-db",
                        "Enable": false,
                        "Groups": [
                                "specstest",
                                "benchmark/monitor",
                                "network"
                        ]
                },
                {
                        "ID": "724eb5049ee0d89c2d805cfa95a3c66c",
                        "Name": "github.com/liangchenye/oct-case",
                        "CaseFolder": "casedb",
                        "Enable": true,
                        "Groups": [
                                "oci-spec"
                        ]
                }
        ]
}
```

###List Repo
```
GET /repo/repo-id
```
Get the detailed info of a repo by its ID.

**Response**
```
{
        "Status": "OK",
        "Message": "",
        "Data": {
                "ID": "724eb5049ee0d89c2d805cfa95a3c66c",
                "Name": "github.com/liangchenye/oct-case",
                "CaseFolder": "casedb",
                "Enable": true,
                "Groups": [
                        "oci-spec"
                ]
        }
}
```

###Add a repo or Refresh repos
```
POST /repo?Action=Add
```
Using 'Action=Add' to add a repo. Using 'Action=Refresh' to refresh all the enabled repos.
Fill body with repo content.

###Modify or Refresh a repo
```
POST /repo/repo-id?Action=Refresh
```
Using 'Action=Modify' to modify a repo. Using 'Action=Refresh' to refresh a repo.
Fill body with repo content in 'Modify' mode.

###List Cases
```
GET /case
```
List the idle/tested test cases.

**Parameters**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| Status |	string | "idle/tested". Default: all |
| Page | int | The page number of the test cases, sort by time. Default: 0 |
| Pagesize | int | The pagesize of the test cases. Default: 10, no more than 100 |

**Response**

```
[
  {
    "ID": "10100",
    "CaseName": "performance/network-latency",
    "Status": "tested"
  },
  {
    "ID": "10102",
    "CaseName": "function/fake-support",
    "Status": "idle"
  }
]
```

###Get a case

```
GET /case/10100
```
Fetch the case files. 

**Response**

The whole %caseid.tar.gz file.


###Case Report

```
GET /case/10100/report
```
Fetch the case report. 

**Response**

The whole %caseid-report file.


###List Tasks
```
GET /task
```
List the running tasks.

**Parameters**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| Page | int | The page number of the tasks, sort by time. Default: 0 |
| Pagesize | int | The pagesize of the tasks. Default: 30.|

**Response**

```
        "Status": "ok",
        "Message": "1 task founded",
        "Data": [
                {
                        "ID": "8c0bc6d41ff780c05131a2d98a1c00a9",
                        "PostURL": "http://localhost:8001",
                        "BundleURL": "/tmp/.testcase_cache/github.com/liangchenye/oct-case/casedb/oci-spec/bundle.tar.gz",
                        "Status": "init",
                        "Priority": 100,
                        "SchedulerID": ""
                }
        ]
```

###List Task
```
GET /task/:TaskID
```
Show the status of a single task.

**Response**

```
	"Status": "ok",
        "Message": "",
        "Data": {
                "ID": "047a07950446e24c1aa0c6324abf6770",
                "PostURL": "http://localhost:8001",
                "BundleURL": "/tmp/.testcase_cache/github.com/liangchenye/oct-case/casedb/oci-spec/bundle.tar.gz",
                "Status": "init",
                "Priority": 100,
                "SchedulerID": ""
        }
```

###Get Report
```
GET /task/:TaskID/report > report.tar.gz
```
Get the task report and redirect to a 'tar.gz' file.

###Add Task
```
POST {caseID} /Task
```
Turn a test case into a task.

**Example**
```
curl -d caseID localhost:8011/task
```

**Response**
```
{
        "Status": "ok",
        "Message": "df43a849ceb93c8235f017f25a95dcf5",
        "Data": null
}
NOTE: The content of 'Message' is the task ID.
```

###Run Task
```
POST {action} /Task/:TaskID
```

**Example**
```
curl -d run localhost:8011/task/:taskid
```

**Response**
```
{
        "Status": "ok",
        "Message": "",
        "Data": null
}
```

**ID**
`The string ID of the test case.`

**CaseName**
`The name of the test case, same with the API URL for fetching the event.`

**Status**
In `List` : `'idle/tested'`

In `Refresh` : `'OK/Failed'`



