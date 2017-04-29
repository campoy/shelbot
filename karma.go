package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type karmaMap struct {
	m    map[string]int
	path string
}

func (k *karmaMap) increment(item string) int {
	k.m[item]++
	return k.m[item]
}

func (k *karmaMap) decrement(item string) int {
	k.m[item]--
	return k.m[item]
}

func (k *karmaMap) query(item string) int {
	return k.m[item]
}

func readKarma(path string) (*karmaMap, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", path, err)
	}

	var m map[string]int
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("could not decode karma from %s: %v", path, err)
	}
	return &karmaMap{m: m, path: path}, nil
}

func (k karmaMap) save() error {
	data, err := json.Marshal(k)
	if err != nil {
		return fmt.Errorf("could not encode karma: %v", err)
	}

	if err := ioutil.WriteFile(k.path, data, 0666); err != nil {
		return fmt.Errorf("could not write to %s: %v", k.path, err)
	}
	return nil
}
