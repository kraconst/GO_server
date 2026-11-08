package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/usecase"
)

type Handler struct {
	accountUsecase  *usecase.AccountUsecase
	employeeUsecase *usecase.EmployeeUsecase
}

func New(accountUsecase *usecase.AccountUsecase, employeeUsecase *usecase.EmployeeUsecase) *Handler {
	return &Handler{
		accountUsecase:  accountUsecase,
		employeeUsecase: employeeUsecase,
	}
}

var mockAccounts = []domain.Account{
	{ID: 1, Phone: "+79991112233", Username: "@sales_anna", EmployeeName: "Анна", Status: domain.StatusActive, CreatedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Phone: "+79994445566", Username: "@support_oleg", EmployeeName: "Олег", Status: domain.StatusDisconnected, CreatedAt: time.Date(2026, 9, 15, 14, 30, 0, 0, time.UTC)},
}

var mockEmployees = []domain.Employee{
	{ID: 1, CompanyID: 1, Name: "Анна", Position: "Менеджер по продажам", IsActive: true},
	{ID: 2, CompanyID: 1, Name: "Олег", Position: "Служба поддержки", IsActive: true},
}

type createAccountRequest struct {
	Phone        string `json:"phone" binding:"required"`
	Username     string `json:"username" binding:"required"`
	EmployeeName string `json:"employee_name" binding:"required"`
}

type createEmployeeRequest struct {
	Name     string `json:"name" binding:"required"`
	Position string `json:"position"`
}

type updateEmployeeRequest struct {
	Name     string `json:"name" binding:"required"`
	Position string `json:"position"`
}

func (h *Handler) Run(addr string) error {
	r := gin.Default()
	r.GET("/health", h.Health)

	// Accounts
	r.GET("/accounts", h.List)
	r.GET("/accounts/:id", h.GetByID)
	r.POST("/accounts", h.Create)

	// Employees
	r.GET("/employees", h.ListEmployees)
	r.GET("/employees/:id", h.GetEmployeeByID)
	r.POST("/employees", h.CreateEmployee)
	r.POST("/employees/:id", h.UpdateEmployee)
	r.POST("/employees/:id/deactivate", h.DeactivateEmployee)

	return r.Run(addr)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// --- Accounts ---

func (h *Handler) List(c *gin.Context) {
	if h.accountUsecase == nil {
		c.JSON(200, mockAccounts)
		return
	}

	accounts, err := h.accountUsecase.List(c.Request.Context(), 1)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, accounts)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "id must be a number"})
		return
	}

	if h.accountUsecase == nil {
		for _, a := range mockAccounts {
			if a.ID == int64(id) {
				c.JSON(200, a)
				return
			}
		}
		c.JSON(404, gin.H{"error": "account not found"})
		return
	}

	acc, err := h.accountUsecase.GetByID(c.Request.Context(), int64(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": "account not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, acc)
}

func (h *Handler) Create(c *gin.Context) {
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if h.accountUsecase == nil {
		acc := domain.Account{
			ID:           int64(len(mockAccounts) + 1),
			Phone:        req.Phone,
			Username:     req.Username,
			EmployeeName: req.EmployeeName,
			Status:       domain.StatusActive,
			CreatedAt:    time.Now(),
		}
		c.JSON(201, acc)
		return
	}

	acc, err := h.accountUsecase.Create(c.Request.Context(), usecase.CreateAccountInput{
		CompanyID:    1,
		EmployeeID:   1,
		Phone:        req.Phone,
		Username:     req.Username,
		EmployeeName: req.EmployeeName,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, acc)
}

// --- Employees ---

func (h *Handler) ListEmployees(c *gin.Context) {
	if h.employeeUsecase == nil {
		c.JSON(200, mockEmployees)
		return
	}

	employees, err := h.employeeUsecase.List(c.Request.Context(), 1)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, employees)
}

func (h *Handler) GetEmployeeByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "id must be a number"})
		return
	}

	if h.employeeUsecase == nil {
		for _, e := range mockEmployees {
			if e.ID == int64(id) {
				c.JSON(200, e)
				return
			}
		}
		c.JSON(404, gin.H{"error": "employee not found"})
		return
	}

	emp, err := h.employeeUsecase.GetByID(c.Request.Context(), int64(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, emp)
}

func (h *Handler) CreateEmployee(c *gin.Context) {
	var req createEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if h.employeeUsecase == nil {
		emp := domain.Employee{
			ID:        int64(len(mockEmployees) + 1),
			CompanyID: 1,
			Name:      req.Name,
			Position:  req.Position,
			IsActive:  true,
		}
		c.JSON(201, emp)
		return
	}

	emp, err := h.employeeUsecase.Create(c.Request.Context(), usecase.CreateEmployeeInput{
		CompanyID: 1,
		Name:      req.Name,
		Position:  req.Position,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, emp)
}

func (h *Handler) UpdateEmployee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "id must be a number"})
		return
	}

	var req updateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if h.employeeUsecase == nil {
		c.JSON(200, gin.H{"status": "updated"})
		return
	}

	err = h.employeeUsecase.Update(c.Request.Context(), domain.Employee{
		ID:        int64(id),
		CompanyID: 1,
		Name:      req.Name,
		Position:  req.Position,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "updated"})
}

func (h *Handler) DeactivateEmployee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "id must be a number"})
		return
	}

	if h.employeeUsecase == nil {
		c.JSON(200, gin.H{"status": "deactivated"})
		return
	}

	err = h.employeeUsecase.Deactivate(c.Request.Context(), int64(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "deactivated"})
}
