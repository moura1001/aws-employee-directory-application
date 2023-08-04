package model

type Employee struct {
	Id       string   `dynamodbav:"id"`
	Photo    *Photo   `dynamodbav:"photo"`
	FullName string   `dynamodbav:"full_name"`
	Location string   `dynamodbav:"location"`
	JobTitle string   `dynamodbav:"job_title"`
	Badges   []string `dynamodbav:"badges"`
}

type Photo struct {
	ObjectKey string `dynamodbav:"object_key"`
	SignedUrl string `dynamodbav:"-"`
}

func (e Employee) HasBadge(badge string) bool {
	for _, b := range e.Badges {
		if b == badge {
			return true
		}
	}

	return false
}
