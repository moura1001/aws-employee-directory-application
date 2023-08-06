package store

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/moura1001/aws-employee-directory-application/server/model"
	"github.com/moura1001/aws-employee-directory-application/server/utils"
)

type DynamoStore struct {
	table string
}

func NewDynamoStore() *DynamoStore {
	return &DynamoStore{
		table: "Employees",
	}
}

func (db *DynamoStore) ListEmployees() ([]*model.Employee, error) {
	errMsg := "error to get employee list%s. Details: '%s'"

	var emps []*model.Employee

	svc, err := db.getDynamoClient()
	if err != nil {
		return nil, fmt.Errorf(errMsg, "", err)
	}

	empData, err := svc.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(db.table),
	})
	if err != nil {
		return nil, fmt.Errorf(errMsg, " Query", err)
	}

	err = attributevalue.UnmarshalListOfMaps(empData.Items, &emps)
	if err != nil {
		return nil, fmt.Errorf(errMsg, " UnmarshalListOfMaps", err)
	}

	return emps, nil
}

func (db *DynamoStore) LoadEmployee(employeeId string) (*model.Employee, error) {
	errMsg := "error to get employee data%s. Details: '%s'"

	emp := &model.Employee{Photo: new(model.Photo)}

	svc, err := db.getDynamoClient()
	if err != nil {
		return nil, fmt.Errorf(errMsg, "", err)
	}

	selectedKeys := map[string]string{
		"id": employeeId,
	}
	key, _ := attributevalue.MarshalMap(selectedKeys)

	empItem, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(db.table),
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf(errMsg, " GetItem", err)
	}

	if empItem.Item == nil {
		return nil, fmt.Errorf(errMsg, "", "data not found")
	}

	err = attributevalue.UnmarshalMap(empItem.Item, emp)
	if err != nil {
		return nil, fmt.Errorf(errMsg, " UnmarshalMap", err)
	}

	return emp, nil
}

func (db *DynamoStore) AddEmployee(objectKey, fullName, location, jobTitle string, badges []string) (string, error) {
	errMsg := "error to insert employee data%s. Details: '%s'"

	svc, err := db.getDynamoClient()
	if err != nil {
		return "", fmt.Errorf(errMsg, "", err)
	}

	emp := &model.Employee{
		Id: uuid.NewString(),
		Photo: &model.Photo{
			ObjectKey: objectKey,
		},
		FullName: fullName,
		Location: location,
		JobTitle: jobTitle,
		Badges:   badges,
	}

	empItem, err := attributevalue.MarshalMap(emp)
	if err != nil {
		return "", fmt.Errorf(errMsg, " MarshalMap", err)
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(db.table),
		Item:      empItem,
	})
	if err != nil {
		return "", fmt.Errorf(errMsg, " PutItem", err)
	}

	return emp.Id, nil
}

func (db *DynamoStore) UpdateEmployee(employeeId, objectKey, fullName, location, jobTitle string, badges []string) error {
	errMsg := "error to update employee data%s. Details: '%s'"

	svc, err := db.getDynamoClient()
	if err != nil {
		return fmt.Errorf(errMsg, "", err)
	}

	selectedKeys := map[string]string{
		"id": employeeId,
	}
	key, _ := attributevalue.MarshalMap(selectedKeys)

	upd := expression.
		Set(expression.Name("photo.object_key"), expression.Value(objectKey)).
		Set(expression.Name("full_name"), expression.Value(fullName)).
		Set(expression.Name("location"), expression.Value(location)).
		Set(expression.Name("job_title"), expression.Value(jobTitle)).
		Set(expression.Name("badges"), expression.Value(badges))

	expr, _ := expression.NewBuilder().WithUpdate(upd).Build()

	_, err = svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 aws.String(db.table),
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       aws.String("attribute_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf(errMsg, " UpdateItem", err)
	}

	return nil
}

func (db *DynamoStore) DeleteEmployee(employeeId string) error {
	errMsg := "error to delete employee data%s. Details: '%s'"

	svc, err := db.getDynamoClient()
	if err != nil {
		return fmt.Errorf(errMsg, "", err)
	}

	selectedKeys := map[string]string{
		"id": employeeId,
	}
	key, _ := attributevalue.MarshalMap(selectedKeys)

	_, err = svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(db.table),
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf(errMsg, " DeleteItem", err)
	}

	return nil
}

func (db *DynamoStore) getDynamoClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(opts *config.LoadOptions) error {
		opts.Region = utils.AWS_DEFAULT_REGION
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error to get dynamo connection. Details: '%s'", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}
