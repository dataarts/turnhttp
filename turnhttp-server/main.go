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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/doug/turnhttp"
)

var (
	port       = flag.String("port", "8080", "port to run on")
	servers    = flag.String("servers", "", "comma seperated list of turn server IPs")
	serversUrl = flag.String("servers-url", "", "json resource returning list of turn server urls")
	domains    = flag.String("domains", "", "comma seperated list of acceptable domains")
	domainsUrl = flag.String("domains-url", "", "json resource returning list of acceptable domains")
	secret     = flag.String("secret", "notasecret", "shared secret to use")
	secretUrl  = flag.String("secret-url", "", "json resource returning shared secret to use")
	rateString = flag.String("rate", "30s", "rate of url updating e.g. 30s or 1m15s")
	rate       time.Duration
	domainList []string
	urls       []string
)

func update(url string, ptr interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(ptr)
	return err
}

func synchronize(url string, ptr interface{}) {
	for {
		update(url, ptr)
		time.Sleep(rate)
	}
}

// run a server
func main() {
	flag.Parse()
	var err error
	rate, err = time.ParseDuration(*rateString)
	if err != nil {
		panic(err)
	}
	for _, ip := range strings.Split(*servers, ",") {
		urls = append(urls, fmt.Sprintf("turn:%s:3478?transport=udp", ip))
		urls = append(urls, fmt.Sprintf("turn:%s:3478?transport=tcp", ip))
		urls = append(urls, fmt.Sprintf("turn:%s:3479?transport=udp", ip))
		urls = append(urls, fmt.Sprintf("turn:%s:3479?transport=tcp", ip))
	}
	domainList = strings.Split(*domains, ",")

	turn := &turnhttp.Service{
		Secret:  *secret,
		Urls:    urls,
		Domains: domainList,
	}

	if *serversUrl != "" {
		go synchronize(*serversUrl, &turn.Uris)
	}
	if *domainsUrl != "" {
		go synchronize(*domainsUrl, &turn.Domains)
	}
	if *secretUrl != "" {
		go synchronize(*secretUrl, &turn.Secret)
	}

	http.Handle("/", turn)

	fmt.Printf("Starting turnhttp on port %v\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
