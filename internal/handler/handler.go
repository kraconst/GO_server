package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kraconst/GO_server/internal/domain"
)

type Handler struct{}

func New() *Handler { return &Handler{} }

var mockAccounts = []domain.Account{
	{ID: 1, Phone: "+79991112233", Username: "@sales_anna", EmployeeName: "Анна", Status: domain.StatusActive, CreatedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Phone: "+79994445566", Username: "@support_oleg", EmployeeName: "Олег", Status: domain.StatusDisconnected, CreatedAt: time.Date(2026, 9, 15, 14, 30, 0, 0, time.UTC)},
}

type createAccountRequest struct {
	Phone        string `json:"phone" binding:"required"`
	Username     string `json:"username" binding:"required"`
	EmployeeName string `json:"employee_name" binding:"required"`
}

func (h *Handler) Run(addr string) error {
	r := gin.Default()
	r.GET("/health", h.Health)
	r.GET("/accounts", h.List)
	r.GET("/accounts/:id", h.GetByID)
	r.POST("/accounts", h.Create)
	return r.Run(addr)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func (h *Handler) List(c *gin.Context) {
	c.JSON(200, mockAccounts)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "id must be a number"})
		return
	}

	for _, a := range mockAccounts {
		if a.ID == id {
			c.JSON(200, a)
			return
		}
	}
	c.JSON(404, gin.H{"error": "account not found"})
}

// Create пока ничего не сохраняет: хранение появится в storage.
func (h *Handler) Create(c *gin.Context) {
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	acc := domain.Account{
		ID:           len(mockAccounts) + 1,
		Phone:        req.Phone,
		Username:     req.Username,
		EmployeeName: req.EmployeeName,
		Status:       domain.StatusActive,
		CreatedAt:    time.Now(),
	}
	c.JSON(201, acc)
}
