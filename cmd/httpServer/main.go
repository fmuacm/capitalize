package main

import (
	"net/http"
	"os"
	"time"

	"fmt"

	"strings"

	"encoding/json"

	"net/url"

	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/fmuacm/capitalize"
	"github.com/joho/godotenv"
)

const (
	//Creating default values if one is not supplied.
	defaultServerHost = "localhost"
	defaultServerPort = 8081
)

var (
	//Setting global variables for the main package.
	//This is an okay way to get values out of the init function to the main function.
	serverHost string
	serverPort int
)

//Create a stucture for a response to send back to client
type defaultResponse struct {
	Msg string `json:"msg,omitempty"`
	Err string `json:"error,omitempty"`
}

//This is to extantiate the defaultResponse structure
func newDefaultResponse(msg string, err error) *defaultResponse {
	resp := &defaultResponse{
		Msg: msg,
	}
	if err != nil {
		resp.Err = err.Error()
	}
	return resp
}

//This is to capture the status of the response
type loggingResponseWriter struct {
	status int
	http.ResponseWriter
}

//This is to impliment the ResponseWriter interface
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

//This handler is used to handle logging
type loggingHandler struct {
	startTime time.Time
	h         http.HandlerFunc
}

//This is to turn the loggingHandler into a handler
func (l *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.startTime = time.Now()
	myW := &loggingResponseWriter{-1, w}
	logCtx := log.WithFields(log.Fields{
		"url":    r.URL.Path,
		"query":  r.URL.RawQuery,
		"method": r.Method,
	})
	logCtx.Debug("Started request")
	l.h.ServeHTTP(myW, r)
	duration := time.Since(l.startTime)
	logCtx.WithFields(log.Fields{
		"status":   myW.status,
		"duration": duration,
	}).Debug("End request")
}

//this is to extantiate the loggingHandler
func newLoggingHandler(h http.HandlerFunc) *loggingHandler {
	return &loggingHandler{
		h: h,
	}
}

//This is of type HandlerFunc which is used by all responses to the root path
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	basePath := new(url.URL)
	basePath.Host = r.URL.Host
	basePath.Path = r.URL.Path
	if strings.ToUpper(r.Method) != http.MethodGet {
		respStruct := newDefaultResponse("", fmt.Errorf("%s only supports the %s method", basePath.String(), http.MethodGet))
		write(w, respStruct, http.StatusBadRequest)
		return
	}
	msg := r.URL.Query().Get("s")
	if msg == "" {
		respStruct := newDefaultResponse("", fmt.Errorf("%s expects a query parameter s to be present", basePath.String()))
		write(w, respStruct, http.StatusBadRequest)
		return
	}

	respStruct := newDefaultResponse(capitalize.Format(msg), nil)
	write(w, respStruct, http.StatusOK)
}

//Initialize the environment of the server
func init() {
	godotenv.Load()
	environment := os.Getenv("ENVIRONMENT")
	switch strings.ToLower(environment) {
	case "develop":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	serverHost = os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = defaultServerHost
	}
	var err error
	serverPortString := os.Getenv("SERVER_PORT")
	if serverPortString == "" {
		log.Warnf("Could not get server port. Defaulting to %d", defaultServerPort)
		serverPortString = strconv.Itoa(defaultServerPort)
	}
	serverPort, err = strconv.Atoi(serverPortString)
	if err != nil || serverPort == 0 {
		log.WithError(err).Warnf("Could not get server port. Defaulting to %d", defaultServerPort)
		serverPort = defaultServerPort
	}
}

//This starts the server
func main() {
	http.Handle("/", newLoggingHandler(defaultHandler))
	listenString := fmt.Sprintf("%s:%d", serverHost, serverPort)
	log.Infof("Listening on %s", listenString)
	if err := http.ListenAndServe(listenString, nil); err != nil {
		log.Errorf("Could not listen on %s", listenString)
		os.Exit(0)
	}
}

//This is to write a response to the client
func write(w http.ResponseWriter, respStruct interface{}, respStatus int) {
	w.Header().Add("Content-Type", "application/json")
	resp, err := json.Marshal(respStruct)
	if err != nil {
		errStruct := newDefaultResponse("Could not marshal response into json object", err)
		log.Error("Could not marshal response into json object.")
		write(w, errStruct, http.StatusInternalServerError)
		return
	}
	resp = append(resp, '\r', '\n')
	w.WriteHeader(respStatus)
	w.Write(resp)
}
