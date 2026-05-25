package handler

import (
	"app/internal/models"
	"time"
)

type DepartmentResponse struct {
	ID        int                   `json:"id"`
	Name      string                `json:"name"`
	ParentID  *int                  `json:"parent_id,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	Employees []models.Employee     `json:"employees,omitempty"`
	Children  []*DepartmentResponse `json:"children,omitempty"`
}

func createDepartmentTree(rootID int, departments []models.Department, employees []models.Employee) *DepartmentResponse {
	departmentMap := make(map[int]*DepartmentResponse)
	employeeMap := make(map[int][]models.Employee)

	for _, emp := range employees {
		employeeMap[emp.DepartmentID] = append(employeeMap[emp.DepartmentID], emp)
	}

	for _, dept := range departments {
		departmentMap[dept.ID] = &DepartmentResponse{
			ID:        dept.ID,
			Name:      dept.Name,
			ParentID:  dept.ParentID,
			CreatedAt: dept.CreatedAt,
			Employees: employeeMap[dept.ID],
			Children:  []*DepartmentResponse{},
		}
	}

	var root *DepartmentResponse
	for _, node := range departmentMap {
		if node.ParentID != nil {
			if parent, ok := departmentMap[*node.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
		if node.ID == rootID {
			root = node
		}
	}

	return root
}
