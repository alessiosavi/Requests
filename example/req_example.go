package main

import (
	"fmt"

	requests "github.com/alessiosavi/Requests"
)

func main() {
	exampleGETRequest()
	examplePOSTRequest()
}

func examplePOSTRequest() {
	// Initialize request
	var req requests.Request

	//Set custom headers directly in the method
	req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "Basic cG9zdG1hbjpwYXNzd29yZA==")
	// Or create a list of headers and use in the method
	var headers []string
	headers = append(headers, "Content-Type")
	headers = append(headers, "text/plain; charset=UTF-8")
	headers = append(headers, "Authorization")
	headers = append(headers, "Basic cG9zdG1hbjpwYXNzd29yZA==")
	req.CreateHeaderList(headers...)

	body := []byte("This is the body data of the POST request")
	// Send the request and save to a properly structure
	// GET, without BODY data (only used in POST), and enabling SSL certificate validation (skipTLS: false)
	response := req.SendRequest("https://postman-echo.com/post", "POST", body, false)

	// Use the response data
	fmt.Println("Headers: ", response.Headers)
	fmt.Println("Status code: ", response.StatusCode)
	fmt.Println("Time elapsed: ", response.Time)
	fmt.Println("Error: ", response.Error)
	fmt.Println("Body: ", string(response.Body))

	// Or print them
	fmt.Println(response.Dump())

}

func exampleGETRequest() {
	// Initialize request
	var req requests.Request

	//Set custom headers directly in the method
	req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "Basic cG9zdG1hbjpwYXNzd29yZA==")
	// Or create a list of headers and use in the method
	var headers []string
	headers = append(headers, "Content-Type")
	headers = append(headers, "text/plain; charset=UTF-8")
	headers = append(headers, "Authorization")
	headers = append(headers, "Basic cG9zdG1hbjpwYXNzd29yZA==")
	req.CreateHeaderList(headers...)

	// Send the request and save to a properly structure
	// GET, without BODY data (only used in POST), and enabling SSL certificate validation (skipTLS: false)
	response := req.SendRequest("https://postman-echo.com/get?foo1=bar1&foo2=bar2", "GET", nil, false)

	fmt.Println("Headers: ", response.Headers)
	fmt.Println("Status code: ", response.StatusCode)
	fmt.Println("Time elapsed: ", response.Time)
	fmt.Println("Error: ", response.Error)
	fmt.Println("Body: ", string(response.Body))

}
