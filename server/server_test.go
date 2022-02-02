package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

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
		{input: Call{Payload: postgres.Jsonb{RawMessage: json.RawMessage(`{"blaat": 0}`)}, Organisation: "testorg", Repository: "testrepo"}, expectErr: false},
		{input: Call{Payload: postgres.Jsonb{RawMessage: json.RawMessage(`invalid`)}, Organisation: "testorg", Repository: "testrepo"}, expectErr: true},
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
	testOrg := uuid.NewV1().String()
	testRepo := uuid.NewV1().String()

	err := registerCall(Call{Organisation: testOrg, Repository: testRepo, Payload: postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "am_sheep"}`, testKey))}})
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{fq: FilterQuery{Key: testKey, Organisation: testOrg, Repository: testRepo}, expectErr: false, expectedLen: 1},
		{fq: FilterQuery{Key: "moot", Organisation: testOrg, Repository: testRepo}, expectErr: false, expectedLen: 0},
	}

	for _, test := range tests {
		cs, err := getCalls(test.fq)
		assert.Equal(t, test.expectedLen, len(cs))
		assert.Equal(t, test.expectErr, err != nil)
	}
}

func TestCountCalls(t *testing.T) {
	type test struct {
		fq          FilterQuery
		expectErr   bool
		expectedLen int64
	}

	testKey := uuid.NewV1().String()
	testKey2 := uuid.NewV1().String()
	testOrg := uuid.NewV1().String()
	testRepo := uuid.NewV1().String()

	// run this twice to know what to expect
	err := registerCall(Call{Organisation: testOrg, Repository: testRepo, Payload: postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "am_sheep"}`, testKey))}})
	if err != nil {
		t.Fatal(err)
	}
	err = registerCall(Call{Organisation: testOrg, Repository: testRepo, Payload: postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "am_sheep"}`, testKey))}})
	if err != nil {
		t.Fatal(err)
	}

	tests := []test{
		{fq: FilterQuery{Key: testKey, Organisation: testOrg, Repository: testRepo}, expectErr: false, expectedLen: 2},
		{fq: FilterQuery{Key: testKey2, Organisation: testOrg, Repository: testRepo}, expectErr: false, expectedLen: 0},
	}

	for _, test := range tests {
		cc, err := getCountCalls(test.fq)
		assert.Equal(t, test.expectedLen, cc)
		assert.Equal(t, test.expectErr, err != nil)
	}
}
