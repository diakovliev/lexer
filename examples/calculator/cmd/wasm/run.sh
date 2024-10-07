#!/bin/sh
docker build -t gotest .
docker kill gotest
docker run --name gotest --rm -d -p 8080:80 gotest
