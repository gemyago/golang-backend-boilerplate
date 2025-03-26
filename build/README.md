# Build

This folder contains the build tools for the project.

## Build Binaries

Golang binaries are build for platforms defined in [build.cfg](build.cfg) file (see `platforms` section).

## Docker

To enable multi-platform builds please enable [container image storage](https://docs.docker.com/build/building/multi-platform/#prerequisites) for your docker daemon.

## Build Scripts

The build scripts are located in the [scripts](scripts) folder.

If iterating on scripts, please make sure to run the tests:

```sh
# Run tests for all scripts
make test

# Run specific python tests
python -m unittest discover -v -s ./scripts/tests -k TestListVersions
```