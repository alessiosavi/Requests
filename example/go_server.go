// A go webserver for benchmark the request
package main

import (
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)

var (
	counter int
	addr    = flag.String("addr", ":9999", "TCP address to listen to")
)

func main() {
	if err := fasthttp.ListenAndServe(*addr, requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	//counter = counter + 1
	//_, _ = fmt.Fprintf(ctx, "Counter: "+strconv.Itoa(counter))
	fmt.Fprintf(ctx, "")
}
