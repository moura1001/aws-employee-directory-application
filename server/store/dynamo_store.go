package store

import (
	"github.com/moura1001/aws-employee-directory-application/server/model"
)

type DynamoStore struct{}

func NewDynamoStore() *DynamoStore {
	return new(DynamoStore)
}

func (db *DynamoStore) ListEmployees() []*model.Employee {
	return []*model.Employee{}
}

func (db *DynamoStore) LoadEmployee(employeeId string) *model.Employee {
	return nil
}

func (db *DynamoStore) AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) error {
	return nil
}

func (db *DynamoStore) UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string) {

}

func (db *DynamoStore) DeleteEmployee(employeeId string) {

}
