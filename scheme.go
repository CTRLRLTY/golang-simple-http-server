package main

import (
	"math/rand"
)

type DataJsonMap struct {
	Id    int    `json:"id"`
	Value string `json:"value"`
}

func CreateDataJson(value string) DataJsonMap {
	return DataJsonMap{Id: rand.Int(), Value: value}
}
