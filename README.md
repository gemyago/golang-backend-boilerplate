# golang-boilerplate

This project includes a boilerplate structure for a golang project. It includes service and functions and is targetting to be deployed on GCP.

## Dev Env Setup

This is a golang module. In order to contribute please make sure you have golang installed.
Simplest is to use [gvm](https://github.com/moovweb/gvm). Additionally [direnv](https://github.com/direnv/direnv) is used to manage shell environment.

Run `$(grep "^go " go.mod | awk '{print $2}')` to check a version of golang used.

Install required version of golang:

`gvm install $(grep "^go " go.mod | awk '{print $2}')`

Have deps installed:
```
make deps
```

When contributing make sure to push code with green tests: `make test`