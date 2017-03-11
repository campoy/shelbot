package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type karma struct {
	db     map[string]int
	dbFile io.ReadWriteSeeker
}

func (k *karma) increment(item string) int {
	k.db[item]++
	return k.db[item]
}

func (k *karma) decrement(item string) int {
	k.db[item]--
	return k.db[item]
}

func (k *karma) query(item string) int {
	return k.db[item]
}

func newKarma(d io.ReadWriteSeeker) *karma {
	k := &karma{
		db:     make(map[string]int),
		dbFile: d,
	}

	return k
}

func (k *karma) read() error {
	if _, err := k.dbFile.Seek(io.SeekStart, 0); err != nil {
		return err
	}
	decoder := json.NewDecoder(k.dbFile)
	if err := decoder.Decode(&k.db); err != nil {
		if err != io.EOF {
			return err
		}
	}

	return nil
}

func readKarmaFileJSON(fileLoc string) (*karma, error) {
	var err error
	var dbFile *os.File
	if _, err := os.Stat(fileLoc); err != nil {
		if os.IsNotExist(err) {
			log.Println("No karma JSON found, creating.")
		}
		if dbFile, err = os.OpenFile(fileLoc, os.O_RDWR|os.O_CREATE, 0644); err != nil {
			return nil, err
		}
		return newKarma(dbFile), nil
	}
	log.Println("Loading karma JSON from disk and populating karmaDB map.")
	if dbFile, err = os.OpenFile(fileLoc, os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return nil, err
	}
	k := newKarma(dbFile)
	if err = k.read(); err != nil {
		return nil, err
	}

	return k, nil
}

func (k *karma) save() error {
	marshaledKarmaData, err := json.MarshalIndent(k.db, "", "    ")
	if err != nil {
		return err
	}

	log.Println("Writing karma JSON to file.")
	//ioutil.WriteFile(k.dbFile, marshaledKarmaData, 0644)
	if _, err := k.dbFile.Seek(io.SeekStart, 0); err != nil {
		return err
	}
	if _, err := k.dbFile.Write(marshaledKarmaData); err != nil {
		return err
	}

	return nil
}
