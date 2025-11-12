package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// GetAllTransaction godoc
// @Summary          Get all transaction
// @Description  	 Retrieving all transaction data with pagination support
// @Tags         	 admin/transactions
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

// GetTransactionById godoc
// @Summary      Get transaction by Id
// @Description  Retrieving transaction detail data based on Id including ordered products
// @Tags         admin/transactions
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Transaction Id"
// @Success      200  {object}  lib.ResponseSuccess{data=models.TransactionDetail}  "Successfully retrieved transaction detail"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Transaction not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while fetching transaction from database"
// @Router       /admin/transactions/{id} [get]
func GetTransactionById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			t.id,
			t.user_id,
			t.no_order,
			t.date_order,
			t.full_name,
			t.email,
			t.address,
			t.phone,
			COALESCE(t.payment_method, '') AS payment_method,
			t.shipping,
			t.status,
			t.total_transaction,
			COALESCE(t.delivery_fee, 0) AS delivery_fee,
			COALESCE(t.tax, 0) AS tax
		FROM transactions t
		WHERE t.id = $1`, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch transaction from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	transaction, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.TransactionDetail])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "Transaction not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process transaction data",
			Error:   err.Error(),
		})
		return
	}

	productRows, err := config.DB.Query(context.Background(),
		`SELECT 
			id,
			product_id,
			product_name,
			product_price,
			COALESCE(discount_percent, 0) AS discount_percent,
			amount,
			subtotal,
			size
		FROM ordered_products
		WHERE order_id = $1
		ORDER BY id ASC`, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch ordered products from database",
			Error:   err.Error(),
		})
		return
	}
	defer productRows.Close()

	orderedProducts, err := pgx.CollectRows(productRows, pgx.RowToStructByName[models.OrderedProduct])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process ordered products data",
			Error:   err.Error(),
		})
		return
	}

	transaction.OrderedProducts = orderedProducts

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Success get transaction detail",
		Data:    transaction,
	})
}

// UpdateTransactionStatus godoc
// @Summary      Update transaction status
// @Description  Updating transaction status based on Id
// @Tags         admin/transactions
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path      int     true  "Transaction Id"
// @Param        status         formData  string  true  "Transaction status (On Progress, Sending, Finished, Cancelled)"
// @Success      200  {object}  lib.ResponseSuccess  "Transaction status updated successfully"
// @Failure      400  {object}  lib.ResponseError   "Invalid Id format or invalid request body"
// @Failure      404  {object}  lib.ResponseError   "Transaction not found"
// @Failure      500  {object}  lib.ResponseError   "Internal server error while updating transaction status"
// @Router       /admin/transactions/{id} [patch]
func UpdateTransactionStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	status := ctx.PostForm("status")
	if status == "" {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Status is required",
		})
		return
	}

	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	var isExists bool
	err = config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM transactions WHERE id = $1)", id,
	).Scan(&isExists)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking transaction existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Transaction not found",
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		`UPDATE transactions 
		 SET status     = $1,
		     updated_by = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		status,
		userIdFromToken,
		id,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating transaction status",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Transaction status updated successfully",
	})
}
