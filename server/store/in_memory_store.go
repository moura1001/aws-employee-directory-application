package store

import (
	"fmt"
	"strconv"

	"github.com/moura1001/aws-employee-directory-application/server/model"
)

type InMemoryStore struct {
	employees []*model.Employee
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		employees: []*model.Employee{},
	}
}

func (db *InMemoryStore) ListEmployees() ([]*model.Employee, error) {
	return db.employees, nil
}

func (db *InMemoryStore) LoadEmployee(employeeId string) (*model.Employee, error) {
	for _, e := range db.employees {
		if e.Id == employeeId {
			return e, nil
		}
	}
	return nil, nil
}

func (db *InMemoryStore) AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) (string, error) {
	id := strconv.Itoa(len(db.employees))
	db.employees = append(db.employees, &model.Employee{
		Id: id,
		Photo: &model.Photo{
			ObjectKey: objectKey,
		},
		FullName: fullName,
		Location: location,
		JobTitle: jobTitle,
		Badges:   badges,
	})
	return id, nil
}

func (db *InMemoryStore) UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string) error {
	var employee *model.Employee = nil
	for _, e := range db.employees {
		if e.Id == employeeId {
			employee = e
		}
	}

	if employee != nil {
		employee.Photo.ObjectKey = objectKey
		employee.FullName = fullName
		employee.Location = location
		employee.JobTitle = jobTitle
		employee.Badges = badges

		return nil
	}

	return fmt.Errorf("employee '%s' does not exist", employeeId)
}

func (db *InMemoryStore) DeleteEmployee(employeeId string) error {
	for i, e := range db.employees {
		if e.Id == employeeId {
			db.employees = append(db.employees[:i], db.employees[i+1:]...)
			return nil
		}
	}

	return nil
}

func (db *InMemoryStore) IsHealthy() bool {
	return true
}
