package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/moura1001/aws-employee-directory-application/server/model"
	"github.com/moura1001/aws-employee-directory-application/server/utils"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlStore struct{}

func NewMysqlStore() *MysqlStore {
	return new(MysqlStore)
}

func (db *MysqlStore) ListEmployees() ([]*model.Employee, error) {
	errMsg := "error to get employee list. Details: '%s'"
	conn, err := db.getDatabaseConnection()

	if err == nil {
		defer conn.Close()

		selEmp, err := conn.Query("SELECT id, object_key, full_name, location, job_title, badges FROM employee ORDER BY id DESC")
		if err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}

		res := []*model.Employee{}
		for selEmp.Next() {
			emp := &model.Employee{Photo: new(model.Photo)}
			var b string
			err = selEmp.Scan(&(emp.Id), &(emp.Photo.ObjectKey), &(emp.FullName), &(emp.Location), &(emp.JobTitle), &b)
			if err == nil {

				var badges []string
				if len(b) > 0 {
					badges = strings.Split(b, ",")
				} else {
					badges = []string{}
				}
				emp.Badges = badges

				res = append(res, emp)
			}
		}

		return res, nil

	} else {
		return nil, fmt.Errorf(errMsg, err)
	}
}

func (db *MysqlStore) LoadEmployee(employeeId string) (*model.Employee, error) {
	errMsg := "error to get employee data. Details: '%s'"
	conn, err := db.getDatabaseConnection()

	if err == nil {
		defer conn.Close()

		selEmp, err := conn.Query("SELECT id, object_key, full_name, location, job_title, badges FROM employee WHERE id=?", employeeId)
		if err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}

		emp := &model.Employee{Photo: new(model.Photo)}
		for selEmp.Next() {
			var b string
			err = selEmp.Scan(&(emp.Id), &(emp.Photo.ObjectKey), &(emp.FullName), &(emp.Location), &emp.JobTitle, &b)
			if err == nil {
				var badges []string
				if len(b) > 0 {
					badges = strings.Split(b, ",")
				} else {
					badges = []string{}
				}
				emp.Badges = badges
			}
		}

		if emp.Id != "" {
			return emp, nil
		} else {
			return nil, nil
		}

	} else {
		return nil, fmt.Errorf(errMsg, err)
	}
}

func (db *MysqlStore) AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) (string, error) {
	errMsg := "error to insert employee data. Details: '%s'"
	conn, err := db.getDatabaseConnection()

	if err == nil {
		defer conn.Close()

		query := "INSERT INTO employee(object_key, full_name, location, job_title, badges) VALUES(?,?,?,?,?)"

		b := strings.Join(badges, ",")

		_, err = conn.Exec(query, objectKey, fullName, location, jobTitle, b)
		if err != nil {
			return "", fmt.Errorf(errMsg, err)
		}

		var empId string
		err = conn.QueryRow("SELECT LAST_INSERT_ID()").Scan(&empId)
		if err != nil {
			return "", fmt.Errorf(errMsg, fmt.Errorf("failed to get last inserted id: '%s'", err))
		}

		return empId, nil

	} else {
		return "", fmt.Errorf(errMsg, err)
	}
}

func (db *MysqlStore) UpdateEmployee(employeeId string, objectKey, fullName, location, jobTitle string, badges []string) error {
	errMsg := "error to update employee data. Details: '%s'"

	empId, err := strconv.ParseInt(employeeId, 10, 32)
	if err != nil {
		return fmt.Errorf("employee '%s' does not exist", employeeId)
	}

	conn, err := db.getDatabaseConnection()

	if err == nil {
		defer conn.Close()

		query := "SELECT id FROM employee WHERE id=?"
		err = conn.QueryRow(query, empId).Scan(&empId)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		query = "UPDATE employee SET object_key=?, full_name=?, location=?, job_title=?, badges=? WHERE id=?"

		b := strings.Join(badges, ",")

		_, err = conn.Exec(query, objectKey, fullName, location, jobTitle, b, empId)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		return nil

	} else {
		return fmt.Errorf(errMsg, err)
	}
}

func (db *MysqlStore) DeleteEmployee(employeeId string) error {
	errMsg := "error to delete employee data. Details: '%s'"
	conn, err := db.getDatabaseConnection()

	if err == nil {
		defer conn.Close()

		delEmp, err := conn.Prepare("DELETE FROM employee WHERE id=?")
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}
		defer delEmp.Close()

		_, err = delEmp.Exec(employeeId)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		return nil

	} else {
		return fmt.Errorf(errMsg, err)
	}
}

func (db *MysqlStore) getDatabaseConnection() (*sql.DB, error) {
	connectStr := fmt.Sprintf("%s:%s@(%s)/%s", utils.DATABASE_USER, utils.DATABASE_PASSWORD, utils.DATABASE_HOST, utils.DATABASE_DB_NAME)
	conn, err := sql.Open("mysql", connectStr)

	if err == nil {
		ctx, canc := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer canc()
		err = conn.PingContext(ctx)
		return conn, err
	} else {
		return nil, fmt.Errorf("error to open mysql database connection: Details: '%s'", err)
	}
}
