package main

import (
	"api"
	"log"
)

func main() {
	if err := api.Run(); err != nil {
		log.Fatalf("run api server: %v\n", err)
	}
}


