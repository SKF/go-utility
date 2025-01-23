[![Build Status on master](https://travis-ci.org/SKF/go-utility.svg?branch=master)](https://travis-ci.org/SKF/go-utility) [![Go Report Card](https://goreportcard.com/badge/github.com/SKF/go-utility)](https://goreportcard.com/report/github.com/SKF/go-utility)

# go-utility 

## Supported packages
- array
- auth
  - secretsmanagerauth
- datadog
- env
- grpc-interceptor
- http-middleware
- http-model
- http-server
- jwk
- jwt
- log
- timeutils
- uuid

## Migration
If you previously were using `github.com/SKF/go-utility/env` and want to use version `vN.*.*`, update import path to `github.com/SKF/go-utility/vN/env`.

For more information:
- https://blog.golang.org/v2-go-modules
- https://research.swtch.com/vgo-import

Migrations
- [Migration from `1.*` to `2.*`](v2/README.md)
- [Migration from `2.*` to `3.*`](v3/README.md)
