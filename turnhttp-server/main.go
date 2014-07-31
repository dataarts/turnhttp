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
	"sort"
	"strings"
	"time"

	"github.com/doug/turnhttp"
	"github.com/garyburd/redigo/redis"
)

var (
	port       = flag.String("port", "8080", "port to run on")
	servers    = flag.String("servers", "", "comma seperated list of turn server IPs")
	hosts      = flag.String("hosts", "", "comma seperated list of acceptable hosts")
	secret     = flag.String("secret", "", "shared secret to use")
	redisAddr  = flag.String("redis", "", "Redis connection settings, if secret or hosts is not provided it will try and fetch it from redis with KEYS 'turn/secret/*' and SMEMBERS 'turn/hosts'.")
	ttlString  = flag.String("ttl", "24h", "ttl of credential e.g. 24h33m5s")
	rateString = flag.String("rate", "5m", "Rate at which to pole the redis server.")

	turn     *turnhttp.Service
	conn     redis.Conn
	hostList []string
	uris     []string
	ttl      time.Duration
	rate     time.Duration
)

func updateSecret() {
	values, err := redis.Values(conn.Do("KEYS", "turn/secret/*"))
	if err != nil {
		panic(err)
	}
	var keys []string
	if err := redis.ScanSlice(values, &keys); err != nil {
		panic(err)
	}

	if len(keys) == 0 {
		return
	}

	sort.Sort(sort.StringSlice(keys))
	key := keys[0]
	item, err := redis.String(conn.Do("GET", key))
	if err != nil {
		panic(err)
	}
	turn.Secret = item
}

func updateHosts() {
	values, err := redis.Values(conn.Do("SMEMBERS", "turn/hosts"))
	if err != nil {
		panic(err)
	}
	var hosts []string
	if err := redis.ScanSlice(values, &hosts); err != nil {
		panic(err)
	}
	turn.Hosts = hosts
}

// run a server
func main() {
	flag.Parse()
	var err error
	ttl, err = time.ParseDuration(*ttlString)
	if err != nil {
		panic(err)
	}
	rate, err = time.ParseDuration(*rateString)
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
		go func() {
			for {
				// listen for changes and update secret
				if *secret == "" {
					updateSecret()
				}
				if *hosts == "" {
					updateHosts()
				}
				time.Sleep(rate)
			}
		}()
	}

	turn = &turnhttp.Service{
		Secret: *secret,
		Uris:   uris,
		Hosts:  hostList,
		TTL:    ttl,
	}

	http.Handle("/", turn)

	fmt.Printf("Starting turnhttp on port %v\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
