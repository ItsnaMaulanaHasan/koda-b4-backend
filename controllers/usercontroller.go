package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetAllUser(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Page must be greater than 0",
		})
		return
	}

	if limit < 1 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Limit must be greater than 0",
		})
		return
	}

	if limit > 100 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Limit cannot exceed 100",
		})
		return
	}

	var totalData int
	err := config.DB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM users").Scan(&totalData)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to count users",
		})
		return
	}

	offset := (page - 1) * limit
	rows, err := config.DB.Query(context.Background(),
		"SELECT id, first_name, last_name, email, role FROM users LIMIT $1 OFFSET $2",
		limit, offset)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to query users",
		})
		return
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.User])

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to collect users",
		})
		return
	}

	totalPage := (totalData + limit - 1) / limit

	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success get all user",
		"data":    users,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
		},
	})
}
