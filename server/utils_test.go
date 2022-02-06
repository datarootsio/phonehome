package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGHRepoChecker(t *testing.T) {
	assert.True(t, githubRepoExists("datarootsio", "cheek"))
	assert.False(t, githubRepoExists("datarootsi000o", "cheek"))
}

func TestPayloadStripper(t *testing.T) {
	type test struct {
		input []byte
		refac []byte
	}

	tests := []test{
		// json.Unmarshal(input, )
		{input: []byte(`{"key":"value"}`), refac: []byte(`{"key":"value"}`)},
		{input: []byte(`{"key":"value", "foo": {"bar": 3}}`), refac: []byte(`{"key":"value"}`)},
		{input: []byte(`{"key":{"a":3}}`), refac: []byte(`{}`)},
		{input: []byte(`{"key":3}`), refac: []byte(`{"key":3}`)},
	}

	for _, test := range tests {
		var got CallPayload
		var want CallPayload

		if err := json.Unmarshal(test.input, &got); err != nil {
			t.Error(err)
		}

		if err := json.Unmarshal(test.refac, &want); err != nil {
			t.Error(err)
		}

		got_stripped, _ := payloadStripper(got)

		assert.Equal(t, got_stripped, want)
	}
}
