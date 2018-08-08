package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/thehowl/fasthttprouter"
)

// currentServer is the current VPN connection
var currentServer *VPN

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
	router.GET("/disconnect", disconnectHandler)
	router.GET("/where/am/i/connected", statusHandler)
	router.GET("/connecto/:country/:proto", connHandler)
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
		fmt.Fprintf(ctx, "Connected to %s", currentServer.Domain)
	}
}

func statusHandler(ctx *fasthttp.RequestCtx) {
	if currentServer == nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
	} else {
		io.WriteString(ctx, currentServer.Country)
	}
}

func disconnectHandler(ctx *fasthttp.RequestCtx) {
	currentServer = nil
	stopOpenVPN()
	io.WriteString(ctx, "Disconnected")
}
