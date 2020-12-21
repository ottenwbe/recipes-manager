# go-cook

[![Build Status](https://travis-ci.org/ottenwbe/go-cook.svg?branch=master)](https://travis-ci.org/ottenwbe/go-cook)

Backend service to manage recipes. 
Go-Cook supports managing the recipes via API and persistence of the recipes in a database.

## Related projects

|Tool|URL|
|---|---|
| Web Frontend  |  https://github.com/ottenwbe/go-cook-ui |

## Development 

### Dependencies

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

## Builds

A Makefile supports the build process. This includes building a development and release version of the app. Furthermore, it includes building docker images.

### Build Snapshot

```
make build 
```

### Build Release Version

```
make release
```


### Docker builds

```
make docker
```

### ARM Docker builds 

```
make docker-arm 
```

### Docker Tips and Tricks

* If necessary, stop all container; i.e., if they hang
    ```    
    docker stop $(docker ps -a -q)
    ```    

* Remove all container and their volumes
    ```    
    docker rm -v $(docker ps -a -q)      
    ```
 
## Deployment

### File-based Configuration 

## API Documentation
 
 Note: Incomplete
 
 ### Generate the Documentation 
 
The Swagger API documentation is based on [gin-swagger](https://github.com/swaggo/gin-swagger):
 
    swag init -d ./core -g http.go --parseInternal
 
 ### Disclaimer
 
 I created this project for the purpose of educating myself and personal use.
 If you are interested in the outcome, feel free to contribute; this work is published under the MIT license. 
 
### Notice
The base MAKEFILE is adapted from: https://github.com/vincentbernat/hellogopher 