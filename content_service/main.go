package main

import (
  "content_service/conf"
  "content_service/env"
  "content_service/server"
)

func main() {
  conf := conf.New()
  env.InitLog(conf)
  env := env.New(conf)
  server.Start(env)
}
