package main

import (
	"io/ioutil"
	"strings"

	"github.com/gocarina/gocsv"
)

type Airport struct {
	Name      string  `csv:"Airport Name"`
	City      string  `csv:"City"`
	Country   string  `csv:"Country"`
	IATA      string  `csv:"IATA"`
	ICAO      string  `csv:"ICAO"`
	Longitude float64 `csv:"Longitude`
	Latitude  float64 `csv:"Latitude"`
	Altitude  int     `csv:"Altitude"`
}

var Airports []*Airport

func LoadAirports(af string) error {
	data, err := ioutil.ReadFile(af)
	if err != nil {
		return err
	}

	if err = gocsv.UnmarshalBytes(data, &Airports); err != nil {
		return err
	}

	return nil
}

func LookupAirport(q string) *Airport {
	for _, a := range Airports {
		q = strings.ToLower(q)
		if q == strings.ToLower(a.IATA) || q == strings.ToLower(a.ICAO) {
			return a
		}
	}

	return nil
}
