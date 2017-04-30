package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type karma map[string]int

func (k karma) increment(item string) int {
	k[item]++
	return k[item]
}

func (k karma) decrement(item string) int {
	k[item]--
	return k[item]
}

func (k karma) query(item string) int {
	return k[item]
}

func (k *karma) load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", path, err)
	}

	if err := json.Unmarshal(data, k); err != nil {
		return fmt.Errorf("could not decode karma from %s: %v", path, err)
	}
	return nil
}

func (k *karma) write(path string) error {
	data, err := json.Marshal(k)
	if err != nil {
		return fmt.Errorf("could not encode karma: %v", err)
	}

	if err := ioutil.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("could not write to %s: %v", path, err)
	}
	return nil
}
