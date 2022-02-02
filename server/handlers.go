package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
)

func getCountCalls(fq FilterQuery) (int64, error) {
	var count int64

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return count, nil
	}

	result := gq.Model(&Call{}).Count(&count)
	if result.Error != nil {
		return count, result.Error
	}

	return count, nil
}

func getCountCallsByDate(fq FilterQuery) (DayCounts, error) {
	dc := DayCounts{}

	res := db.Raw(`
	select timestamp::date as date, count(*) 
	from calls
	group by timestamp::date
	order by timestamp::date ASC
	`).Scan(&dc)
	if res.Error != nil {
		return dc, res.Error
	}

	return dc, nil
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bi := BadgeInfo{}.Get(count)

	c.JSON(200, bi)

}

func getCountCallsHandler(c *gin.Context) {
	var fq FilterQuery
	var or OrgRepoURI
	resp := gin.H{}

	c.ShouldBind(&fq)
	if err := c.ShouldBindUri(&or); err != nil {
		resp["error"] = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return

	}

	fq.Organisation = or.Organisation
	fq.Repository = or.Repository
	resp["query"] = fq

	switch fq.GroupBy {
	case "":
		count, err := getCountCalls(fq)
		if err != nil {
			resp["error"] = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		resp["data"] = count
		c.JSON(200, resp)
	case "day":
		spew.Dump("day")
		dc, err := getCountCallsByDate(fq)
		if err != nil {
			resp["error"] = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		resp["data"] = dc
		c.JSON(200, resp)
	default:
		resp["error"] = fmt.Sprintf(`group_by key '%s' not supported`, fq.GroupBy)
		c.JSON(http.StatusBadRequest, resp)
	}
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

// @Summary Fetch telemetry calls with filtering ability.
// @Schemes FilterQuery
// @Description Fetch calls.
// @Accept json
// @Produce json
// @Success 200 {array} Call
// @Router /:organisation/:repository [get]
func getCallsHandler(c *gin.Context) {
	var fq FilterQuery
	var or OrgRepoURI

	resp := gin.H{}

	c.ShouldBind(&fq)
	if err := c.ShouldBindUri(&or); err != nil {
		resp["error"] = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return

	}

	fq.Organisation = or.Organisation
	fq.Repository = or.Repository

	resp["query"] = fq

	cs, err := getCalls(fq)
	if err != nil {
		resp["error"] = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp["data"] = cs
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
