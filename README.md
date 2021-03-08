# Requests

A golang library for avoid headache during HTTP request

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/50714e195c544ab1b5e4e40d94a43998)](https://www.codacy.com/manual/alessiosavi/Requests?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=alessiosavi/Requests&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/alessiosavi/Requests)](https://goreportcard.com/report/github.com/alessiosavi/Requests) [![GoDoc](https://godoc.org/github.com/alessiosavi/Requests?status.svg)](https://godoc.org/github.com/alessiosavi/Requests) [![License](https://img.shields.io/github/license/alessiosavi/Requests)](https://img.shields.io/github/license/alessiosavi/Requests) [![Version](https://img.shields.io/github/v/tag/alessiosavi/Requests)](https://img.shields.io/github/v/tag/alessiosavi/Requests) [![Code size](https://img.shields.io/github/languages/code-size/alessiosavi/Requests)](https://img.shields.io/github/languages/code-size/alessiosavi/Requests) [![Repo size](https://img.shields.io/github/repo-size/alessiosavi/Requests)](https://img.shields.io/github/repo-size/alessiosavi/Requests) [![Issue open](https://img.shields.io/github/issues/alessiosavi/Requests)](https://img.shields.io/github/issues/alessiosavi/Requests)
[![Issue closed](https://img.shields.io/github/issues-closed/alessiosavi/Requests)](https://img.shields.io/github/issues-closed/alessiosavi/Requests)

## Example usage

`Requests` can work in two different method:

- Single request
- Multiple request in parallel

### Single Request

In order to make a `Request`, you need to initialize the client:

```go
// Initialize request
var req requests.Request
```

Now the request is ready to be populated with `headers` or `body-data` if necessary.

You can create a list of headers and explode the data into the delegated method. In alternative, you can pass a given number of headers to the method.  
***NOTE***: The headers have to be a "key:value" list so the first argument will be the key of the header, the one after will be the value.

```go
//Set custom headers directly in the method
req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "Basic cG9zdG1hbjpwYXNzd29yZA==")
// Or create a list of headers and use them in the method
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

### Multiple request in parallel

In order to use the parallel request, you have to create a list of request that have to be executed, than you can call the delegated method for send them in parallel, choosing how many requests have to be sent in parallel.

In first instance populate an array with the requests that you need to send:

```go
// This array will contains the list of request
var reqs []requests.Request

// N is the number of request to run in parallel, in order to avoid "TO MANY OPEN FILES"
var N int = 12

// Create the list of request
for i := 0; i < 1000; i++ {
    // Example python server, you can find under the "example" folder
    req, err := requests.InitRequest("https://127.0.0.1:5000", "GET", nil, nil, i%2 == 0) // Alternate cert validation
    if err != nil {
        log.Println("Skipping request [", i, "]. Error: ", err)
    } else {
        // If no error, we can append the request created to the list of request that we need to send
        reqs = append(reqs, *req)
    }
}
```

```go
// This array will contains the response from the given request
var response []datastructure.Response

// send the request using N request to send in parallel
response = requests.ParallelRequest(reqs, N)

// Print the response
for i := range response {
    log.Println("Request [", i, "] -> ", response[i].Dump())
}
```

## More example

Please, refer to the [example code](example/req_example.go)
