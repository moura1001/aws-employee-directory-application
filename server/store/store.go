package store

import (
	"github.com/moura1001/aws-employee-directory-application/server/model"
)

type EmployeeStore interface {
	ListEmployees() []*model.Employee
	LoadEmployee(employeeId string) *model.Employee
	AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) error
	UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string)
	DeleteEmployee(employeeId string)
}
