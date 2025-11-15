package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// ListTransactions  godoc
// @Summary          Get list transactions
// @Description  	 Retrieving list transactions with pagination support and search
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
func ListTransactions(ctx *gin.Context) {
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

	// get total data transactions
	totalData, err := models.GetTotalDataTransactions(search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total transactions in database",
			Error:   err.Error(),
		})
		return
	}

	// get list all transactions
	transactions, message, err := models.GetListAllTransactions(page, limit, search)
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
	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	// hateoas
	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/admin/transactions", scheme, host)

	var next any
	var prev any

	if totalData == 0 {
		page = 0
		next = nil
		prev = nil
	} else if page == 1 && totalPage > 1 {
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = nil
	} else if page == totalPage && totalPage > 1 {
		next = nil
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	} else if totalPage > 1 {
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	} else {
		next = nil
		prev = nil
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
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

// TransactionDetail godoc
// @Summary          Get transaction by Id
// @Description      Retrieving transaction detail data based on Id including transaction items
// @Tags             admin/transactions
// @Accept           x-www-form-urlencoded
// @Produce          json
// @Security         BearerAuth
// @Param            Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param            id             path    int     true  "Transaction Id"
// @Success          200  {object}  lib.ResponseSuccess{data=models.TransactionDetail}  "Successfully retrieved transaction detail"
// @Failure          400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure          404  {object}  lib.ResponseError  "Transaction not found"
// @Failure          500  {object}  lib.ResponseError  "Internal server error while fetching transaction from database"
// @Router           /admin/transactions/{id} [get]
func DetailTransactions(ctx *gin.Context) {
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
			t.no_invoice,
			t.date_transaction,
			t.full_name,
			t.email,
			t.address,
			t.phone,
			pm.name AS payment_method,
			om.name AS order_method,
			s.name AS status,
			t.delivery_fee,
			t.admin_fee,
			t.tax,
			t.total_transaction
		FROM 
			transactions t
		JOIN 
			payment_methods pm ON t.payment_method_id = pm.id
		JOIN
			order_methods om ON t.order_method_id = pm.id
		JOIN 
			status s ON t.status_id = s.id
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
			transaction_id,
			product_id,
			product_name,
			product_price,
			discount_percent,
			discount_price,
			size,
			size_cost,
			variant,
			variant_cost,
			amount,
			subtotal
		FROM transaction_items
		WHERE transaction_id = $1
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

	transactionItems, err := pgx.CollectRows(productRows, pgx.RowToStructByName[models.TransactionItems])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process ordered products data",
			Error:   err.Error(),
		})
		return
	}

	transaction.TransactionItems = transactionItems

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

	statusId := ctx.PostForm("status_id")
	if statusId == "" {
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

	if !isExists {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Transaction not found",
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		`UPDATE transactions 
		 SET status_id  = $1,
		     updated_by = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		statusId,
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

// Checkout      godoc
// @Summary      Checkout carts
// @Description  Checkout products on the cart
// @Tags         transactions
// @Accept       application/json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param        DataCheckout   body      models.TransactionRequest  true  "Data Checkout"
// @Success      201  {object}  lib.ResponseSuccess{data=models.TransactionDetail}  "Transaction created successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body"
// @Failure      401  {object}  lib.ResponseError  "User Id not found in token"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while acces database"
// @Router       /transactions [post]
func Checkout(ctx *gin.Context) {
	var bodyCheckout models.TransactionRequest
	err := ctx.ShouldBindJSON(&bodyCheckout)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid JSON body",
			Error:   err.Error(),
		})
		return
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

	// get user profile data based on user Id from token
	user, message, err := models.GetDetailUser(userId.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// cek payment info
	if bodyCheckout.FullName == "" {
		bodyCheckout.FullName = user.FullName
	}
	if bodyCheckout.Email == "" {
		bodyCheckout.Email = user.Email
	}
	if bodyCheckout.Address == "" {
		bodyCheckout.Address = user.Address
	}
	if bodyCheckout.Phone == "" {
		bodyCheckout.Phone = user.Phone
	}

	if strings.TrimSpace(bodyCheckout.FullName) == "" ||
		strings.TrimSpace(bodyCheckout.Email) == "" ||
		strings.TrimSpace(bodyCheckout.Address) == "" ||
		strings.TrimSpace(bodyCheckout.Phone) == "" {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Payment info is incomplete. Please update your profile or provide data in the request body",
		})
		return
	}

	bodyCheckout.DeliveryFee, bodyCheckout.AdminFee, message, err = models.GetDeliveryFeeAndAdminFee(bodyCheckout.OrderMethodId, bodyCheckout.PaymentMethodId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// get list cart by user id from token
	carts, message, err := models.GetListCart(userId.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	if len(carts) == 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Cart is empty, cannot checkout",
		})
		return
	}

	// calculate total transaction
	var total float64
	for _, c := range carts {
		total += c.Subtotal
	}
	bodyCheckout.Tax = total * 0.10
	bodyCheckout.TotalTransaction = total + bodyCheckout.Tax + bodyCheckout.DeliveryFee + bodyCheckout.AdminFee

	// get date transaction
	bodyCheckout.DateTransaction = time.Now()

	// generate invoice
	dateStr := time.Now().Format("20060102")
	randomNum := rand.Intn(99999)
	bodyCheckout.NoInvoice = fmt.Sprintf("INV-%s-%05d", dateStr, randomNum)

	// insert data to transactions
	transactionId, message, err := models.MakeTransaction(userId.(int), bodyCheckout, carts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: message,
		Data: gin.H{
			"transactionId":    transactionId,
			"noInvoice":        bodyCheckout.NoInvoice,
			"dateTransaction":  bodyCheckout.DateTransaction,
			"deliveryFee":      bodyCheckout.DeliveryFee,
			"adminFee":         bodyCheckout.AdminFee,
			"tax":              bodyCheckout.Tax,
			"totalTransaction": bodyCheckout.TotalTransaction,
		},
	})
}
