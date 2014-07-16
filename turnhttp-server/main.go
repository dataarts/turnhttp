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
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/doug/turnhttp"
	"github.com/garyburd/redigo/redis"
)

var (
	port       = flag.String("port", "8080", "port to run on")
	servers    = flag.String("servers", "", "comma seperated list of turn server IPs")
	hosts      = flag.String("hosts", "", "comma seperated list of acceptable hosts")
	secret     = flag.String("secret", "notasecret", "shared secret to use XOR(redis,secret)")
	redisAddr  = flag.String("redis", "", "Redis connection settings XOR(redis,secret)")
	rateString = flag.String("rate", "30s", "rate of updating from redis e.g. 30s or 1m15s")
	ttlString  = flag.String("ttl", "24h", "ttl of credential e.g. 24h33m5s")
	rate       time.Duration
	hostList   []string
	uris       []string
	ttl        time.Duration
	conn       redis.Conn
)

// run a server
func main() {
	flag.Parse()
	var err error
	rate, err = time.ParseDuration(*rateString)
	if err != nil {
		panic(err)
	}
	ttl, err = time.ParseDuration(*ttlString)
	if err != nil {
		panic(err)
	}
	for _, ip := range strings.Split(*servers, ",") {
		uris = append(uris, fmt.Sprintf("turn:%s:3478?transport=udp", ip))
		uris = append(uris, fmt.Sprintf("turn:%s:3478?transport=tcp", ip))
		uris = append(uris, fmt.Sprintf("turn:%s:3479?transport=udp", ip))
		uris = append(uris, fmt.Sprintf("turn:%s:3479?transport=tcp", ip))
	}
	hostList = strings.Split(*hosts, ",")

	// if redis is enabled hash the current day into A, B
	// set based on even or odd day since epoc
	// once a day update
	// serve the current update the old
	if *redisAddr != "" {
		conn, err = redis.Dial("tcp", *redisAddr)
		if err != nil {
			panic(err)
		}
		// inital get setting
	}

	turn := &turnhttp.Service{
		Secret: secret,
		Uris:   uris,
		Hosts:  hostList,
		TTL:    ttl,
	}

	http.Handle("/", turn)

	fmt.Printf("Starting turnhttp on port %v\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
