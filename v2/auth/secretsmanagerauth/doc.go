// Package secretsmanagerauth provides a package for
// authenticating with credentials stored in Secrets Manager.
//
// Examples
//
// SignIn
//     func main() {
//         sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")}))
//         sess = aws_trace.WrapSession(sess)
//
//         ctx := context.Background()
//
//         conf := auth.Config{
//             WithDatadogTracing        bool // used when you trace your application with Datadog
//             WithOpenCensusTracing     bool // default and used when you trace your application with Open Census
//             ServiceName               string // needed when using lambda and Datadog for tracing
//             AWSSession:               sess,
//             AWSSecretsManagerAccount: "...",
//             AWSSecretsManagerRegion:  "eu-west-1",
//             Stage:                    "sandbox",
//             SecretKey:                "user-credentials/iot_service",
//         }
//
//         auth.Configure(conf)
//         fmt.Println(auth.SignIn(ctx))
//         fmt.Println(auth.GetTokens())
//     }
//
package secretsmanagerauth
