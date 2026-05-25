package storage

import (
	"app/internal/models"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateDepartment(ctx context.Context, d *models.Department) error {
	args := m.Called(ctx, d)
	if args.Error(0) == nil {
		d.ID = 1
	}
	return args.Error(0)
}

func (m *MockStorage) CreateEmployee(ctx context.Context, e *models.Employee) error {
	args := m.Called(ctx, e)
	if args.Error(0) == nil {
		e.ID = 101
	}
	return args.Error(0)
}

func (m *MockStorage) ChangeParent(ctx context.Context, id int, pid *int) (*models.Department, error) {
	args := m.Called(ctx, id, pid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

func (m *MockStorage) DeleteDepartmentCascade(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStorage) DeleteDepartmentReassign(ctx context.Context, id int, reassignID int) error {
	args := m.Called(ctx, id, reassignID)
	return args.Error(0)
}

func (m *MockStorage) GetDepartmentTree(ctx context.Context, id int, depth int, includeEmployees bool) ([]models.Department, []models.Employee, error) {
	args := m.Called(ctx, id, depth, includeEmployees)
	var depts []models.Department
	var emps []models.Employee
	if args.Get(0) != nil {
		depts = args.Get(0).([]models.Department)
	}
	if args.Get(1) != nil {
		emps = args.Get(1).([]models.Employee)
	}
	return depts, emps, args.Error(2)
}
