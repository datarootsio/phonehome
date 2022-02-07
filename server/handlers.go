package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
)

func getCountCalls(fq FilterQuery) (int64, error) {
	var count int64

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return count, err
	}

	result := gq.Model(&Call{}).Count(&count)
	if result.Error != nil {
		return count, result.Error
	}

	return count, nil
}

func getCountCallsByDate(fq FilterQuery) (DayCounts, error) {
	dc := DayCounts{}

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return dc, err
	}

	res := gq.Model(&Call{}).
		Group("timestamp::date").
		Select("timestamp::date as date, count(*) as count").
		Find(&dc)
	if res.Error != nil {
		return dc, res.Error
	}

	// perhaps unmarshal to actual time.Time first before moving to string?
	for i, v := range dc {
		dc[i].Date = strings.Split(v.Date, "T")[0]
	}

	return dc, nil
}

// @Summary      shield.io badge information.
// @Description  Will give back a full count of telemetry calls.
// @Description  Check out the documentation at [shields.io](https://shields.io/endpoint) for more details.
// @Param        organisation  path   string  true   "github organisation"
// @Param        repository    path   string  true   "repository name"
// @Produce      json
// @Success      200  {object}  BadgeInfo
// @Router       /{organisation}/{repository}/count/badge [get]
func getCountCallsBadgeHandler(c *gin.Context) {
	var or OrgRepoURI
	if err := c.ShouldBindUri(&or); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fq := FilterQuery{}
	fq.AddOrgRepo(or)

	count, err := getCountCalls(fq)
	if err != nil {
		resp := DefaultResp{Error: err.Error()}
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	bi := BadgeInfo{}.Create(count)

	c.JSON(200, bi)
}

// @Summary      Count telemetry calls.
// @Description  Count telemetry calls with optional filtering.
// @Param        organisation  path   string  true   "github organisation"
// @Param        repository    path   string  true   "repository name"
// @Param        key           query  string  false  "filter by key passed in POST payload"
// @Param        from_date     query  string  false  "from date to filter on"
// @Param        to_date       query  string  false  "to date to filter on"
// @Produce      json
// @Success      200  {array}  Call
// @Router       /{organisation}/{repository}/count [get]
func getCountCallsHandler(c *gin.Context) {
	var fq FilterQuery
	var or OrgRepoURI
	resp := CountResp{}

	c.ShouldBind(&fq)
	if err := c.ShouldBindUri(&or); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return

	}

	fq.Organisation = or.Organisation
	fq.Repository = or.Repository
	resp.Query = &fq

	count, err := getCountCalls(fq)
	if err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	resp.Data = count
	c.JSON(200, resp)
}

// @Summary      Count telemetry calls grouped by date.
// @Description  Count telemetry calls with optional filtering.
// @Param        organisation  path   string  true   "github organisation"
// @Param        repository    path   string  true   "repository name"
// @Param        key           query  string  false  "filter by key passed in POST payload"
// @Param        from_date     query  string  false  "from date to filter on"
// @Param        to_date       query  string  false  "to date to filter on"
// @Produce      json
// @Success      200  {array}  Call
// @Router       /{organisation}/{repository}/count/daily [get]
func getCountCallsByDayHandler(c *gin.Context) {
	var fq FilterQuery
	var or OrgRepoURI
	resp := DailyCountResp{}

	c.ShouldBind(&fq)
	if err := c.ShouldBindUri(&or); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return

	}

	fq.Organisation = or.Organisation
	fq.Repository = or.Repository
	resp.Query = &fq

	dc, err := getCountCallsByDate(fq)
	if err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	resp.Data = dc
	c.JSON(200, resp)
}

func getCalls(fq FilterQuery) ([]Call, error) {
	var calls []Call

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return calls, err
	}

	r := gq.Find(&calls).Limit(getCallsLimit)
	return calls, r.Error
}

// @Summary      Fetch telemetry calls.
// @Description  Fetch telemetry calls with optional filtering.
// @Param        organisation  path   string  true   "github organisation"
// @Param        repository    path   string  true   "repository name"
// @Param        key           query  string  false  "filter by key passed in POST payload"
// @Param        from_date     query  string  false  "from date to filter on"
// @Param        to_date       query  string  false  "to date to filter on"
// @Produce      json
// @Success      200  {object}  CallsResp
// @Router       /{organisation}/{repository} [get]
func getCallsHandler(c *gin.Context) {
	var fq FilterQuery
	var or OrgRepoURI

	resp := CallsResp{}

	c.ShouldBind(&fq)
	if err := c.ShouldBindUri(&or); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	fq.Organisation = or.Organisation
	fq.Repository = or.Repository

	resp.Query = &fq

	cs, err := getCalls(fq)
	if err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp.Data = cs
	c.JSON(200, resp)
}

// @Summary      Register new telemetry call.
// @Description  Register new call.
// @Description
// @Description  Requires a JSON body in the shape of `{"foo": "bar", "coffee": 432}`.
// @Description  Expects either an empty object `{}` or an object that only contains keys and **unnested** values.
// @Description  Nested objects will be stripped from the payload and a warning message will be returned.
// @Accept json
// @Param        organisation  path   string  true   "github organisation"
// @Param        repository    path   string  true   "repository name"
// @Param        repository    path   string  true   "repository name"
// @Produce      json
// @Success      200  {object}  RegisterResp
// @Router       /{organisation}/{repository} [post]
func registerCallHander(c *gin.Context) {
	var or OrgRepoURI
	var call Call
	var resp RegisterResp

	// read json payload in body
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	call.Payload.RawMessage = json.RawMessage(buf.Bytes())

	if err := c.ShouldBindUri(&or); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	call.Organisation = or.Organisation
	call.Repository = or.Repository
	originSha := sha256.Sum256([]byte(c.ClientIP()))
	call.Origin = hex.EncodeToString(originSha[:])

	cpl, stripped, err := registerCall(call)
	if err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	payloadClean, err := json.Marshal(cpl)
	if err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp.Payload = payloadClean
	if stripped {
		resp.Message = "WARN: payload got stripped of non-allowed content"
	}
	spew.Dump(call)

	c.JSON(200, resp)
}

func registerCall(c Call) (CallPayload, bool, error) {
	var pl CallPayload
	var stripped bool

	// still put valid value in jsonb col if body empoty
	if reflect.DeepEqual(c.Payload.RawMessage, json.RawMessage{}) {
		c.Payload.RawMessage = []byte(`{}`)
	}

	// check for validity
	// this err might also be raised by unmarshalling
	// to be checked
	if !json.Valid(c.Payload.RawMessage) {
		return pl, stripped, fmt.Errorf("'%s' is invalid JSON", c.Payload.RawMessage)
	}

	// make sure that no nested objects are passed
	err := json.Unmarshal(c.Payload.RawMessage, &pl)
	if err != nil {
		return pl, stripped, err
	}

	// strip out unwanted stuff
	pl, stripped = payloadStripper(pl)
	c.Timestamp = time.Now()

	result := db.Create(&c)
	return pl, stripped, result.Error
}

func githubRepoExistsMW(c *gin.Context) {
	if !checkRepoExistence {
		c.Next()
		return
	}

	org := c.Param("organisation")
	repo := c.Param("repository")

	if !githubRepoExists(org, repo) {
		resp := DefaultResp{
			Error: fmt.Sprintf("github repository doesn't seem to exist: %s/%s",
				org, repo),
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, resp)
	}
	c.Next()
}
