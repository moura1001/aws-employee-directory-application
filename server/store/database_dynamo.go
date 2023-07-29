package store

import (
	"github.com/moura1001/aws-employee-directory-application/server/model"
)

type DatabaseDynamo struct{}

func NewDatabaseDynamo() *DatabaseDynamo {
	return new(DatabaseDynamo)
}

func (db *DatabaseDynamo) ListEmployees() []*model.Employee {
	return []*model.Employee{}
}

func (db *DatabaseDynamo) LoadEmployee(employeeId string) *model.Employee {
	return nil
}

func (db *DatabaseDynamo) AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) error {
	return nil
}

func (db *DatabaseDynamo) UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string) {

}

func (db *DatabaseDynamo) DeleteEmployee(employeeId string) {

}
