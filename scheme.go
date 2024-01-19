package main

import (
	"math/rand"
	"net/http"
	"time"
)

type DataJsonMap struct {
	Id           int    `json:"id"`
	Value        string `json:"value"`
	LastModified string `json:"last_modified"`
}

func CreateDataJson(value string) DataJsonMap {
	return DataJsonMap{Id: rand.Int(), Value: value, LastModified: time.Now().UTC().Format(http.TimeFormat)}
}
