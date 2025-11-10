package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// GetAllTransaction godoc
// @Summary          Get all transaction
// @Description  	 Retrieving all transaction data with pagination support
// @Tags         	 transactions
// @Produce      	 json
// @Security     	 BearerAuth
// @Param        	 Authorization  header    string  true   "Bearer token" default(Bearer <token>)
// @Param        	 page           query     int     false  "Page number"  default(1)  minimum(1)
// @Param        	 limit          query     int     false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        	 search         query     string  false  "Search value"
// @Success      	 200            {object}  object{success=bool,message=string,data=[]models.Transaction,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved transaction list"
// @Failure      	 400            {object}  lib.ResponseError  "Invalid pagination parameters or page out of range."
// @Failure      	 500            {object}  lib.ResponseError  "Internal server error while fetching or processing transaction data."
// @Router       	 /admin/transactions [get]
func GetAllTransaction(ctx *gin.Context) {
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

	var totalData int
	var err error
	searchParam := "%" + search + "%"

	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) 
			 FROM transactions
			 WHERE no_order ILIKE $1`, searchParam).Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM transactions`).Scan(&totalData)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total transactions in database",
			Error:   err.Error(),
		})
		return
	}

	offset := (page - 1) * limit
	var rows pgx.Rows

	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				id,
				no_order,
				date_order,
				status,
				total_transaction
			FROM transactions
			WHERE no_order ILIKE $3
			ORDER BY date_order DESC, id DESC
			LIMIT $1 OFFSET $2`, limit, offset, searchParam)
	} else {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				id,
				no_order,
				date_order,
				status,
				total_transaction
			FROM transactions
			ORDER BY date_order DESC, id DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch transactions from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Transaction])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process transaction data from database",
			Error:   err.Error(),
		})
		return
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
	baseURL := fmt.Sprintf("%s://%s/admin/transactions", scheme, host)

	var next any
	var prev any
	switch page {
	case 1:
		if page < totalPage {
			next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		} else {
			next = nil
		}
		prev = nil
	case totalPage:
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
		"message": "Success get all transaction",
		"data":    transactions,
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
