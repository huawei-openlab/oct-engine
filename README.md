# Test engine for Open Container

The oct(open container testing) project aims to promote the [Open Container Initiative](http://www.opencontainers.org/) by providing a universal testing framework for all the container projects. The testing includes: [oci specs](https://github.com/opencontainers/specs), container function and container performance.

The oct-engine provides the framework for the open container testing.

## The framework architecture
![Framework](docs/static/test_framework.png "Framework")
  * `OCTD` :  
    OCTD acts as a container pool or test sever, up to the configuration file.
    When configurated as 'container pool', it uses 'Harbour' to run the testing inside containers.
    When configurated as a 'test server', it simplely run the testing.

  * `Testing Scheduler` :  
    As the main scheduler, the Test Case Scheduler will:
    1. Parse the testing request
    2. Apply hardware resources from the Open Test Server
    3. Register container images from the Open Container Pool
    4. Deploy the testing environment
    5. Run the test
    6. Collect and publish the testing report
   
  * `Case Manager` :  
    The Case Manager provides RESTful API for user to list/get the test cases, and send testing request to the `Scheduler`.
    It uses the github as the static test case database.
    
## TestCase specification
Refers to [cases/spec](cases/spec.md) 

## Who should join
- Container project developer
- Operating system distributor
- Hardware company
- IASS provider
- PASS provider
- Any container user


### How to involve the oct-service
- Fork the repository on GitHub
- Read the [APIs](services/API.md) to test each service
- Play with the project, submit bugs, submit patches!

If any issues are encountered while using the oct-engine project, several avenues are available for support:
<table>
<tr>
	<th align="left">
	Issue Tracker
	</th>
	<td>
	https://github.com/huawei-openlab/oct-engine/issues
	</td>
</tr>
<tr>
	<th align="left">
	Google Groups
	</th>
	<td>
	https://groups.google.com/forum/#!forum/oci-testing
	</td>
</tr>
</table>


### [Getting Started with oct-service](HowTO.md)
OCT-engine support multiple servers, `testserver` will manage all the end-node machines
with `octd` deployed.
But We can run `testserver` and `octd` on a single machine with the default configuration (ip/port)
to get a first impression of the OCT-engine.

Also, we provide a scheduler client to test single case, as a demo, you can try this:
```
## Prepare 
git clone https://github.com/huawei-openlab/oct-engine.git
cd oct-engine/services
make
cd scheduler
./scheduler &
cd ../octd 
./octd &
cd ../scheduler/client
make
./client --path=./bundle.tar.gz
```

### Quick look of the testing report

The [testing report](cases/case-report) will keep updating.


