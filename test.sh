#!/usr/bin/env bash

set -uex pipefail

echo "Prepare tests..."

# remove existing test container
# or ignore the error if the container does not exist
docker stop test-mongo | true
docker rm test-mongo | true
# run a mongo-db
docker run -d --name=test-mongo -p 27017:27017 mongo:3

mkdir -p ~/.go-cook

{
    echo "recipeDB:"
    echo "  host: mongodb://localhost:27017"
} > ~/.go-cook/go-cook-config.yml

echo "Testing..."

ginkgo -v -cover ./...

echo "Collecting results"

# delete outdated result iff they exist
rm -rf test/results
rm -rf test/coverage

mkdir -p test/results
mkdir -p test/coverage

for d in $(go list -f '{{.Dir}}' ./...); do
    mv $d/*junit.xml test/results
    mv $d/*coverprofile test/coverage
done

echo "Cleanup..."
