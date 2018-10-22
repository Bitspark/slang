package daemon

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func proxyRequestToOperator(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if len(path) == 0 || path[0] != '/' {
		w.WriteHeader(400)
		return
	}

	path = path[1:]
	sp := strings.Index(path, "/")
	handle := ""
	newPath := ""

	if sp == -1 {
		handle = path
		newPath = "/"
	} else {
		handle = path[:sp]
		newPath = path[sp:]
	}

	handleID, err := strconv.ParseInt(handle, 10, 64)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	operator, ok := runningInstances[handleID]
	if !ok {
		w.WriteHeader(404)
		return
	}

	newPort := operator.port
	newURL := url.URL{}
	newURL.Scheme = "http"
	newURL.Host = "localhost:" + strconv.Itoa(newPort)
	newURL.Path = newPath
	newURL.RawQuery = r.URL.RawQuery

	req, err := http.NewRequest(r.Method, newURL.String(), r.Body)
	for key, values := range r.Header {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	for key, values := range resp.Header {
		w.Header().Del(key)
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}