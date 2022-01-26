package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
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
		assert.Equal(t, err != nil, test.expectErr)
	}

}

func TestGetCalls(t *testing.T) {

	type test struct {
		fq          FilterQuery
		expectErr   bool
		expectedLen int
	}

	testKey := uuid.NewV1().String()

	err := registerCall(Call{Payload: postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "am_sheep"}`, testKey))}})
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{fq: FilterQuery{Key: testKey, Organisation: "testorg", Repository: "testrepo"}, expectErr: false, expectedLen: 1},
		{fq: FilterQuery{Key: "moot", Organisation: "testorg", Repository: "testrepo"}, expectErr: false, expectedLen: 0},
	}

	for _, test := range tests {
		cs, err := getCalls(test.fq)
		spew.Dump(cs)
		assert.Equal(t, test.expectedLen, len(cs))
		assert.Equal(t, test.expectErr, err != nil)
	}

}
