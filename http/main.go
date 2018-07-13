package main

import (
	"io"
	"net/http"
)

type a struct{}

func (*a) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.String()
	switch path {
	case "/":
		io.WriteString(w, "<h1>root</h1><a href=\"abc\">abc</a>")
	case "/abc":
		io.WriteString(w, "<h1>abc</h1><a href=\"/\">root</a>")
	case "/script":
		io.WriteString(w, "hello world")
	}
}

func main() {
	http.ListenAndServe(":8080", &a{}) //第2个参数需要实现Hander接口的struct，a满足
}
