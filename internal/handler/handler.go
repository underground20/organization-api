package handler

import (
	"app/internal/http/response"
	"app/internal/models"
	"app/internal/storage"
	"app/lib/logger"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	deleteModeCascade  = "cascade"
	deleteModeReassign = "reassign"
)

type createDepartmentRequest struct {
	Name     string `json:"name" validate:"required,max=200"`
	ParentID *int   `json:"parent_id"`
}

type deleteDepartmentRequest struct {
	Mode                   string `form:"mode" validate:"required,oneof=cascade reassign"`
	ReassignToDepartmentID *int   `form:"reassign_to_department_id" validate:"required_if=Mode reassign"`
}

type changeParentRequest struct {
	Name     *string `json:"name"`
	ParentID *int    `json:"parent_id"`
}

type addEmployeeRequest struct {
	FullName string     `json:"full_name" validate:"required,max=200"`
	Position string     `json:"position" validate:"required,max=200"`
	HiredAt  *time.Time `json:"hired_at"`
}

type getDepartmentRequest struct {
	Depth            int  `form:"depth" validate:"min=1,max=5"`
	IncludeEmployees bool `form:"include_employees"`
}

type Storage interface {
	CreateDepartment(ctx context.Context, e *models.Department) error
	CreateEmployee(ctx context.Context, e *models.Employee) error
	ChangeParent(ctx context.Context, departmentId int, newParentId *int) (*models.Department, error)
	DeleteDepartmentCascade(ctx context.Context, departmentID int) error
	DeleteDepartmentReassign(ctx context.Context, departmentID int, reassignToDepartmentID int) error
	GetDepartmentTree(ctx context.Context, departmentID int, depth int, includeEmployees bool) ([]models.Department, []models.Employee, error)
}

type Handler struct {
	departmentStorage Storage
	logger            *slog.Logger
	validator         *validator.Validate
}

func NewHandler(
	departmentStorage Storage,
	logger *slog.Logger,
	validator *validator.Validate,
) *Handler {
	return &Handler{
		departmentStorage: departmentStorage,
		logger:            logger,
		validator:         validator,
	}
}

func (h *Handler) CreateDepartment(c *gin.Context) {
	var request createDepartmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("failed to bind json", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid json format",
		})
		return
	}

	request.Name = strings.TrimSpace(request.Name)

	if err := h.validator.Struct(request); err != nil {
		resp := response.ValidationError(err.(validator.ValidationErrors))
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	department := &models.Department{
		Name:     request.Name,
		ParentID: request.ParentID,
	}

	err := h.departmentStorage.CreateDepartment(c.Request.Context(), department)
	if err != nil {
		if errors.Is(err, storage.DepartmentNameNotUniqueErr) {
			c.JSON(http.StatusConflict, response.Response{Message: err.Error()})
			return
		}
		h.logger.Error("Failed to create department", logger.Err(err))
		c.JSON(http.StatusInternalServerError, response.UnhandledError())
		return
	}

	c.JSON(http.StatusCreated, department)
}

func (h *Handler) GetDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Error("failed to convert department id param to int", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{Message: "Invalid department ID format"})
		return
	}

	var req getDepartmentRequest
	req.Depth = 1
	req.IncludeEmployees = true
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("failed to bind query params", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid query parameters",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		resp := response.ValidationError(err.(validator.ValidationErrors))
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	departments, employees, err := h.departmentStorage.GetDepartmentTree(c.Request.Context(), id, req.Depth, req.IncludeEmployees)
	if err != nil {
		if storage.IsDepartmentNotFound(err) {
			c.JSON(http.StatusNotFound, response.Response{Message: err.Error()})
			return
		}
		h.logger.Error("failed to get department tree", logger.Err(err))
		c.JSON(http.StatusInternalServerError, response.UnhandledError())
		return
	}

	tree := createDepartmentTree(id, departments, employees)
	c.JSON(http.StatusOK, tree)
}

func (h *Handler) AddEmployee(c *gin.Context) {
	departmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Error("failed to convert department id param to int", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid department ID format",
		})
		return
	}

	var req addEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind json", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid json format",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		resp := response.ValidationError(err.(validator.ValidationErrors))
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	employee := &models.Employee{
		DepartmentID: departmentID,
		FullName:     req.FullName,
		Position:     req.Position,
		HiredAt:      req.HiredAt,
	}

	err = h.departmentStorage.CreateEmployee(c.Request.Context(), employee)
	if err != nil {
		if storage.IsDepartmentNotFound(err) {
			c.JSON(http.StatusNotFound, response.Response{Message: err.Error()})
			return
		}
		h.logger.Error("failed to create employee", logger.Err(err))
		c.JSON(http.StatusInternalServerError, response.UnhandledError())
		return
	}

	c.JSON(http.StatusCreated, employee)
}

func (h *Handler) DeleteDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Error("failed to convert id param to int", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid department ID format",
		})
		return
	}

	var req deleteDepartmentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("failed to bind query params", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{Message: "Invalid query parameters"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		resp := response.ValidationError(err.(validator.ValidationErrors))
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	switch req.Mode {
	case deleteModeCascade:
		err = h.departmentStorage.DeleteDepartmentCascade(c.Request.Context(), id)
	case deleteModeReassign:
		err = h.departmentStorage.DeleteDepartmentReassign(c.Request.Context(), id, *req.ReassignToDepartmentID)
	}

	if err != nil {
		if storage.IsDepartmentNotFound(err) {
			c.JSON(http.StatusNotFound, response.Response{Message: err.Error()})
			return
		}
		h.logger.Error("failed to delete department", logger.Err(err))
		c.JSON(http.StatusInternalServerError, response.UnhandledError())
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

func (h *Handler) ChangeParent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Error("failed to convert id param to int", logger.Err(err))
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid department ID",
		})
		return
	}

	var req changeParentRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "Invalid json format",
		})
		return
	}

	department, err := h.departmentStorage.ChangeParent(c.Request.Context(), id, req.ParentID)
	if err != nil {
		if storage.IsDepartmentNotFound(err) {
			c.JSON(http.StatusNotFound, response.Response{Message: err.Error()})
			return
		}

		if errors.Is(err, storage.CircularDependencyErr) {
			c.JSON(http.StatusBadRequest, response.Response{Message: err.Error()})
			return
		}

		h.logger.Error("failed to change parent", logger.Err(err))
		c.JSON(http.StatusInternalServerError, response.UnhandledError())
		return
	}

	c.JSON(http.StatusOK, department)
}
