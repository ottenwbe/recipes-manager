#!/usr/bin/env bash

set -uex pipefail

echo "Prepare tests..."

# remove existing test container
# or ignore the error if the container does not exist
docker stop test-db | true
docker rm -v test-db | true
# run a mongo-db
docker run -d --name=test-db -p 27018:27017 mongo:7

mkdir -p ~/.recipes-manager

{
    echo "recipeDB:"
    echo "  host: mongodb://localhost:27018"
} > ~/.recipes-manager/recipes-manager-config.yml

echo "Testing..."

ginkgo -v -cover ./...

echo "Collecting results"

# delete outdated result iff they exist
rm -rf test/results
rm -rf test/coverage

mkdir -p test/results
mkdir -p test/coverage

# remove existing test container
# or ignore the error if the container does not exist
#docker stop test-db | true
#docker rm -v test-db | true

echo "Cleanup..."
