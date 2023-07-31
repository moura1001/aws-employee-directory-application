package store

import (
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

func (db *InMemoryStore) ListEmployees() []*model.Employee {
	return db.employees
}

func (db *InMemoryStore) LoadEmployee(employeeId string) *model.Employee {
	for _, e := range db.employees {
		if e.Id == employeeId {
			return e
		}
	}
	return nil
}

func (db *InMemoryStore) AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) error {
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
	return nil
}

func (db *InMemoryStore) UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string) {
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
	}
}

func (db *InMemoryStore) DeleteEmployee(employeeId string) {
	for i, e := range db.employees {
		if e.Id == employeeId {
			db.employees = append(db.employees[:i], db.employees[i+1:]...)
			return
		}
	}
}
