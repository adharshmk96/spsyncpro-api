# go_starter API

## Prerequisites

- Go `brew install go`
- mockery `brew install mockery`
- swaggo 

## Run project

- Run `task run` to run the project

## Structure

- pkg/domain - contains all the core structure and interfaces <modulename>.go
- internal/<modulename> - contains implementation of the module including handler ( http ), service ( business logic ), repository ( database ops )
- infra - contains server, routing, db etc.. to run the server.
- cmd - contains cobra cli commands like serve