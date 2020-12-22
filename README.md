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

## Building Go-Cook

A Makefile supports the build process. This includes building a development and release version of the go-cook service. Furthermore, it includes building docker images to easily deploy the go-cook service.

### Build Snapshot

```
make build 
```

Builds a fully functioning binary named ```go-cook-snapshot```. In contrast to the release version, there is still debugging informaiton included.

### Build Release Version

```
make release
```

Builds a fully functioning binary named ```go-cook```. 

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
 
## Configuration

### File-based Configuration 

Configuraiton files are expected at ```~/.go-cook/go-cook-config.yml```.

```yaml
recipeDB:
  host: <db host>

html:
  address: <server listens on this address>
  cors:
    origin: <Access-Control-Allow-Origin>

drive:
  connection:
    secret:
      file: <location of secret>
  recipes:
    folder: <folder name in drive>

sourceClient:
  host: <source host>
```


### Environment-based Configuration

By prepending all variables (see file-based configuration) with ```GO_COOK_``` the configuration can be set in the environment.

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