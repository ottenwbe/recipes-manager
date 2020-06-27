# go-cook

Backend Service to Manage Recipes.

### Development Dependencies

* Go >= 1.13

* For linting install ```go lint```
    ```    
    go get -u golang.org/x/lint/golint
    ```
  
* For testing install ``ginkgo``
    ```
    go get github.com/onsi/ginkgo/ginkgo
    go get github.com/onsi/gomega/...
    ```

### Notice
* base makefile from: https://github.com/vincentbernat/hellogopher

## ARM

### Docker builds 

make docker-arm 

### Tips and Tricks

* If necessary, stop all container; i.e., if they hang
    ```    
    docker stop $(docker ps -a -q)
    ```    

1. Remove all container and their volumes
    ```    
    docker rm -v $(docker ps -a -q)      
    ```

### approvals in jenkins /scriptapprovals

method hudson.plugins.git.GitSCM getBranches
method hudson.plugins.git.GitSCM getUserRemoteConfigs
method hudson.plugins.git.GitSCM isDoGenerateSubmoduleConfigurations
method hudson.plugins.git.GitSCMBackwardCompatibility getExtensions
 
 ### Disclaimer
 
 I created this project for the purpose of educating myself and personal use.
 If you are interested in the outcome, feel free to contribute; this work is published under the MIT license. 
 
### Notice
The base MAKEFILE is adapted from: https://github.com/vincentbernat/hellogopher 