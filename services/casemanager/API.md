#Test Case Server
The 'Test Case Server' provides all the test cases which stored at github.

- The [configuration](#configs "configuration") file is used for the OCT users to set his/her own configuration.
- The [openAPIs](#apis "APIs") are used for the developers to integrate the 'Test Case Server' with the 'Scheduler'.
- The [attributes](#attributes "attributes") are listed at the end of this document.

##Configs
|Key|Type|Description|Example|
|------|----|------| ----- |
| GitRepo | string | The git repo url.| "https://github.com/huawei-openlab/oct.git" |
| CaseDir | string | The case dir of the git repo.| "cases" |
| Group | []string | The group names, also means the sub directory| ["Performance"]|
| CacheDir | string | The cache dir where the repo stored.| "/tmp/tcserver_cache" |
| Debug | bool | Print the debug information on the screen| true, default to false |

```
{
	"GitRepo":  "https://github.com/huawei-openlab/oct.git",
        "CaseDir":  "Cases",
	"Group": ["Network", "Benchmark", "Specstest"],
	"CacheDir": "/tmp/tcserver_cache",
	"Port":  8011
}
```


##APIs

|Method|Path|Summary|Description|
|------|----|------|-----------|

| GET | `/repo` | [List Repos](#list-repos) | List repos. |
| GET | `/repo/:ID` | [List Repo](#list-repo) | List repo by ID. |
| POST | `/repo` | [Add a repo, or Refresh repos](#add-a-repo-or-refresh-repos) | Add a repo or refresh all the enabled repos. |
| POST | `/repo/:ID` | [Modify or refresh a repo](#modify-or-refresh-a-repo) | Modify or refresh a repo by ID. |
| GET | `/case` | [List cases](#list-cases) | List the cases. |
| GET | `/case/:ID` | [Get Case](#get-a-case) | Fetch the case files. %caseid.tar.gz |
| GET | `/case/:ID/report` | [Case Report](#case-report) | Get the case report|

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

##Attributes

**ID**
`The string ID of the test case.`

**CaseName**
`The name of the test case, same with the API URL for fetching the event.`

**Status**
In `List` : `'idle/tested'`

In `Refresh` : `'OK/Failed'`



