 # recipes-manager

[![CI](https://github.com/ottenwbe/recipes-manager/actions/workflows/go.yml/badge.svg)](https://github.com/ottenwbe/recipes-manager/actions/workflows/go.yml)

Backend service to manage recipes. 
Recipes-manager supports managing recipes via API while persisting them in a database.

## Related projects

|Tool|URL|
|---|---|
| Web Frontend  |  https://github.com/ottenwbe/recipes-manager-ui |
| Deployment    |  https://github.com/ottenwbe/recipes-manager-deployment |

## How to Use?

### Deployment

We briefly describe two deployment options. First, a Kubernetes-based deployment, including the deployment of all other cooking-related tools. Secondly, a stand-alone deployment:

1. See https://github.com/ottenwbe/recipes-manager-deployment for how to run the whole suite of micro-services on a Kubernetes cluster, including the frontend and a database.

2. To run recipes-manager as standalone service (either amd64 or arm64): 

    1. Prepare a configuration file (see next section).
    1. Start a database

            docker run -d --name=db-recipes-manager -p 27018:27017 mongo:6

    1. Run the container
        
            docker run -p 8080:8080 --name=backend-recipes-manager -v <local-config>:/etc/recipes-manager/recipes-manager-config.yml ottenwbe/recipes-manager:v0.3.0
    
    1. Check if everything is running:

            curl localhost:8080/api/v1/recipes

    1. Details about the API can then be checked in a browser:

            localhost:8080/swagger/swagger_index.html            

### File-based Configuration 

Configuration files are expected at ```~/.recipes-manager/recipes-manager-config.yml``` or ```/etc/recipes-manager/recipes-manager-config.yml```.

```yaml
# Mandatory Configuration
recipeDB:
  host: <db host>

# Optional Configuration
html:
  address: <server listens on this address>
  cors:
    origin: <Access-Control-Allow-Origin>

drive: # To fetch recipes from Goolge Drive
  connection:
    secret:
      file: <location of secret>
  recipes:
    folder: <folder name in drive>
    ingredients: <name of ingredients section in the drive files>
    instructions: <name of the instructions section in the drive files>

source:
  host: <source host, i.e., aka host of ui>
```

#### Configuration with Environment Variables

By prepending all variables (see file-based configuration) with ```GO_COOK_``` the configuration can be set in the environment.

## Development 

### Dependencies

* Go >= 1.19
  
* For testing install ``ginkgo``
    ```
    go install github.com/onsi/ginkgo/v2/ginkgo
    go get github.com/onsi/gomega/...
    ```

## Building recipes-manager

A Makefile supports the build process. This includes building a development and release version of the recipes-manager service. Furthermore, it includes building docker images to easily deploy the recipes-manager service.

### Build Snapshot

```
make snapshot 
```

Builds a fully functioning binary named ```recipes-manager-snapshot```. In contrast to the release version, there is still debugging informaiton included.

### Build Release Version

```
make release
```

Builds a fully functioning binary named ```recipes-manager```. 

### Docker builds

```
make docker
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