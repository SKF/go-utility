package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	const timeBetweenGets = 50 * time.Millisecond

	for i := 0; true; i++ {
		start := time.Now()

		resp, err := http.Get("http://localhost:8080/")
		if err != nil {
			fmt.Printf("Error: %v\n", err.Error())
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("%v: time: %v, code: %d\n", time.Now(), time.Since(start), resp.StatusCode)
		}

		time.Sleep(timeBetweenGets)
	}
}
