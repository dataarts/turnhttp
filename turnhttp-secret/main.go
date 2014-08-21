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
	"crypto/rand"
	"flag"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	redisAddr = flag.String("redis", "", "Redis connection settings.")
	ttlString = flag.String("ttl", "24h", "Rate of url updating e.g. 30s or 1m15s")
	ttl       time.Duration
	conn      redis.Conn
)

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func updateSecret() {
	for {
		secret := randString(15)
		key := fmt.Sprintf("turn/secret/%d", time.Now().Unix())
		expire := (ttl * 2).Seconds()
		_, err := conn.Do("SETEX", key, expire, secret)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(ttl)
	}
}

func main() {
	flag.Parse()
	var err error
	ttl, err = time.ParseDuration(*ttlString)
	if err != nil {
		panic(err)
	}
	conn, err = redis.Dial("tcp", *redisAddr)
	if err != nil {
		panic(err)
	}
	updateSecret()
}
