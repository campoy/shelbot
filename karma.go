package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type karmaMap map[string]int

func (k karmaMap) increment(item string) int {
	k[item]++
	return k[item]
}

func (k karmaMap) decrement(item string) int {
	k[item]--
	return k[item]
}

func (k karmaMap) query(item string) int {
	return k[item]
}

func readKarma(path string) (karmaMap, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", path, err)
	}

	var k karmaMap
	if err := json.Unmarshal(data, &k); err != nil {
		return nil, fmt.Errorf("could not decode karma from %s: %v", path, err)
	}
	return k, nil
}

func writeKarma(path string, k karmaMap) error {
	data, err := json.Marshal(k)
	if err != nil {
		return fmt.Errorf("could not encode karma: %v", err)
	}

	if err := ioutil.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("could not write to %s: %v", path, err)
	}
	return nil
}
