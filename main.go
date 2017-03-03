package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

type Options struct {
	Port          string
	HeartBeat     bool
	ShutdownDelay int
}

var opts Options

func main() {
	app := cli.NewApp()
	app.Name = "explorer"
	app.Usage = "web server to introspect a running container"
	app.Version = "0.3.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Usage:       "port number",
			Value:       "5002",
			EnvVar:      "PORT",
			Destination: &opts.Port,
		},
		cli.BoolFlag{
			Name:        "heartbeat",
			Usage:       "heartbeat",
			EnvVar:      "HEARTBEAT",
			Destination: &opts.HeartBeat,
		},
		cli.IntFlag{
			Name:        "delay",
			Usage:       "seconds to wait after shutdown completes before exiting",
			Value:       5,
			EnvVar:      "DELAY",
			Destination: &opts.ShutdownDelay,
		},
	}
	app.Action = Run
	app.Run(os.Args)
}

func Log(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%v %v\n", req.Method, req.RequestURI)
		h.ServeHTTP(w, req)
	})
}

func Run(_ *cli.Context) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exit := make(chan int)
	go HandleSignals(ch, exit)

	ctx, cancel := context.WithCancel(context.Background())

	if opts.HeartBeat {
		go HeartBeat(ctx)
	}

	mux := http.NewServeMux()
	mux.Handle("/_/echo", Log(http.HandlerFunc(Echo)))
	mux.Handle("/_/env", Log(http.HandlerFunc(Env)))
	mux.Handle("/_/healthcheck", Log(http.HandlerFunc(Health)))
	mux.Handle("/", Log(http.FileServer(http.Dir("/"))))

	server := &http.Server{
		Addr:    ":" + opts.Port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-exit
	child, _ := context.WithTimeout(ctx, 5*time.Second)
	server.Shutdown(child)
	cancel()

	log.Println("server graceful shutdown")

	if opts.ShutdownDelay > 0 {
		log.Printf("delaying an additional %v seconds\n", opts.ShutdownDelay)
		for i := 0; i < opts.ShutdownDelay; i++ {
			time.Sleep(time.Second)
			log.Println("delay ...", i+1)
		}
	}
}

func HandleSignals(ch <-chan os.Signal, exit chan<- int) {
	for {
		s := <-ch

		switch s {
		// kill -SIGHUP XXXX
		case syscall.SIGHUP:
			log.Println("received SIGHUP")

		// kill -SIGINT XXXX or Ctrl+c
		case syscall.SIGINT:
			exit <- 0
			return

		// kill -SIGTERM XXXX
		case syscall.SIGTERM:
			exit <- 0
			return

		// kill -SIGQUIT XXXX
		case syscall.SIGQUIT:
			exit <- 0
			return

		default:
			exit <- 0
			return
		}
	}
}

func HeartBeat(ctx context.Context) {
	t := time.NewTimer(time.Second)
	defer t.Stop()

	for {
		t.Reset(time.Second)

		select {
		case <-ctx.Done():
			return
		case <-t.C:
			log.Println("heartbeat")
		}
	}
}

func Echo(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "<html>")
	io.WriteString(w, `<head>
	<style type="text/css">
		table {
			border-collapse: collapse;
			border-spacing: 0;
		}

		tr:nth-child(odd) {
			background-color: #f0f0f0;
		}

		tr:first-of-type {
			border-top: 1px solid #c0c0c0;
		}

		td {
			font-family: arial, sans-serif;
			padding: 5px 10px;
			border-bottom: 1px solid #c0c0c0;
		}

	</style>
</head>`)
	io.WriteString(w, "<table>")
	for k, values := range req.Header {
		for _, v := range values {
			io.WriteString(w, `<tr><td style="width: 400px;">`)
			io.WriteString(w, k)
			io.WriteString(w, "</td><td>")
			io.WriteString(w, v)
			io.WriteString(w, "</td></tr>")
		}
	}
	io.WriteString(w, "</table>")

	io.WriteString(w, "<pre>")
	io.Copy(w, req.Body)
	io.WriteString(w, "</pre>")

	io.WriteString(w, "</html>")
	req.Body.Close()
}

func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"status":"ok"}`)
}

func Env(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "<html>")
	io.WriteString(w, `<head>
	<style type="text/css">
		table {
			border-collapse: collapse;
			border-spacing: 0;
		}

		tr:nth-child(odd) {
			background-color: #f0f0f0;
		}

		tr:first-of-type {
			border-top: 1px solid #c0c0c0;
		}

		td {
			font-family: arial, sans-serif;
			padding: 5px 10px;
			border-bottom: 1px solid #c0c0c0;
		}

	</style>
</head>`)
	io.WriteString(w, "<table>")

	keys := make([]string, 0, len(os.Environ()))
	m := map[string]string{}

	for _, v := range os.Environ() {
		segments := strings.SplitN(v, "=", 2)
		keys = append(keys, segments[0])
		m[segments[0]] = segments[1]
	}

	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		io.WriteString(w, `<tr><td style="width: 400px;">`)
		io.WriteString(w, k)
		io.WriteString(w, "</td><td>")
		io.WriteString(w, v)
		io.WriteString(w, "</td></tr>")
	}

	io.WriteString(w, "</table>")
	io.WriteString(w, "</html>")
}
