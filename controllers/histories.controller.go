package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListHistories   godoc
// @Summary        Get list histories
// @Description    Retrieving list histories with pagination support
// @Tags           histories
// @Produce        json
// @Security       BearerAuth
// @Param          Authorization  header    string  true   "Bearer token" default(Bearer <token>)
// @Param          page           query     int     false  "Page number"  default(1)  minimum(1)
// @Param          limit          query     int     false  "Number of items per page"  default(5)  minimum(1)  maximum(10)
// @Param          month          query     int     false  "Month filter (1–12)"
// @Param          status         query     int     false  "Id of status"  default(1)
// @Success        200            {object}  object{success=bool,message=string,data=[]models.Transaction,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved transaction list"
// @Failure        400            {object}  lib.ResponseError  "Invalid pagination parameters or page out of range."
// @Failure        500            {object}  lib.ResponseError  "Internal server error while fetching or processing transaction data."
// @Router         /history [get]
func ListHistories(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	month, _ := strconv.Atoi(ctx.Query("month"))
	statusId, _ := strconv.Atoi(ctx.DefaultQuery("status", "1"))

	if page < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'page' must be greater than 0",
		})
		return
	}

	if limit < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' must be greater than 0",
		})
		return
	}

	if limit > 10 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' cannot exceed 10",
		})
		return
	}

	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	histories, totalData, message, err := models.GetListHistories(userId.(int), page, limit, month, statusId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/histories", scheme, host)

	var next any
	var prev any
	switch {
	case totalData == 0:
		page = 0
		next, prev = nil, nil
	case page == 1 && totalPage > 1:
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = nil
	case page == totalPage && totalPage > 1:
		next = nil
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	default:
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	}

	if totalData == 0 {
		page = 0
		next = nil
		prev = nil
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    histories,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
			"next":        next,
			"prev":        prev,
		},
	})
}

func DetailHistoriy(ctx *gin.Context) {

}
