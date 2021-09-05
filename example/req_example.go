package main

import (
	"fmt"
	requests "github.com/alessiosavi/Requests"
	"github.com/alessiosavi/Requests/datastructure"
	"log"
	"net/http"
	"time"
)

func main() {
	exampleGETRequest()
	examplePOSTRequest()
	exampleParallelRequest()
	exampleBasicAuth()
}

func examplePOSTRequest() {
	// Initialize request
	var req *requests.Request
	var err error
	if req, err = requests.InitRequest("https://postman-echo.com", "POST", []byte("get?foo1=bar1&foo2=bar2"), false, true); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	//Set custom headers directly in the method
	if err = req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "Basic cG9zdG1hbjpwYXNzd29yZA=="); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Or create a list of headers and use in the method
	var headers []string
	headers = append(headers, "Content-Type")
	headers = append(headers, "text/plain; charset=UTF-8")
	headers = append(headers, "Authorization")
	headers = append(headers, "Basic cG9zdG1hbjpwYXNzd29yZA==")
	if err = req.CreateHeaderList(headers...); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Send the request and save to a properly structure
	// GET, without BODY data (only used in POST), and enabling SSL certificate validation (skipTLS: false)
	response := req.ExecuteRequest(&http.Client{})

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
	var req *requests.Request
	req, err := requests.InitRequest("https://postman-echo.com/get?foo1=bar1&foo2=bar2", "GET", nil, false, true)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	//Set custom headers directly in the method
	err = req.CreateHeaderList("Content-Type", "text/plain; charset=UTF-8", "Authorization", "alessio:savi")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Or create a list of headers and use in the method
	var headers []string
	headers = append(headers, "Content-Type")
	headers = append(headers, "text/plain; charset=UTF-8")
	headers = append(headers, "Authorization")
	headers = append(headers, "alessio:savi")
	err = req.CreateHeaderList(headers...)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Send the request and save to a properly structure
	// GET, without BODY data (only used in POST), and enabling SSL certificate validation (skipTLS: false)
	response := req.ExecuteRequest(&http.Client{})

	fmt.Println("Headers: ", response.Headers)
	fmt.Println("Status code: ", response.StatusCode)
	fmt.Println("Time elapsed: ", response.Time)
	fmt.Println("Error: ", response.Error)
	fmt.Println("Body: ", string(response.Body))
}

func exampleParallelRequest() {
	start := time.Now()
	// This array will contains the list of request
	var reqs []requests.Request
	// This array will contains the response from the given request
	var response []datastructure.Response

	// Set to run at max 100 request in parallel (use CPU count for best effort)
	var N = 100
	// Create the list of request
	for i := 0; i < 20000; i++ {
		// Run against the `server_example.py` present in this folder
		req, err := requests.InitRequest("https://127.0.0.1:5000", "GET", nil, i%2 == 0, false) // Alternate cert validation
		if err != nil {
			log.Println("Skipping request [", i, "]. Error: ", err)
		} else {
			req.SetTimeout(10 * time.Second)
			reqs = append(reqs, *req)
		}
	}

	// Run the request in parallel
	response = requests.ParallelRequest(reqs, N)

	elapsed := time.Since(start)

	for i := range response {
		// Print the response
		log.Println("Request [", i, "] -> ", response[i].Dump())
	}
	log.Printf("Requests took %s", elapsed)
}

func exampleBasicAuth() {
	req, err := requests.InitRequest("https://postman-echo.com/basic-auth", "GET", []byte{}, false, false)
	if err != nil {
		fmt.Println("ERROR! ", err)
	}
	_ = req.CreateHeaderList("Accept", "application/json", "Accept-Language", "en_US", "Authorization", "postman:password")
	client := &http.Client{}
	resp := req.ExecuteRequest(client)
	fmt.Println(resp.Dump())
}
