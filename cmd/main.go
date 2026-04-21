package main

import (
	"fmt"

	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http"
	"github.com/jeremyseow/unravel-be/application/adapter/persistence/postgres"
	"github.com/jeremyseow/unravel-be/application/usecase"
	"github.com/jeremyseow/unravel-be/config"
)

func main() {
	cfg, err := config.Load("local")
	if err != nil {
		panic(err)
	}

	allStorages, err := postgres.NewAllStorages(cfg)
	if err != nil {
		fmt.Println(err)
		// panic(err)
	}

	allServices := usecase.NewAllServices(allStorages)

	s := http.NewServer(cfg, allServices)
	s.StartServer()
}
