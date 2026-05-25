package postgres

import (
	"app/internal/models"
	"app/internal/storage"
	"app/lib/logger"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type Storage struct {
	Db     *gorm.DB
	logger *slog.Logger
}

func NewStorage(Db *gorm.DB, logger *slog.Logger) *Storage {
	return &Storage{Db: Db, logger: logger}
}

func (s *Storage) CreateDepartment(ctx context.Context, department *models.Department) error {
	return s.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if department.ParentID != nil {
			if !s.isParentExists(tx, department.ParentID) {
				return &storage.DepartmentNotFoundErr{Id: *department.ParentID}
			}
		}

		// Проверка уникальности имени
		var count int64
		query := tx.Model(&models.Department{}).Where(
			"name = ? AND parent_id = ?",
			department.Name,
			department.ParentID,
		)
		if err := query.Count(&count).Error; err != nil {
			s.logger.Error("failed to check department name uniqueness", logger.Err(err))
			return err
		}

		if count > 0 {
			return storage.DepartmentNameNotUniqueErr
		}

		if err := tx.Create(department).Error; err != nil {
			s.logger.Error("failed to create department", logger.Err(err))
			return err
		}

		return nil
	})
}

func (s *Storage) ChangeParent(ctx context.Context, departmentID int, newParentID *int) (*models.Department, error) {
	var department models.Department
	err := s.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&department, departmentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &storage.DepartmentNotFoundErr{Id: departmentID}
			}
			return err
		}

		if newParentID != nil {
			if departmentID == *newParentID {
				return storage.CircularDependencyErr
			}

			if !s.isParentExists(tx, newParentID) {
				return &storage.DepartmentNotFoundErr{Id: *newParentID}
			}

			childIDs, err := s.getDepartmentSubtreeIDs(tx, departmentID)
			if err != nil {
				return err
			}
			for _, childID := range childIDs {
				if childID == *newParentID {
					return storage.CircularDependencyErr
				}
			}
		}

		department.ParentID = newParentID
		if err := tx.Save(&department).Error; err != nil {
			s.logger.Error("failed to change parent department", logger.Err(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &department, nil
}

func (s *Storage) GetDepartmentTree(
	ctx context.Context,
	departmentID int,
	depth int,
	includeEmployees bool,
) ([]models.Department, []models.Employee, error) {
	query := `
        WITH RECURSIVE department_subtree AS (
            SELECT id, name, parent_id, created_at, 1 as depth
            FROM departments
            WHERE id = ?
            UNION ALL
            SELECT d.id, d.name, d.parent_id, d.created_at, ds.depth + 1
            FROM departments d
            JOIN department_subtree ds ON d.parent_id = ds.id
            WHERE ds.depth <= ?
        )
        SELECT id, name, parent_id, created_at FROM department_subtree;
    `
	var departments []models.Department
	if err := s.Db.WithContext(ctx).Raw(query, departmentID, depth).Scan(&departments).Error; err != nil {
		s.logger.Error("failed to get department tree", logger.Err(err))
		return nil, nil, err
	}

	if len(departments) == 0 {
		return nil, nil, &storage.DepartmentNotFoundErr{Id: departmentID}
	}

	var employees []models.Employee
	if includeEmployees {
		var departmentIDs []int
		for _, d := range departments {
			departmentIDs = append(departmentIDs, d.ID)
		}

		if err := s.Db.WithContext(ctx).Where("department_id IN ?", departmentIDs).Order("full_name asc").Find(&employees).Error; err != nil {
			s.logger.Error("failed to get employees for department tree", logger.Err(err))
			return nil, nil, err
		}
	}

	return departments, employees, nil
}

func (s *Storage) CreateEmployee(ctx context.Context, employee *models.Employee) error {
	return s.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var departmentExists int64
		if err := tx.Model(&models.Department{}).Where("id = ?", employee.DepartmentID).Count(&departmentExists).Error; err != nil {
			s.logger.Error("failed to check department existence", logger.Err(err))
			return err
		}
		if departmentExists == 0 {
			return &storage.DepartmentNotFoundErr{Id: employee.DepartmentID}
		}

		if err := tx.Create(employee).Error; err != nil {
			s.logger.Error("failed to create employee", logger.Err(err))
			return err
		}

		return nil
	})
}

