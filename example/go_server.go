// A go webserver for benchmark the request
package main

import (
	"flag"
	"github.com/valyala/fasthttp"
)

var (
	//counter int
	addr    = flag.String("addr", ":9999", "HTTP address to listen to")
	addrTLS    = flag.String("addrTLS", ":9990", "HTTPS address to listen to")
)

func main() {
	flag.Parse()
	if *addr == ":9999" && *addrTLS == ":9990"{
		flag.PrintDefaults()
	}
	go fasthttp.ListenAndServe(*addr, requestHandler)
	fasthttp.ListenAndServe(*addrTLS, requestHandler)

}

func requestHandler(ctx *fasthttp.RequestCtx) {
	//counter = counter + 1
	//_, _ = fmt.Fprintf(ctx, "Counter: "+strconv.Itoa(counter))
	//_, _ = fmt.Fprintf(ctx, "")
	ctx.SetConnectionClose()
}
