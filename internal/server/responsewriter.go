package server

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	code int 
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijack not supported by underlying ResponseWriter")
	}
	return h.Hijack()
}