func (s *Storage) DeleteDepartmentCascade(ctx context.Context, departmentID int) error {
	return s.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		departmentIDs, err := s.getDepartmentSubtreeIDs(tx, departmentID)
		if err != nil {
			return err
		}

		return s.deleteCascade(tx, departmentIDs)
	})
}

func (s *Storage) DeleteDepartmentReassign(ctx context.Context, departmentID int, reassignToDepartmentID int) error {
	return s.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		departmentIDs, err := s.getDepartmentSubtreeIDs(tx, departmentID)
		if err != nil {
			return err
		}

		return s.deleteReassign(tx, departmentID, departmentIDs, &reassignToDepartmentID)
	})
}

func (s *Storage) getDepartmentSubtreeIDs(tx *gorm.DB, departmentID int) ([]int, error) {
	var ids []int
	err := tx.Raw(`
        WITH RECURSIVE department_tree AS (
            SELECT id FROM departments WHERE id = ?
            UNION ALL
            SELECT d.id FROM departments d JOIN department_tree dt ON d.parent_id = dt.id
        )
        SELECT id FROM department_tree;
    `, departmentID).Scan(&ids).Error

	if err != nil {
		s.logger.Error("failed to get department tree", logger.Err(err))
		return nil, fmt.Errorf("failed to get department tree: %w", err)
	}

	if len(ids) == 0 {
		return nil, &storage.DepartmentNotFoundErr{Id: departmentID}
	}

	return ids, nil
}

func (s *Storage) isParentExists(tx *gorm.DB, parentID *int) bool {
	var parentExists int64
	if err := tx.Model(&models.Department{}).Where("id = ?", parentID).Count(&parentExists).Error; err != nil {
		s.logger.Error("failed to check department existence", logger.Err(err))
		return false
	}

	return parentExists != 0
}

func (s *Storage) deleteCascade(tx *gorm.DB, departmentIDs []int) error {
	if err := tx.Where("department_id IN ?", departmentIDs).Delete(&models.Employee{}).Error; err != nil {
		s.logger.Error("failed to delete employees in cascade mode", logger.Err(err))
		return err
	}
	if err := tx.Where("id IN ?", departmentIDs).Delete(&models.Department{}).Error; err != nil {
		s.logger.Error("failed to delete departments in cascade mode", logger.Err(err))
		return err
	}
	return nil
}

func (s *Storage) deleteReassign(tx *gorm.DB, departmentID int, departmentIDs []int, reassignToDepartmentID *int) error {
	if reassignToDepartmentID == nil {
		return errors.New("reassign_to_department_id is required for reassign mode")
	}

	var count int64
	tx.Model(&models.Department{}).Where("id = ? AND id NOT IN ?", *reassignToDepartmentID, departmentIDs).Count(&count)
	if count == 0 {
		return fmt.Errorf("reassign target department with id %d not found or is a child of the department being deleted", *reassignToDepartmentID)
	}

	if err := tx.Model(&models.Employee{}).Where("department_id IN ?", departmentIDs).Update("department_id", *reassignToDepartmentID).Error; err != nil {
		s.logger.Error("failed to reassign employees", logger.Err(err))
		return err
	}
	if err := tx.Model(&models.Department{}).Where("parent_id = ?", departmentID).Update("parent_id", *reassignToDepartmentID).Error; err != nil {
		s.logger.Error("failed to reassign child departments", logger.Err(err))
		return err
	}
	if err := tx.Where("id = ?", departmentID).Delete(&models.Department{}).Error; err != nil {
		s.logger.Error("failed to delete single department in reassign mode", logger.Err(err))
		return err
	}
	return nil
}
