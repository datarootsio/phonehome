package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

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

	d1 := JsonDate(time.Now().AddDate(0, 0, 1))
	d2 := JsonDate(time.Now().AddDate(0, 0, -2))

	tests := []test{
		{fq: FilterQuery{Key: testKey, Organisation: testOrg, Repository: testRepo}, expectErr: false, expectedLen: 2},
		{fq: FilterQuery{Key: testKey2, Organisation: testOrg, Repository: testRepo}, expectErr: false, expectedLen: 0},
		{fq: FilterQuery{Key: testKey,
			Organisation: testOrg, Repository: testRepo,
			FromDate: &d1},
			expectErr: false, expectedLen: 0},
		{fq: FilterQuery{Key: testKey,
			Organisation: testOrg, Repository: testRepo,
			FromDate: &d2},
			expectErr: false, expectedLen: 2},
		{fq: FilterQuery{Key: testKey,
			Organisation: testOrg, Repository: testRepo,
			FromDate: &d2},
			expectErr: false, expectedLen: 2},
		{fq: FilterQuery{Key: testKey,
			Organisation: testOrg, Repository: testRepo,
			FromDate: &d2, ToDate: &d1},
			expectErr: false, expectedLen: 2},
	}

	for _, test := range tests {
		cc, err := getCountCalls(test.fq)
		assert.Equal(t, test.expectedLen, cc)
		assert.Equal(t, test.expectErr, err != nil)
	}
}

func TestGetOrgRepoHTTP(t *testing.T) {
	router := buildServer()

	testKey := uuid.NewV4().String()
	testVal := uuid.NewV4().String()
	testKey2 := uuid.NewV4().String()
	testVal2 := uuid.NewV4().String()
	testOrg := uuid.NewV4().String()
	testRepo := uuid.NewV4().String()

	// create new call
	m, b := map[string]interface{}{testKey: testVal, testKey2: testVal2}, new(bytes.Buffer)
	json.NewEncoder(b).Encode(m)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/%s", testOrg, testRepo), b)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)

	// check if we can count that testKey
	req, _ = http.NewRequest("GET", fmt.Sprintf("/%s/%s/count", testOrg, testRepo), b)
	q := req.URL.Query()
	q.Add("key", testKey)
	req.URL.RawQuery = q.Encode()

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
	var cs CountResp
	json.NewDecoder(w.Body).Decode(&cs)

	assert.Equal(t, w.Result().StatusCode, 200)
	assert.EqualValues(t, 1, cs.Data) // should count on registered call

	// check if we can get back that call
	req, _ = http.NewRequest("GET", fmt.Sprintf("/%s/%s", testOrg, testRepo), b)
	q = req.URL.Query()
	q.Add("key", testKey)
	req.URL.RawQuery = q.Encode()

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
	var cr CallsResp
	json.NewDecoder(w.Body).Decode(&cr)

	assert.Equal(t, w.Result().StatusCode, 200)
	assert.True(t, len(cr.Data) > 0)
	assert.Equal(t, cr.Data[0].Organisation, testOrg)

	assert.Contains(t, string(cr.Data[0].Payload.RawMessage), testKey)

	// check if get count calls by date works

	// create second call to same org/repo
	m, b = map[string]interface{}{testKey: testVal, testKey2: testVal2}, new(bytes.Buffer)
	json.NewEncoder(b).Encode(m)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/%s/%s", testOrg, testRepo), b)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Result().StatusCode)

	// let fetch the counts per day
	req, _ = http.NewRequest("GET", fmt.Sprintf("/%s/%s/count/daily", testOrg, testRepo), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var dcr DailyCountResp
	json.NewDecoder(w.Body).Decode(&dcr)
	assert.Equal(t, 200, w.Result().StatusCode)
	assert.True(t, len(dcr.Data) > 0)
	assert.Equal(t, 1, len(dcr.Data))
	assert.EqualValues(t, 2, dcr.Data[0].Count)

	// finally lets get a total count for our badge
	req, _ = http.NewRequest("GET", fmt.Sprintf("/%s/%s/count/badge", testOrg, testRepo), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var bdg BadgeInfo
	json.NewDecoder(w.Body).Decode(&bdg)
	assert.Equal(t, 200, w.Result().StatusCode)
	bc, err := strconv.Atoi(bdg.Message)
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, bc, 0)
	assert.True(t, bdg.Label != "")
}
