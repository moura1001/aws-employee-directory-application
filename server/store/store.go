package store

import (
	"github.com/moura1001/aws-employee-directory-application/server/model"
)

type EmployeeStore interface {
	ListEmployees() ([]*model.Employee, error)
	LoadEmployee(employeeId string) (*model.Employee, error)
	AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) (string, error)
	UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string) error
	DeleteEmployee(employeeId string) error
}
