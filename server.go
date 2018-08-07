package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/signal"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/thehowl/fasthttprouter"
)

var templates = func() *template.Template {
	t, err := template.ParseGlob(*templateDir)
	if err != nil {
		panic(err)
	}
	return t
}()

func startServer() error {
	debug("starting server")
	fs := fasthttp.FS{
		Root:       *httpDir,
		IndexNames: []string{"index.html"},
		Compress:   true,
	}
	router := fasthttprouter.New()
	router.GET("/", rootHandler)
	router.GET("/:country/:proto", connHandler)
	router.NotFound = fs.NewRequestHandler()
	// TODO: Implement TLS
	server := fasthttp.Server{
		Handler:          router.Handler,
		LogAllErrors:     false,
		Name:             "GoNordVPN 0.1",
		DisableKeepalive: true,
	}
	go server.ListenAndServe(":9114")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	select {
	case <-ch:
	}
	log.Println("Shutting down")
	return server.Shutdown()
}

func rootHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html")
	templates.ExecuteTemplate(ctx, "map", config.VPNList)
}

func connHandler(ctx *fasthttp.RequestCtx) {
	country := ctx.UserValue("country")
	if country == nil {
		ctx.Error("bad country", fasthttp.StatusBadRequest)
		return
	}
	proto := ctx.UserValue("proto")
	if proto != "tcp" || proto != "udp" {
		proto = "udp"
	}
	stopOpenVPN()
	err := startOpenVPN(country.(string), proto.(string))
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
		fmt.Fprintf(ctx, "Connected to %s", country)
	}
}
