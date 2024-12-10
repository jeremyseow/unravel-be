package main

import (
	"github.com/jeremyseow/unravel-be/config"
	"github.com/jeremyseow/unravel-be/server"
	"github.com/jeremyseow/unravel-be/storage"
)

func main() {
	cfg, err := config.Load("local")
	if err != nil {
		panic(err)
	}

	allStorages, err := storage.NewAllStorages(cfg)
	if err != nil {
		panic(err)
	}

	s := server.NewServer(cfg, allStorages)
	s.StartServer()
}
