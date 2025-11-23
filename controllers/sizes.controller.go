package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"backend-daily-greens/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListSizes         godoc
// @Summary      	 Get list sizes
// @Description  	 Retrieving list sizes data with pagination support
// @Tags         	 /sizes
// @Produce      	 json
// @Param        	 page           query     int     false  "Page number"               default(1)   minimum(1)
// @Param        	 limit          query     int     false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        	 search         query     string  false  "Search value"
// @Success      	 200  {object}  object{success=bool,message=string,data=[]models.History,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int},_links=lib.HateoasLink}  "Successfully retrieved sizes list"
// @Failure      	 400  {object}  lib.ResponseError  "Invalid pagination parameters or page out of range"
// @Failure      	 500  {object}  lib.ResponseError  "Internal server error while fetching or processing variant data"
// @Router       	 /sizes [get]
func ListSizes(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")

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

	if limit > 100 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' cannot exceed 100",
		})
		return
	}

	// get total data sizes
	totalData, err := models.GetTotalDataSizes(search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total sizes in database",
			Error:   err.Error(),
		})
		return
	}

	// get list all sizes
	sizes, message, err := models.GetListAllSizes(page, limit, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// get total page
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
		"data":    sizes,
		"_links":  links,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
		},
	})
}
