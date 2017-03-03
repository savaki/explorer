package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli"
)

type Options struct {
	Port string
}

var opts Options

func main() {
	app := cli.NewApp()
	app.Name = "explorer"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Usage:       "port number",
			Value:       "5002",
			EnvVar:      "PORT",
			Destination: &opts.Port,
		},
	}
	app.Action = Run
	app.Run(os.Args)
}

func Run(_ *cli.Context) {
	http.HandleFunc("/_/echo", Echo)
	http.HandleFunc("/_/env", Env)
	http.Handle("/", http.FileServer(http.Dir("/")))

	err := http.ListenAndServe(":"+opts.Port, nil)
	if err != nil {
		log.Fatalln(err)
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
	fmt.Println("")
}
