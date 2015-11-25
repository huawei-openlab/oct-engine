If you just want to try the framework.
The simplest way is to deploy all the micro-services on a simple system and run the testcase.

(The default configure file is already set all the IPs to the local IP address.)

Following these steps:


```
## Prepare
git clone https://github.com/huawei-openlab/oct-engine.git
cd oct-engine/services
make
cd scheduler
./scheduler &
cd ../octd
./octd &
cd ../casemanager
./casemanager &

## Choose a testcase
1. curl localhost:8011/case
   {
        "Status": "ok",
        "Message": "1 cases founded",
        "Data": [
                {
                        "ID": "71e80571dbe09a0fb166949000
                        ...
                }
   }

## Start a task bases on a case
2. curl  -d  71e80571dbe09a0fb16694900095e429  localhost:8011/task
   {
        "Status": "ok",
        "Message": "047a07950446e24c1aa0c6324abf6770",
        "Data": null
   }

## Run the task
3. curl -d apply localhost:8011/task/047a07950446e24c1aa0c6324abf6770
   curl -d deploy localhost:8011/task/047a07950446e24c1aa0c6324abf6770
   curl -d run localhost:8011/task/047a07950446e24c1aa0c6324abf6770

## Get the report
4. curl localhost:8011/task/047a07950446e24c1aa0c6324abf6770/report > report.tar.gz

```
