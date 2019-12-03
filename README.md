# Requests

A golang library for avoid headache during HTTP request

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/50714e195c544ab1b5e4e40d94a43998)](https://www.codacy.com/manual/alessiosavi/Requests?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=alessiosavi/Requests&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/alessiosavi/Requests)](https://goreportcard.com/report/github.com/alessiosavi/Requests) [![GoDoc](https://godoc.org/github.com/alessiosavi/Requests?status.svg)](https://godoc.org/github.com/alessiosavi/Requests) [![License](https://img.shields.io/github/license/alessiosavi/Requests)](https://img.shields.io/github/license/alessiosavi/Requests) [![Version](https://img.shields.io/github/v/tag/alessiosavi/Requests)](https://img.shields.io/github/v/tag/alessiosavi/Requests) [![Code size](https://img.shields.io/github/languages/code-size/alessiosavi/Requests)](https://img.shields.io/github/languages/code-size/alessiosavi/Requests) [![Repo size](https://img.shields.io/github/repo-size/alessiosavi/Requests)](https://img.shields.io/github/repo-size/alessiosavi/Requests) [![Issue open](https://img.shields.io/github/issues/alessiosavi/Requests)](https://img.shields.io/github/issues/alessiosavi/Requests)
[![Issue closed](https://img.shields.io/github/issues-closed/alessiosavi/Requests)](https://img.shields.io/github/issues-closed/alessiosavi/Requests)

## Example usage

In order to make a `Request`, you need to initialize the client:

```go
// Initialize request
var req requests.Request
```

Now the request is ready to be populated with `headers` or `body-data` if necessary.

You create a list of headers and the explode the data into the delegated method, or pass a given number of headers to the method:

```go
//Set custom headers directly in the method
req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "Basic cG9zdG1hbjpwYXNzd29yZA==")
// Or create a list of headers and use in the method
var headers []string
headers = append(headers, "Content-Type")
headers = append(headers, "text/plain; charset=UTF-8")
headers = append(headers, "Authorization")
headers = append(headers, "Basic cG9zdG1hbjpwYXNzd29yZA==")
req.CreateHeaderList(headers...)

```

Now you can send a request to the URL

```go
// Send the request and save to a properly structure
// GET, without BODY data (only used in POST), and enabling SSL certificate validation (skipTLS: false)
response := req.SendRequest("https://postman-echo.com/get?foo1=bar1&foo2=bar2", "GET", nil, false)

// Debug the response
fmt.Println("Headers: ", response.Headers)
fmt.Println("Status code: ", response.StatusCode)
fmt.Println("Time elapsed: ", response.Time)
fmt.Println("Error: ", response.Error)
fmt.Println("Body: ", string(response.Body))
```

Please, refer to [the example code](example/req_example.go)