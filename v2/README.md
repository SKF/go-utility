[![Build Status on master](https://travis-ci.org/SKF/go-utility.svg?branch=master)](https://travis-ci.org/SKF/go-utility) [![Go Report Card](https://goreportcard.com/badge/github.com/SKF/go-utility)](https://goreportcard.com/report/github.com/SKF/go-utility)

# go-utility 

## Supported packages
- array
- auth
  - secretsmanagerauth
- datadog
- env
- http-middleware
- http-model
- http-server
- jwk
- jwt
- log
- timeutils
- useridcontext
- uuid

## Migration
If you previously were using `github.com/SKF/go-utility/env` and want to use version `v2.*.*`, update import path to `github.com/SKF/go-utility/v2/env`.

For more information:
- https://blog.golang.org/v2-go-modules
- https://research.swtch.com/vgo-import

### Migration from `1.*` to `2.*`
- `http-middleware` have some updates, more info [here](http-middleware/README.md).
- `grpc-interceptor/requestid` has been removed in favor of Opencensus.
