package main

import (
	"flag"
	"net/http"
	"time"

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
	log.Infof("%s %s %s <--%s -->%s", r.Method, r.Host, r.URL.Path, r.Header, w.Header())
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

type slowResponse struct {
	next  http.Handler
	delay time.Duration
}

func (l slowResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(l.delay)
	l.next.ServeHTTP(w, r)
}
func wrapSlowResponse(delay time.Duration, next http.Handler) http.Handler {
	return slowResponse{next: next, delay: delay}
}

func main() {
	flag.Parse()

	delay := time.Second * 10

	log.Infof("listening on %s", listenOn)
	log.Infof("prefix with '_slow/' for delayed response by %s", delay)

	s := http.FileServer(http.Dir("./"))
	s = wrapLogRequest(s)
	s = wrapAddHeader("X-Listening-On", listenOn, s)
	http.Handle("/", s)

	s = http.FileServer(http.Dir("./"))
	s = wrapLogRequest(s)
	s = wrapAddHeader("X-Listening-On", listenOn, s)
	s = wrapSlowResponse(delay, s)
	http.Handle("/_slow/", http.StripPrefix("/_slow/", s))

	log.Error(http.ListenAndServe(listenOn, nil))
}
