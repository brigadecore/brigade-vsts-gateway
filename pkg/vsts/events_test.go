package vsts

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	tests := []string{
		"testdata/spec-json-01.json",
		"testdata/spec-json-02.json",
		"testdata/spec-json-03.json",
	}

	for _, file := range tests {
		raw, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		ev := new(Event)
		if err = json.Unmarshal(raw, ev); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewFromRequest(t *testing.T) {
	is := assert.New(t)

	data, err := ioutil.ReadFile("testdata/spec-json-02.json")
	if err != nil {
		t.Fatal(err)
	}

	body := bytes.NewReader(data)

	ev, err := NewFromRequestBody(body)
	if err != nil {
		t.Fatal(err)
	}

	is.Equal("git.push", ev.EventType)
	is.Equal("tfs", ev.PublisherID)
}
