[![Build Status on master](https://travis-ci.org/SKF/go-utility.svg?branch=master)](https://travis-ci.org/SKF/go-utility) [![Go Report Card](https://goreportcard.com/badge/github.com/SKF/go-utility)](https://goreportcard.com/report/github.com/SKF/go-utility)

# go-utility 

## Migration
If you previously were using `github.com/SKF/go-utility/v2/env` and want to use version `v3.*.*`, update import path to `github.com/SKF/go-utility/v3/env`. 

For more information:
- https://blog.golang.org/v2-go-modules
- https://research.swtch.com/vgo-import

### Migration from `2.*` to `3.*`
- `github.com/aws/aws-sdk-go` has been removed in favor of `github.com/aws/aws-sdk-go-v2` [Guide](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/migrate-gosdk.html)
- `trace/aws.WrapSession` has been removed and replaced with `trace/aws.AddTraceMiddleware`