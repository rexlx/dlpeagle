package main

import (
	"log"
	"net/http"
	"os"
)

type Instance struct {
	API     API          `json:"api"`
	Logger  *log.Logger  `json:"-"`
	Gateway *http.Client `json:"-"`
}

type API struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewInstance(api API, logname string) *Instance {
	f, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	logger := log.New(f, "instance: ", log.LstdFlags)
	logger.Println("Creating new instance...")
	return &Instance{
		Logger:  logger,
		API:     api,
		Gateway: http.DefaultClient,
	}
}

func (i *Instance) SendTag(tag Tag) error {
	return nil
}
