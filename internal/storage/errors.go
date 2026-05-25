package storage

import (
	"errors"
	"fmt"
)

var (
	// CircularDependencyErr возникает при попытке создать циклическую зависимость в дереве департаментов.
	// Это включает в себя попытку сделать департамент своим собственным родителем
	// или переместить его в один из своих дочерних департаментов.
	CircularDependencyErr      = errors.New("circular dependency detected: cannot make a department its own parent or move it into its own subtree")
	DepartmentNameNotUniqueErr = fmt.Errorf("department name not unique")
)

type DepartmentNotFoundErr struct {
	Id int
}

func (e *DepartmentNotFoundErr) Error() string {
	return fmt.Sprintf("department with id=%d not found", e.Id)
}

func IsDepartmentNotFound(err error) bool {
	var departmentNotFoundErr *DepartmentNotFoundErr
	ok := errors.As(err, &departmentNotFoundErr)
	return ok
}
