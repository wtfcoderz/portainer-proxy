package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	const (
		defaultPort          = ":9001"
		defaultPortUsage     = "default server port, ':9001'"
		defaultTarget        = "http://127.0.0.1:9000"
		defaultTargetUsage   = "default redirect url, 'http://127.0.0.1:9000'"
		defaultUsername      = "admin"
		defaultUsernameUsage = "username to connect to portainer, 'admin'"
		defaultPassword      = "password"
		defaultPasswordUsage = "password to connect to portainer, 'password'"
		defaultEndpointID    = "1"
		defaultEndpointUsage = "portainer endpoint id, '1'"
	)

	// flags
	port := flag.String("port", defaultPort, defaultPortUsage)
	redirecturl := flag.String("url", defaultTarget, defaultTargetUsage)
	username := flag.String("username", defaultUsername, defaultUsernameUsage)
	password := flag.String("password", defaultPassword, defaultPasswordUsage)
	endpointID := flag.String("endpoint", defaultEndpointID, defaultEndpointUsage)
	flag.Parse()

	// auth
	user := AuthRequest{Username: *username, Password: *password}

	// prefix
	prefix := "/api/endpoints/" + *endpointID + "/docker"

	// proxy
	remote, err := url.Parse(*redirecturl)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	director := func(req *http.Request) {
		// Request JWT Token
		authBytes := new(bytes.Buffer)
		json.NewEncoder(authBytes).Encode(user)
		res, _ := http.Post(*redirecturl+"/api/auth", "application/json", authBytes)
		var authResponse AuthResponse
		json.NewDecoder(res.Body).Decode(&authResponse)

		// Add Auth header and prefix
		req.Header.Add("Authorization", "Bearer "+authResponse.JWT)
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = prefix + req.URL.Path
	}
	proxy.Director = director

	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServe(*port, nil)
	if err != nil {
		panic(err)
	}
}

type AuthRequest struct {
	Username string `json:username`
	Password string `json:password`
}
type AuthResponse struct {
	JWT string `json:jwt`
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		p.ServeHTTP(w, r)
	}
}
