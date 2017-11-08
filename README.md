# Register Ticket

[![Coverage Status](https://coveralls.io/repos/github/OpenIndustryCloud/fission-go-register-ticket/badge.svg?branch=master)](https://coveralls.io/github/OpenIndustryCloud/fission-go-register-ticket?branch=master)


The `go` runtime uses the [`plugin` package](https://golang.org/pkg/plugin/) to dynamically load an HTTP handler.

## Requirements

First, set up your fission deployment with the go environment.

```
fission env create --name go-env --image fission/go-env:1.8.1
```

To ensure that you build functions using the same version as the
runtime, fission provides a docker image and helper script for
building functions.

## Example Usage

### register-ticket.go

`register-ticket.go` is an API which accepts JSON payload compliant to Zen Deks and creates Ticket with the Payload

```bash
# Download the build helper script
$ curl https://raw.githubusercontent.com/fission/fission/master/environments/go/builder/go-function-build > go-function-build
$ chmod +x go-function-build

# Build the function as a plugin. Outputs result to 'function.so'
$ go-function-build register-ticket.go

# Upload the function to fission
$ fission function create --name register-ticket --env go-env --package function.so

# Map /hello to the hello function
$ fission route create --method POST --url /register-ticket --function register-ticket

# Run the function
$ curl -d '{--INPUT JSON--}' -H "Content-Type: application/json" -X POST http://$FISSION_ROUTER/register-ticket

#sample input

#sample output

##Code Coverage



```
test
test
