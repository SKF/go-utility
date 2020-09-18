package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	for i := 0; true; i++ {
		now := time.Now()
		if now.Second() == 0 && now.Nanosecond() < 500*1000*1000 || now.Second() == 59 && now.Nanosecond() > 500*1000*1000 {

			start := time.Now()
			resp, err := http.Get("http://localhost:8080/")
			if err != nil {
				fmt.Printf("Error: %v\n", err.Error())
				continue
			}

			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				fmt.Printf("%v: time: %v, code: %d\n", time.Now(), time.Since(start), resp.StatusCode)
			}
		}

		time.Sleep(50 * time.Millisecond)
	}
}
