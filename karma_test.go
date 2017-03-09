package main

import (
	"testing"

	"strings"

	"github.com/traherom/memstream"
)

var testConf = `{
	"test1" :100,
	"test2" :200,
	"test3" :300
}`

func TestLoadKarma(t *testing.T) {
	m := memstream.New()
	m.Write([]byte(testConf))

	k := newKarma(m)
	if err := k.read(); err != nil {
		t.Fatal(err)
	}
	if k.db["test1"] != 100 {
		t.Fatal("test1 should have value 100")
	}
	if k.db["test2"] != 200 {
		t.Fatal("test1 should have value 100")
	}
	if k.db["test3"] != 300 {
		t.Fatal("test1 should have value 100")
	}
}

func TestWriteKarma(t *testing.T) {
	m := memstream.New()
	k := newKarma(m)

	k.db["test1"] = 11

	k.save()
	data := string(k.dbFile.(*memstream.MemoryStream).Bytes())
	if !strings.Contains(data, `"test1": 11`) {
		t.Fatalf("Saved output was incorrect, should have 'test1' entry with value 11;\n%s\n", data)
	}
}
