package main

import (
	"fmt"
	"log"

	requests "github.com/alessiosavi/Requests"
	"github.com/alessiosavi/Requests/datastructure"
)

func main() {
	// exampleGETRequest()
	// examplePOSTRequest()
	exampleParallelRequest()
}

func examplePOSTRequest() {
	// Initialize request
	var req requests.Request

	// For set the headers you can use two different method

	// Method 1
	//Set custom headers directly in the method
	req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "Basic cG9zdG1hbjpwYXNzd29yZA==")

	// Method 2
	// Create a list of headers and use in the method
	var headers []string
	headers = append(headers, "Content-Type")
	headers = append(headers, "text/plain; charset=UTF-8")
	headers = append(headers, "Authorization")
	headers = append(headers, "Basic cG9zdG1hbjpwYXNzd29yZA==")
	req.CreateHeaderList(headers...)

	// Set the body of the request
	body := []byte("This is the body data of the POST request")

	// Send the request and save to a properly structure
	// POST, with BODY data and enabling SSL certificate validation (skipTLS: false)
	// NOTE: You can skip self signed certificate validatation with skipTLS=True
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

func exampleParallelRequest() {
	// This array will contains the list of request
	var reqs []requests.Request
	// This array will contains the response from the given request
	var response []datastructure.Response

	// Set to run at max 12 request in parallele
	var N int = 12
	// Create the list of request
	for i := 0; i < 1000; i++ {
		req, err := requests.InitRequest("https://127.0.0.1:5000", "GET", nil, nil, i%2 == 0) // Alternate cert validation
		if err != nil {
			log.Println("Skipping request [", i, "]. Error: ", err)
		} else {
			reqs = append(reqs, *req)
		}
	}

	response = requests.ParallelRequest(reqs, N)
	for i := range response {
		log.Println("Request [", i, "] -> ", response[i].Dump())
	}
}
