package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	resp.Query = fq

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
	resp.Query = fq

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

	resp.Query = fq

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
// @Description  Requires a JSON body in the shape of `{"foo": "bar", "coffee": "beans"}`.
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
	var pl CallPayload
	var resp RegisterResp

	// unmarshal and remarshal to strip away nested objects
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	if err := json.Unmarshal(buf.Bytes(), &pl); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	plClean, err := json.Marshal(pl)
	if err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	call.Payload.RawMessage = json.RawMessage(plClean)

	if err := c.ShouldBind(&call); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := c.ShouldBindUri(&or); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	call.Organisation = or.Organisation
	call.Repository = or.Repository

	if err := registerCall(call); err != nil {
		resp.Error = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(200, resp)
}

func registerCall(c Call) error {
	if !json.Valid(c.Payload.RawMessage) {
		return fmt.Errorf("%s is invalid JSON", c.Payload.RawMessage)
	}

	c.Timestamp = time.Now()

	result := db.Create(&c)
	return result.Error
}

func githubRepoExistsMW(c *gin.Context) {
	or := OrgRepoURI{}
	if err := c.ShouldBindUri(or); err != nil {
		// assume we're on a route that doesnt need
		// repo specification
		return
	}

	if !githubRepoExists(or.Organisation, or.Repository) {
		resp := DefaultResp{
			Error: fmt.Sprintf("github repository existence could not be verified (%s/%s)",
				or.Organisation, or.Repository),
		}
		c.JSON(http.StatusBadRequest, resp)
		return
	}
}
