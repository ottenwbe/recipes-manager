# go-cook

[![Build Status](https://travis-ci.org/ottenwbe/go-cook.svg?branch=master)](https://travis-ci.org/ottenwbe/go-cook)

Backend service to manage recipes. 
Go-cook supports managing the recipes via API and persistence of the recipes in a database.

## Related projects

|Tool|URL|
|---|---|
| Web Frontend  |  https://github.com/ottenwbe/go-cook-ui |
| Deployment    |  https://github.com/ottenwbe/go-cook-deployment |

## How to Use?

### Deployment

The two main options to run the app are either Kubernetes-based or as a standalone service:

1. See https://github.com/ottenwbe/go-cook-deployment how to run the whole suite of micro-services on a Kubernetes cluster, including the frontend and database.

2. To run go-cook as standalone service (either amd64 or arm64). Assumption is that a MongoDB is already running:

    1. Prepare a configuraiton file (see next section).
    1. Run the container
        
            docker run -p 8080:8080 -v <local-config>:/etc/go-cook/go-cook-config.yml ottenwbe/go-cook:0.1.0-amd64
    
    1. Check if everything is running:

            curl localhost:8080/api/v1/recipes

    1. Details about the API can be checked in a browser:

            localhost:8080/swagger/swagger_index.html            

### File-based Configuration 

Configuraiton files are expected at ```~/.go-cook/go-cook-config.yml``` or ```/etc/go-cook/go-cook-config.yml```.

```yaml
# Mandatory Configuration
recipeDB:
  host: <db host>

# Optional Configuration
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
    ingredients: <name of ingredients section in the drive file>
    instructions: <name of the instructions section in the drive file>

source:
  host: <source host, i.e., aka host of ui>
```

#### Configuration with Environment Variables

By prepending all variables (see file-based configuration) with ```GO_COOK_``` the configuration can be set in the environment.


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

## API Documentation

 ### Generate the Documentation 
 
The Swagger API documentation is based on [gin-swagger](https://github.com/swaggo/gin-swagger):
 
    swag init --exclude vendor
 
 ### Disclaimer
 
 I created this project for the purpose of educating myself and personal use.
 If you are interested in the outcome, feel free to contribute; this work is published under the MIT license. 
 
### Notice
The base MAKEFILE is adapted from: https://github.com/vincentbernat/hellogopher 