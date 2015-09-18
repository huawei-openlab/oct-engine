#OCT Case Specifications
Before using/reading/writing an OCT case, please read the [Specs](spec.md).

#Cases in github
The 'group' is defined to make it easier to manage cases.
For example, now we have groups like `network`, `specs`. 
Under each group are the named [case bundles](spec.md#case-bundle).

#How to submit your case
- Name your case bundle, choose the right group and submit your bundle there. If your cannot find a suitable 'group', please fire an issue and create a new group.
- We provide a 'casevalidator' for case writers, Read the [HowTO](../engine/tcserver/casevalidator/HowTO.md) to check if your case was valid before submit your cases.

The cases in this oct repo are good to use, if you want to try any of them, please read the [Test Case Server](../engine/tcserver/README.md).

#Implementation
###Name the bundle
  In engine/tcserver, there are several test cases which are well named, it used to better discovery and transfer cases.
  Of cource, the 'Description' session in config.json has more information.
###Tar the bundle
  In oct-engine services (tcserver, ocitd, testserver, schedular), tar.gz is used as the default format to tar a case bundle.
  But if a case bundle has no scripts (which is perfect case!), the config.json will also be recognized as a complete case.
###Transfer the bundle
  In oct-engine services, the bundle is parsed and transfered by RESTful API.
Of cource, other implementation are welcomed.
