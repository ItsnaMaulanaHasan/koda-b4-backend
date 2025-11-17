package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"backend-daily-greens/utils"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
	if page > totalPage && totalPage > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	// hateoas
	links := utils.BuildHateoasPagination(ctx, page, limit, search, totalData)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    transactions,
		"_links":  links,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
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

	// get transaction detail
	transaction, message, err := models.GetDetailTransaction(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "Transaction not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// get transaction items
	transactionItems, message, err := models.GetTransactionItems(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
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
// @Param        statusId       formData  string  true  "Transaction status (1(On Progess), 2(Sending Goods), 3(Finish Order))"
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

	statusId := ctx.PostForm("statusId")
	if statusId == "" {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Status is required",
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

	// check transaction exists
	isExists, err := models.CheckTransactionExists(id)
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

	// update transaction status
	isSuccess, message, err := models.UpdateTransactionStatusById(id, statusId, userId.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: isSuccess,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: message,
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
