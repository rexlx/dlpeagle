package main

import "net/http"

type Instance struct {
	API     API          `json:"api"`
	Gateway *http.Client `json:"-"`
}

type API struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}
