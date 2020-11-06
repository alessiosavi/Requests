package datastructure

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Response is delegated to save the necessary information related to an HTTP call
type Response struct {
	Headers    map[string]string
	Body       []byte
	StatusCode int
	Time       time.Duration
	Error      error
	Cookie     []*http.Cookie
	Response   *http.Response
}

// Dump method is delegated to dump the information related to the request
func (resp *Response) Dump() string {
	var sb strings.Builder

	sb.WriteString("=========================\n")
	sb.WriteString("Headers: ")
	sb.WriteString(fmt.Sprintf("%s", resp.Headers))
	sb.WriteString("\n")

	sb.WriteString("Status Code: ")
	sb.WriteString(fmt.Sprintf("%d", resp.StatusCode))
	sb.WriteString("\n")

	sb.WriteString("Time elapsed: ")
	sb.WriteString(fmt.Sprintf("%v", resp.Time))
	sb.WriteString("\n")

	sb.WriteString("Body: ")
	sb.WriteString(fmt.Sprint(string(resp.Body)))
	sb.WriteString("\n")

	sb.WriteString("Error: ")
	sb.WriteString(fmt.Sprintf("%v", resp.Error))
	sb.WriteString("\n")
	sb.WriteString("=========================\n")

	return sb.String()
}
