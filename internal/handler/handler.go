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
	accountUsecase           *usecase.AccountUsecase
	employeeUsecase          *usecase.EmployeeUsecase
	messageUsecase           *usecase.MessageUsecase
	accountDailyStatsUsecase *usecase.AccountDailyStatsUsecase
}

func New(
	accountUsecase *usecase.AccountUsecase,
	employeeUsecase *usecase.EmployeeUsecase,
	messageUsecase *usecase.MessageUsecase,
	accountDailyStatsUsecase *usecase.AccountDailyStatsUsecase,
) *Handler {
	return &Handler{
		accountUsecase:           accountUsecase,
		employeeUsecase:          employeeUsecase,
		messageUsecase:           messageUsecase,
		accountDailyStatsUsecase: accountDailyStatsUsecase,
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

var mockMessages = []domain.Message{
	{ID: 1, ChatID: 1, TelegramMessageID: 101, Direction: domain.MessageDirectionIn, Text: "Здравствуйте! Подскажите по условиям.", SentAt: time.Date(2026, 9, 1, 10, 5, 0, 0, time.UTC)},
	{ID: 2, ChatID: 1, TelegramMessageID: 102, Direction: domain.MessageDirectionOut, Text: "Добрый день! Сейчас направлю КП.", SentAt: time.Date(2026, 9, 1, 10, 6, 0, 0, time.UTC)},
}

var mockDashboard = domain.DashboardStats{
	MessagesIn:      1420,
	MessagesOut:     1180,
	ActiveChats:     85,
	NewContacts:     34,
	AvgResponseTime: 4 * time.Minute,
	UnansweredChats: 3,
}

var mockAccountRatings = []domain.AccountRating{
	{
		AccountID:       1,
		Username:        "@sales_anna",
		EmployeeName:    "Анна",
		MessagesIn:      850,
		MessagesOut:     720,
		TotalMessages:   1570,
		ActiveChats:     52,
		NewContacts:     21,
		AvgResponseTime: 3 * time.Minute,
	},
	{
		AccountID:       2,
		Username:        "@support_oleg",
		EmployeeName:    "Олег",
		MessagesIn:      570,
		MessagesOut:     460,
		TotalMessages:   1030,
		ActiveChats:     33,
		NewContacts:     13,
		AvgResponseTime: 5 * time.Minute,
	},
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

	// Messages
	r.GET("/chats/:id/messages", h.ListMessagesByChat)

	// Analytics
	r.GET("/analytics/dashboard", h.GetDashboard)
	r.GET("/analytics/accounts", h.GetAccountRatings)

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

// --- Messages ---

func (h *Handler) ListMessagesByChat(c *gin.Context) {
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || chatID <= 0 {
		c.JSON(400, gin.H{"error": "id must be a positive number"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil || parsedLimit <= 0 {
			c.JSON(400, gin.H{"error": "limit must be a positive number"})
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		parsedOffset, err := strconv.Atoi(o)
		if err != nil || parsedOffset < 0 {
			c.JSON(400, gin.H{"error": "offset must be a non-negative number"})
			return
		}
		offset = parsedOffset
	}

	if h.messageUsecase == nil {
		res := make([]domain.Message, 0)
		for _, m := range mockMessages {
			if m.ChatID == chatID {
				res = append(res, m)
			}
		}
		c.JSON(200, res)
		return
	}

	messages, err := h.messageUsecase.ListByChat(c.Request.Context(), chatID, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, messages)
}

// --- Analytics ---

func parseDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	var from, to time.Time
	var err error

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must be in format YYYY-MM-DD")
		}
	} else {
		to = time.Now().UTC()
	}

	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("from must be in format YYYY-MM-DD")
		}
	} else {
		from = to.AddDate(0, 0, -30)
	}

	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("from date must not be after to date")
	}

	return from, to, nil
}

func (h *Handler) GetDashboard(c *gin.Context) {
	from, to, err := parseDateRange(c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if h.accountDailyStatsUsecase == nil {
		c.JSON(200, mockDashboard)
		return
	}

	stats, err := h.accountDailyStatsUsecase.GetDashboard(c.Request.Context(), 1, from, to)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, stats)
}

func (h *Handler) GetAccountRatings(c *gin.Context) {
	from, to, err := parseDateRange(c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if h.accountDailyStatsUsecase == nil {
		c.JSON(200, mockAccountRatings)
		return
	}

	ratings, err := h.accountDailyStatsUsecase.GetAccountRatings(c.Request.Context(), 1, from, to)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, ratings)
}
