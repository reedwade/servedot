package main

import (
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var listenOn = ":8000"

func init() {
	flag.StringVar(&listenOn, "listen", listenOn, "address and port to listen on")
}

type logRequest struct {
	next http.Handler
}

func (l logRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.next.ServeHTTP(w, r)
	log.Infof("%s %s %s %s --> %s", r.Method, r.URL, r.RemoteAddr, r.Header, w.Header())
}
func wrapLogRequest(next http.Handler) http.Handler {
	return logRequest{next: next}
}

type addHeader struct {
	next        http.Handler
	header      string
	headerValue string
}

func (l addHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(l.header, l.headerValue)
	l.next.ServeHTTP(w, r)
}
func wrapAddHeader(header, headerValue string, next http.Handler) http.Handler {
	return addHeader{next: next, header: header, headerValue: headerValue}
}

func main() {
	flag.Parse()

	log.Infof("listening on %s", listenOn)

	s := http.FileServer(http.Dir("./"))

	http.Handle("/",
		wrapAddHeader("X-Listening-On", listenOn, wrapLogRequest(s)),
	)

	log.Error(http.ListenAndServe(listenOn, nil))
}
