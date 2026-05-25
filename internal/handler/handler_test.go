package handler_test

import (
	"app/internal/handler"
	"app/internal/models"
	"app/internal/storage"
	"app/internal/validation"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	validator = validation.New()
)

func TestCreateDepartment(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(100)}))
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("CreateDepartment", mock.Anything, mock.AnythingOfType("*models.Department")).Return(nil)
		body := `{"name": "  Engineering  "}`
		req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateDepartment(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var responseDept models.Department
		err := json.Unmarshal(w.Body.Bytes(), &responseDept)
		assert.NoError(t, err)
		assert.Equal(t, "Engineering", responseDept.Name)
		assert.Equal(t, 1, responseDept.ID)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		body := `{"name": "Test",,}`
		req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateDepartment(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid json format")
	})

	t.Run("Validation Error", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		body := `{"name": ""}`
		req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateDepartment(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "field 'name' - required")
	})

	t.Run("Name Conflict", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("CreateDepartment", mock.Anything, mock.AnythingOfType("*models.Department")).Return(storage.DepartmentNameNotUniqueErr)
		body := `{"name": "Existing Dept"}`
		req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateDepartment(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), storage.DepartmentNameNotUniqueErr.Error())
		mockStorage.AssertExpectations(t)
	})

	t.Run("Storage Internal Error", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("CreateDepartment", mock.Anything, mock.AnythingOfType("*models.Department")).Return(errors.New("something went wrong"))
		body := `{"name": "Test Dept"}`
		req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateDepartment(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestAddEmployee(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(100)}))
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("CreateEmployee", mock.Anything, mock.AnythingOfType("*models.Employee")).Return(nil)
		body := `{"full_name": "John Doe", "position": "Developer"}`
		req := httptest.NewRequest(http.MethodPost, "/departments/1/employees", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request = req

		handler.AddEmployee(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var respEmp models.Employee
		err := json.Unmarshal(w.Body.Bytes(), &respEmp)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", respEmp.FullName)
		assert.Equal(t, 101, respEmp.ID)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Department Not Found", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("CreateEmployee", mock.Anything, mock.AnythingOfType("*models.Employee")).Return(&storage.DepartmentNotFoundErr{Id: 99})
		body := `{"full_name": "John Doe", "position": "Developer"}`
		req := httptest.NewRequest(http.MethodPost, "/departments/99/employees", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "99"}}
		c.Request = req

		handler.AddEmployee(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestChangeParent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(100)}))
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		parentID := 2
		updatedDept := &models.Department{ID: 1, Name: "Test", ParentID: &parentID}
		mockStorage.On("ChangeParent", mock.Anything, 1, &parentID).Return(updatedDept, nil)
		body := `{"parent_id": 2}`
		req := httptest.NewRequest(http.MethodPatch, "/departments/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request = req

		handler.ChangeParent(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var respDept models.Department
		err := json.Unmarshal(w.Body.Bytes(), &respDept)
		assert.NoError(t, err)
		assert.Equal(t, 1, respDept.ID)
		assert.Equal(t, &parentID, respDept.ParentID)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Circular Dependency", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		parentID := 1
		mockStorage.On("ChangeParent", mock.Anything, 1, &parentID).Return(nil, storage.CircularDependencyErr)
		body := `{"parent_id": 1}`
		req := httptest.NewRequest(http.MethodPatch, "/departments/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request = req

		handler.ChangeParent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), storage.CircularDependencyErr.Error())
		mockStorage.AssertExpectations(t)
	})
}

func TestDeleteDepartment(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(100)}))
	gin.SetMode(gin.TestMode)

	t.Run("Success Cascade", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("DeleteDepartmentCascade", mock.Anything, 1).Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/departments/1?mode=cascade", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request = req

		handler.DeleteDepartment(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Success Reassign", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		mockStorage.On("DeleteDepartmentReassign", mock.Anything, 1, 2).Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/departments/1?mode=reassign&reassign_to_department_id=2", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request = req

		handler.DeleteDepartment(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Validation Error - Missing Reassign ID", func(t *testing.T) {
		mockStorage := new(storage.MockStorage)
		handler := handler.NewHandler(mockStorage, logger, validator)
		req := httptest.NewRequest(http.MethodDelete, "/departments/1?mode=reassign", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request = req

		handler.DeleteDepartment(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "field 'reassign_to_department_id' - required_if")
	})
}
