## Local Development
### Requirements
* Golang version higher than 1.13;

>_**NOTE**: The GOPATH and GOROOT environment variables should be added in PATH._
>```
>export GOPATH=C:\Users\<<username>>\go
>export GOROOT=C:\Go
>```

* Configured access to the VCS, for details, refer to the [Gerrit Setup for Developer](https://kb.epam.com/display/EPMDEDP/Gerrit+Setup+for+Developer) page;
* GoLand IntelliJ IDEA or another IDE.

### Operator Launch
In order to run the operator, follow the steps below:

1. Clone repository;
2. Open folder in GoLand Intellij IDEA, click the ![add_config_button](../readme-resource/add_config_button.png "add_config_button") button and select the **Go Build** option:
![add_configuration](../readme-resource/add_configuration.png "add_configuration") 
3. In Configuration tab, fill in the following:

    3.1. In the Field field, indicate the path to the main.go file;
    
    3.2. In the Working directory field, indicate the path to the operator;
    
    3.3. In the Environment field, specify the platform name (OpenShift/Kubernetes);
   ![build-config](../readme-resource/build_config.png "build-config") 

4. Run 'go build main.go' (Shift+F10);
6. Check the operator logs in output console: 
```
{"level":"info","ts":1580820956.2087429,"logger":"cmd","msg":"Go Version: go1.13.7"}
{"level":"info","ts":1580820956.2087429,"logger":"cmd","msg":"Go OS/Arch: windows/amd64"}
{"level":"info","ts":1580820956.2087429,"logger":"cmd","msg":"Version of operator-sdk: v0.8.0+git"}
{"level":"info","ts":1580820956.2087429,"logger":"cmd","msg":""}
{"level":"info","ts":1580820959.592205,"logger":"cmd","msg":"Registering Components."}
{"level":"info","ts":1580820959.592205,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"codebase-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1580820959.592205,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"codebasebranch-controller","source":"kind source: /, Kind="}
2020/02/04 14:55:59 Openshift clients was successfully created
{"level":"info","ts":1580820959.608795,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"gitserver-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1580820959.608795,"logger":"cmd","msg":"Starting the Cmd."}
```

### Exceptional Cases

##### CASE 1

After starting the Go build process, the following error will appear: 
```
go: finding github.com/openshift/api v3.9.0
go: finding github.com/openshift/client-go v3.9.0
go: errors parsing go.mod:
C:\Users\<<username>>\Desktop\EDP\edp-admin-console\go.mod:36: require github.com/openshift/api: version "v3.9.0" invalid: unknown revision v3.9.0

Compilation finished with exit code 1
```

To resolve the issue, update the go dependency by applying the Golang command:
```
go get github.com/openshift/api@v3.9.0
```

##### CASE 2

If you get the **Unauthorized** error in logs, check the current **KUBECONFIG** path. Before running the operator in the Go build configuration (see point 4), you can specify another **KUBECONFIG** path:

```
WATCH_NAMESPACE=test-deploy-platform;PLATFORM_TYPE=openshift;KUBECONFIG=C:\Users\<<username>>\.kube\kubeconfig
```