package main

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

func main() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	keys, err := redis.Values(conn.Do("KEYS", "turn/secret/*"))
	if err != nil {
		panic(err)
	}
	items, err := conn.Do("MGET", keys...)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", items)
}
