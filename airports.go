package main

import (
	"io/ioutil"
	"strings"

	"github.com/gocarina/gocsv"
)

type airport struct {
	Name      string  `csv:"Airport Name"`
	City      string  `csv:"City"`
	Country   string  `csv:"Country"`
	IATA      string  `csv:"IATA"`
	ICAO      string  `csv:"ICAO"`
	Longitude float64 `csv:"Longitude"`
	Latitude  float64 `csv:"Latitude"`
	Altitude  int     `csv:"Altitude"`
}

type airports []airport

func (as airports) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return gocsv.UnmarshalBytes(data, &as)
}

func (as airports) Lookup(q string) *airport {
	for _, a := range as {
		q = strings.ToLower(q)
		if q == strings.ToLower(a.IATA) || q == strings.ToLower(a.ICAO) {
			return &a
		}
	}
	return nil
}
