## Migration
### Migration from `1.*` to `2.*`

- CorsMiddleware
  
  The old `CorsMiddlewareV2` is now replaced by `CorsMiddleware`
  
  Migration steps:
  
  - If you are using `CorsMiddlewareV2`, change it to `CorsMiddleware`

  - If you are using `CorsMiddleware`, add this code to the endpoints that need CORS, remember to configure it to your needs.
    ``` go
	server.mux.
		HandleFunc("<path>", http_middleware.Options(
			[]string{http.MethodGet},
			[]string{http_model.HeaderContentType},
		)).
		Methods(http.MethodOptions)
    ```

- AuthenticateMiddleware

  The old `AuthenticateMiddleware` has been modified to take care of retrieving User ID by itself.
  
  Migration steps:

  - Add the following code to your main file.
    ``` go
  	http_middleware.Configure(
        http_middleware.Config{Stage: authenticateStage},
    )
    ```