// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	port       = flag.String("port", "8080", "port to run on")
	servers    = flag.String("servers", "", "comma seperated list of turn server IPs")
	domains    = flag.String("domains", "", "comma seperated list of acceptable domains")
	secret     = flag.String("secret", "notasecret", "shared secret to use")
	domainList []string
	uris       []string
)

func credentials(username string) (user, pass string) {
	timestamp := time.Now().Unix()
	user = fmt.Sprintf("%d:%s", timestamp, username)
	h := hmac.New(sha1.New, []byte(*secret))
	h.Write([]byte(user))
	pass = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

type TurnResponse struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Uris     []string `json:"uris"`
}

func Handle(rw http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Error(rw, "Must supply a username.", http.StatusBadRequest)
		return
	}

	origin := r.URL.Host
	fmt.Println(origin, domainList)
	for _, domain := range domainList {
		if origin == domain {
			goto ACCEPT
		}
	}
	http.Error(rw, "Invalid host.", http.StatusBadRequest)
	return

ACCEPT:
	user, pass := credentials(username)
	resp := TurnResponse{
		Username: user,
		Password: pass,
		Uris:     uris,
	}

	rw.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	enc.Encode(resp)
}

// run a server
func main() {
	flag.Parse()
	for _, ip := range strings.Split(*servers, ",") {
		uris = append(uris, fmt.Sprintf("turn:%s:3478?transport=udp", ip))
		uris = append(uris, fmt.Sprintf("turn:%s:3478?transport=tcp", ip))
		uris = append(uris, fmt.Sprintf("turn:%s:3479?transport=udp", ip))
		uris = append(uris, fmt.Sprintf("turn:%s:3479?transport=tcp", ip))
	}
	domainList = strings.Split(*domains, ",")

	http.HandleFunc("/", Handle)

	fmt.Printf("Starting turnhttp on port %v\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
