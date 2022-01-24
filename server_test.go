package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm/dialects/postgres"
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
	tests := []Call{
		{Timestamp: time.Now(), Payload: postgres.Jsonb{RawMessage: json.RawMessage(`{"blaat": 0}`)}, Organisation: "testorg", Repository: "testrepo"},
	}

	for _, test := range tests {
		err := registerCall(test)
		spew.Dump(err)

	}

}
