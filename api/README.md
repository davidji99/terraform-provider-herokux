# herokux-go
`herokux-go` is a Go client library for accessing various Heroku APIs that aren't documented as their Platform APIs.

There are plans to extract this package out of the current repository and into a standalone library.

# Example
```go
	api, clientInitErr := api.New(config.APIToken("some_token"))
	if clientInitErr != nil {
		return clientInitErr
	}

	fmt.Println(api)
```