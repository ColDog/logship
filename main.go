package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coldog/logship/pipeline"
)

func handleErr(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	configFile := flag.String("config-file", "logship.json", "Config file path")
	flag.Parse()

	f, err := os.Open(*configFile)
	if err != nil {
		handleErr(err)
	}

	pipe := &pipeline.Pipeline{}

	err = json.NewDecoder(f).Decode(pipe)
	if err != nil {
		handleErr(err)
	}

	err = pipe.Open()
	if err != nil {
		handleErr(err)
	}

	data, _ := json.MarshalIndent(pipe, "", "  ")
	fmt.Println(string(data))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		go func() {
			select {
			case <-c:
			case <-time.After(15 * time.Second):
			}
			os.Exit(1)
		}()
		pipe.Close()
	}()

	pipe.Run()
}
