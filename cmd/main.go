package main

import (
	"context"
	"flag"
	"gateway/config"
	"gateway/internal/bootstrap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "../config.yaml", "path to config file")
	flag.Parse()

	fileConf, err := config.LoadFileConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	envConf, err := config.LoadEnvConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	shutdown := bootstrap.Run(fileConf, envConf)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdown(ctx)
}
