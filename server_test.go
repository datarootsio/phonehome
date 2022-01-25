package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	InitConfig()
	if err := InitDBConn(); err != nil {
		fmt.Print("cannot connect to db")
		os.Exit(1)
	}
	code := m.Run()
	os.Exit(code)
}

func TestRegisterCall(t *testing.T) {
	type test struct {
		input     Call
		expectErr bool
	}

	tests := []test{
		{input: Call{Timestamp: time.Now(), Payload: postgres.Jsonb{RawMessage: json.RawMessage(`{"blaat": 0}`)}, Organisation: "testorg", Repository: "testrepo"}, expectErr: false},
		{input: Call{Timestamp: time.Now(), Payload: postgres.Jsonb{RawMessage: json.RawMessage(`invalid`)}, Organisation: "testorg", Repository: "testrepo"}, expectErr: true},
	}

	for _, test := range tests {
		err := registerCall(test.input)
		spew.Dump(err)
		assert.Equal(t, err != nil, test.expectErr)
	}

}
