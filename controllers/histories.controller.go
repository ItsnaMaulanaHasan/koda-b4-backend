package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"backend-daily-greens/utils"
	"net/http"
	"strconv"
	"time"

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
// @Param          date           query     string  false  "Date filter (format: YYYY-MM-DD)"  example(2024-01-15)
// @Param          statusid       query     int     false  "Id of status"  default(1)
// @Success        200            {object}  object{success=bool,message=string,data=[]models.History,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int},_links=lib.HateoasLink}  "Successfully retrieved histories list"
// @Failure        400            {object}  lib.ResponseError  "Invalid pagination parameters or page out of range"
// @Failure        401            {object}  lib.ResponseError  "User Id not found in token"
// @Failure        500            {object}  lib.ResponseError  "Internal server error while fetching or processing history data"
// @Router         /histories [get]
func ListHistories(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	date := ctx.Query("date")
	statusId, _ := strconv.Atoi(ctx.DefaultQuery("statusid", "1"))

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

	if date != "" {
		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "Invalid date format. Expected format: YYYY-MM-DD",
			})
			return
		}
	}

	// get user id from token
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// get list histories
	histories, totalData, message, err := models.GetListHistories(userId.(int), page, limit, date, statusId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// calculate total page
	totalPage := (totalData + limit - 1) / limit
	if page > totalPage && totalPage > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	// hateoas
	links := utils.BuildHateoasPagination(ctx, page, limit, totalData)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    histories,
		"_links":  links,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
		},
	})
}

// DetailHistory     godoc
// @Summary          Get detail history
// @Description      Retrieving history detail data based on Id including transaction items
// @Tags             histories
// @Produce          json
// @Security         BearerAuth
// @Param            Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param            noInvoice             path    int     true  "Nomor Invoice"
// @Success          200  {object}  lib.ResponseSuccess{data=models.HistoryDetail}  "Successfully retrieved history detail"
// @Failure          400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure          404  {object}  lib.ResponseError  "history not found"
// @Failure          500  {object}  lib.ResponseError  "Internal server error while fetching history from database"
// @Router           /histories/{noinvoice} [get]
func DetailHistory(ctx *gin.Context) {
	noInovice := ctx.Param("noinvoice")

	historyDetail, message, err := models.GetDetailHistory(noInovice)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "History not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Success get transaction detail",
		Data:    historyDetail,
	})
}
