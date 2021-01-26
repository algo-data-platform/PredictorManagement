package main

import (
	"log"
	"server/env"
	"server/server"
)

func main() {
	log.Println("server begin to work....")
	env.New()
	server.Start()
}
