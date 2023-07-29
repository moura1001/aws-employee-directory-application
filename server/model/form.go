package model

import (
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"
)

type Form struct {
	EmployeeId Field
	Photo      Field
	FullName   Field
	Location   Field
	JobTitle   Field
	Badges     Field
}

type Field struct {
	Name       string
	Label      string
	Data       interface{}
	IsRequired bool
}

func NewForm() Form {
	return Form{
		EmployeeId: Field{IsRequired: false, Name: "employee_id"},
		Photo:      Field{IsRequired: false, Name: "photo"},
		FullName:   Field{IsRequired: true, Name: "full_name", Label: "Full Name"},
		Location:   Field{IsRequired: true, Name: "location", Label: "Location"},
		JobTitle:   Field{IsRequired: true, Name: "job_title", Label: "Job Title"},
		Badges:     Field{IsRequired: false, Name: "badges", Label: "Badges"},
	}
}

func (f Field) Contains(key string) bool {
	s, sliceValue := f.Data.([]string)
	if sliceValue {
		for _, b := range s {
			if b == key {
				return true
			}
		}
	}

	return false
}

func (f Field) ToString() (value string) {
	switch t := f.Data.(type) {
	case string:
		value = t
	case []string:
		value = strings.Join(t, ",")
	case nil:
		value = ""
	default:
		value = ""
	}

	return
}

func (f *Form) ValidateOnSubmit(form *multipart.Form) error {
	space := regexp.MustCompile(`\s+`)

	employeeId := form.Value[f.EmployeeId.Name][0]
	fullName := strings.TrimSpace(form.Value[f.FullName.Name][0])
	location := strings.TrimSpace(form.Value[f.Location.Name][0])
	jobTitle := strings.TrimSpace(form.Value[f.JobTitle.Name][0])
	badges := []string{}

	f.EmployeeId.Data = employeeId
	if fullName != "" {
		f.FullName.Data = space.ReplaceAllString(fullName, " ")
	} else {
		return fmt.Errorf("'%s' field is expected", f.FullName.Label)
	}
	if location != "" {
		f.Location.Data = space.ReplaceAllString(location, " ")
	} else {
		return fmt.Errorf("'%s' field is expected", f.Location.Label)
	}
	if jobTitle != "" {
		f.JobTitle.Data = space.ReplaceAllString(jobTitle, " ")
	} else {
		return fmt.Errorf("'%s' field is expected", f.JobTitle.Label)
	}
	for _, b := range form.Value[f.Badges.Name] {
		v := strings.TrimSpace(b)
		_, exist := Badges[v]
		if exist {
			badges = append(badges, v)
		}
	}
	f.Badges.Data = badges

	return nil
}
