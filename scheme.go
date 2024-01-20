package main

import (
	"math/rand"
	"net/http"
	"time"
)

type DataJsonMap struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Value        string `json:"value"`
	LastModified string `json:"last_modified"`
}

func CreateDataJson(name, value string) DataJsonMap {
	return DataJsonMap{
		Id:           rand.Int(),
		Name:         name,
		Value:        value,
		LastModified: time.Now().UTC().Format(http.TimeFormat)}
}
