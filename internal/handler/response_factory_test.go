package handler

import (
	"app/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(i int) *int {
	return &i
}

func TestCreateDepartmentTree(t *testing.T) {
	t.Run("simple tree", func(t *testing.T) {
		departments := []models.Department{
			{ID: 1, Name: "Head"},
			{ID: 2, Name: "Engineering", ParentID: intPtr(1)},
		}
		tree := createDepartmentTree(1, departments, nil)

		assert.NotNil(t, tree)
		assert.Equal(t, 1, tree.ID)
		assert.Equal(t, "Head", tree.Name)
		assert.Len(t, tree.Children, 1)
		assert.Equal(t, 2, tree.Children[0].ID)
		assert.Equal(t, "Engineering", tree.Children[0].Name)
	})

	t.Run("tree with employees", func(t *testing.T) {
		departments := []models.Department{
			{ID: 1, Name: "Head"},
			{ID: 2, Name: "Engineering", ParentID: intPtr(1)},
		}
		employees := []models.Employee{
			{ID: 101, FullName: "John Doe", DepartmentID: 2},
			{ID: 102, FullName: "Jane Smith", DepartmentID: 2},
		}
		tree := createDepartmentTree(1, departments, employees)

		assert.NotNil(t, tree)
		assert.Len(t, tree.Children, 1)
		engDept := tree.Children[0]
		assert.Equal(t, 2, engDept.ID)
		assert.Len(t, engDept.Employees, 2)
		assert.Equal(t, "John Doe", engDept.Employees[0].FullName)
	})

	t.Run("empty input", func(t *testing.T) {
		tree := createDepartmentTree(1, nil, nil)
		assert.Nil(t, tree, "Tree should be nil for non-existent root ID")

		departments := []models.Department{{ID: 2, Name: "Orphan"}}
		tree = createDepartmentTree(1, departments, nil)
		assert.Nil(t, tree, "Tree should be nil if root ID is not in the list")
	})

	t.Run("deep tree", func(t *testing.T) {
		departments := []models.Department{
			{ID: 1, Name: "L1"},
			{ID: 2, Name: "L2", ParentID: intPtr(1)},
			{ID: 3, Name: "L3", ParentID: intPtr(2)},
			{ID: 4, Name: "L2-sibling", ParentID: intPtr(1)},
		}
		tree := createDepartmentTree(1, departments, nil)

		assert.NotNil(t, tree)
		assert.Equal(t, "L1", tree.Name)
		assert.Len(t, tree.Children, 2) // L2 and L2-sibling

		var l2, l2sibling *DepartmentResponse
		for _, child := range tree.Children {
			if child.ID == 2 {
				l2 = child
			} else {
				l2sibling = child
			}
		}

		assert.NotNil(t, l2)
		assert.NotNil(t, l2sibling)
		assert.Equal(t, "L2", l2.Name)
		assert.Len(t, l2.Children, 1)
		assert.Equal(t, "L3", l2.Children[0].Name)
		assert.Len(t, l2sibling.Children, 0)
	})

	t.Run("employees in multiple departments", func(t *testing.T) {
		departments := []models.Department{
			{ID: 1, Name: "Head"},
			{ID: 2, Name: "Engineering", ParentID: intPtr(1)},
			{ID: 3, Name: "Sales", ParentID: intPtr(1)},
		}
		employees := []models.Employee{
			{ID: 101, FullName: "Lead Engineer", DepartmentID: 2},
			{ID: 201, FullName: "Sales Lead", DepartmentID: 3},
		}
		tree := createDepartmentTree(1, departments, employees)

		assert.NotNil(t, tree)
		assert.Len(t, tree.Children, 2)

		var engDept, salesDept *DepartmentResponse
		for _, child := range tree.Children {
			if child.ID == 2 {
				engDept = child
			} else {
				salesDept = child
			}
		}

		assert.NotNil(t, engDept)
		assert.Len(t, engDept.Employees, 1)
		assert.Equal(t, "Lead Engineer", engDept.Employees[0].FullName)

		assert.NotNil(t, salesDept)
		assert.Len(t, salesDept.Employees, 1)
		assert.Equal(t, "Sales Lead", salesDept.Employees[0].FullName)
	})
}
