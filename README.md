 # recipes-manager

[![CI](https://github.com/ottenwbe/recipes-manager/actions/workflows/go.yml/badge.svg)](https://github.com/ottenwbe/recipes-manager/actions/workflows/go.yml)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/ottenwbe/recipes-manager/blob/master/LICENSE)
[![Known Vulnerabilities](https://snyk.io/test/github/ottenwbe/recipes-manager/badge.svg)](https://snyk.io/test/github/ottenwbe/recipes-manager)

Backend service to manage recipes. 
Recipes-manager supports managing recipes via API while persisting them in a database.

## Related projects

| Tool         |URL|
|--------------|---|
| Web Frontend |  https://github.com/ottenwbe/recipes-manager-ui |
| Recommender  |  https://github.com/ottenwbe/recipes-manager-recommender |
| Deployment   |  https://github.com/ottenwbe/recipes-manager-deployment |

## How to Use?

### Deployment

We briefly describe two deployment options. First, a Kubernetes-based deployment, including the deployment of all other cooking-related tools. Secondly, a stand-alone deployment:

1. See https://github.com/ottenwbe/recipes-manager-deployment for how to run the whole suite of micro-services on a Kubernetes cluster, including the frontend and a database.

2. To run recipes-manager as standalone service (either amd64 or arm64): 

    1. Prepare a configuration file (see next section).
    1. Start a database

            docker run -d --name=db-recipes-manager -p 27018:27017 mongo:7

    1. Run the container
        
            docker run -p 8080:8080 --name=backend-recipes-manager -v <local-config>:/etc/recipes-manager/recipes-manager-config.yml ottenwbe/recipes-manager:latest
    
    1. Check if everything is running:

            curl localhost:8080/api/v1/recipes

    1. Details about the API can then be checked in a browser:

            localhost:8080/swagger/index.html

## Configuration

Configuration files are expected at `~/.recipes-manager/recipes-manager-config.yml`, `/etc/recipes-manager/recipes-manager-config.yml`, or in the current directory.

### Configuration File Format (YAML)

Recipes-manager uses YAML-based configuration. Below is a comprehensive reference of all available parameters with their defaults and descriptions:

```yaml
# ==========================================
# DATABASE CONFIGURATION (Required)
# ==========================================
recipeDB:
  host: "mongodb://127.0.0.1:27017"  # MongoDB connection string

# ==========================================
# SERVER CONFIGURATION (Optional)
# ==========================================
html:
  address: ":8080"                    # Server listening address and port
  cors:
    origin: "*"                       # CORS origin header (Access-Control-Allow-Origin)

# ==========================================
# LOGGING CONFIGURATION (Optional)
# ==========================================
log:
  level: "info"                       # Log level: debug, info, warning, error

# ==========================================
# SOURCE/OAUTH CONFIGURATION (Optional)
# ==========================================
source:
  redirect: "http://localhost:8080/#!/src"  # OAuth redirect URL for sources

# ==========================================
# GOOGLE DRIVE INTEGRATION (Optional)
# ==========================================
drive:
  enabled: false                      # Enable/disable Google Drive integration
  connection:
    secret:
      file: ""                        # Path to Google Drive API secret/credentials file
  recipes:
    folder: ""                        # Name of folder in Google Drive containing recipes
    ingredients: "Ingredients"        # Section header for ingredients in Drive documents
    instructions: "Instructions"      # Section header for instructions in Drive documents

# ==========================================
# KEYCLOAK AUTHENTICATION (Optional)
# ==========================================
keycloak:
  enabled: false                      # Enable/disable Keycloak authentication
  address: ""                         # Keycloak server address
  clientID: ""                        # OAuth2 Client ID
  clientSecret: ""                    # OAuth2 Client Secret
  host: ""                            # Keycloak realm host name
```

### Configuration Parameters Reference

| Parameter | Type | Default | Required | Description |
|-----------|------|---------|----------|-------------|
| `recipeDB.host` | string | `mongodb://127.0.0.1:27017` | ✓ Yes | MongoDB connection string |
| `html.address` | string | `:8080` | No | Server listening address (host:port) |
| `html.cors.origin` | string | `*` | No | CORS origin header for browser requests |
| `accountDB.name` | string | `accounts` | No | MongoDB database name for accounts |
| `accountDB.collection` | string | `accounts` | No | MongoDB collection name for accounts |
| `log.level` | string | `info` | No | Logging level: debug, info, warning, error |
| `source.redirect` | string | `http://localhost:8080/#!/src` | No | OAuth redirect URL after source authentication |
| `drive.enabled` | boolean | `false` | No | Enable Google Drive recipe import |
| `drive.connection.secret.file` | string | - | Conditional* | Path to Google Drive credentials JSON file |
| `drive.recipes.folder` | string | - | Conditional* | Google Drive folder name containing recipes |
| `drive.recipes.ingredients` | string | `Ingredients` | No | Section header for ingredients in Drive files |
| `drive.recipes.instructions` | string | `Instructions` | No | Section header for instructions in Drive files |
| `keycloak.enabled` | boolean | `false` | No | Enable Keycloak authentication |
| `keycloak.address` | string | - | Conditional* | Keycloak server address |
| `keycloak.clientID` | string | - | Conditional* | OAuth2 Client ID |
| `keycloak.clientSecret` | string | - | Conditional* | OAuth2 Client Secret |
| `keycloak.host` | string | - | Conditional* | Keycloak realm host |

*Conditional: Required only if the feature is enabled (e.g., if `drive.enabled: true`)

### Environment Variables

All configuration parameters can be set via environment variables by:
1. Prefixing with `RECIPES_MANAGER_`
2. Replacing dots (`.`) with underscores (`_`)
3. Using uppercase letters

**Examples:**
- `recipeDB.host` → `RECIPES_MANAGER_RECIPEDB_HOST`
- `html.address` → `RECIPES_MANAGER_HTML_ADDRESS`
- `keycloak.clientSecret` → `RECIPES_MANAGER_KEYCLOAK_CLIENTSECRET`
- `drive.recipes.folder` → `RECIPES_MANAGER_DRIVE_RECIPES_FOLDER`

**Example with Docker:**
```bash
docker run -p 8080:8080 \
  -e RECIPES_MANAGER_RECIPEDB_HOST="mongodb://my-mongo:27017" \
  -e RECIPES_MANAGER_HTML_ADDRESS=":8080" \
  -e RECIPES_MANAGER_LOG_LEVEL="debug" \
  ottenwbe/recipes-manager:latest
```

### Configuration Examples

#### Minimal Configuration (Database Only)
```yaml
recipeDB:
  host: "mongodb://my-mongo-server:27017"
```

#### Development Configuration with Logging
```yaml
recipeDB:
  host: "mongodb://localhost:27017"
html:
  address: ":8080"
  cors:
    origin: "http://localhost:3000"
log:
  level: "debug"
```

#### Production Configuration with Authentication
```yaml
recipeDB:
  host: "mongodb://mongo-prod:27017/recipes"
html:
  address: ":8080"
  cors:
    origin: "https://recipes.example.com"
log:
  level: "info"
keycloak:
  enabled: true
  address: "https://auth.example.com"
  clientID: "recipes-manager"
  clientSecret: "secret-key-here"
  host: "auth.example.com/auth/realms/recipes"
```

#### Configuration with Google Drive Integration
```yaml
recipeDB:
  host: "mongodb://localhost:27017"
drive:
  enabled: true
  connection:
    secret:
      file: "/etc/recipes-manager/google-drive-secret.json"
  recipes:
    folder: "My Recipes"
    ingredients: "Zutaten"              # German example
    instructions: "Anleitung"
```

### Configuration File Discovery Order

Recipes-manager searches for configuration files in the following order:
1. `/etc/recipes-manager/recipes-manager-config.yml`
2. `$HOME/.recipes-manager/recipes-manager-config.yml`
3. `./recipes-manager-config.yml` (current directory)

The first file found is used. Alternatively, set `RECIPES_MANAGER_*` environment variables to override file-based configuration.

## Development 

### Dependencies

* Go >= 1.25
  
* For testing install ``ginkgo``
    ```
    go install github.com/onsi/ginkgo/v2/ginkgo
    go get github.com/onsi/gomega/...
    ```

## Building recipes-manager

A Makefile supports the build process. This includes building a development and release version of the recipes-manager service, running tests, and building docker images to easily deploy the recipes-manager service.

### Quality Assurance

```
make verify
```

Runs all QA steps including dependency verification, linting, and tests.

```
make test
```

Runs the test suite.

```
make smoke-test
```

Runs the end-to-end smoke test using Podman to verify the containerized application and database connectivity.

### Build Snapshot

```
make snapshot 
```

Builds a fully functioning binary named ```recipes-manager-snapshot```. In contrast to the release version, there is still debugging information included.

### Build Release Version

```
make release
```

Builds a fully functioning binary named ```recipes-manager-<VERSION>``` where VERSION is determined from git tags. 

### Docker builds

```
make docker
```

Builds and tags a Docker image with the current version.

```
make docker-dev
```

Builds a development Docker image with debugging enabled.

### Docker Tips and Tricks

* If necessary, stop all containers; i.e., if they hang
    ```    
    docker stop $(docker ps -a -q)
    ```    

* Remove all containers and their volumes
    ```    
    docker rm -v $(docker ps -a -q)      
    ``` 

## API Documentation

### Generate the Documentation 
 
The Swagger API documentation is based on [gin-swagger](https://github.com/swaggo/gin-swagger):
 
```
swag init --exclude vendor
```

Alternatively, use:

```
make api-docu
```

## Disclaimer

I created this project for the purpose of educating myself and personal use.
If you are interested in the outcome, feel free to contribute; this work is published under the MIT license.

Later changes and improvements have been supported by opencode and other AI Agents.

## Notice

The base Makefile is adapted from: https://github.com/vincentbernat/hellogopher 