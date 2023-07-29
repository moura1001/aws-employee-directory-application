package model

type Employee struct {
	Id       string
	Photo    *Photo
	FullName string
	Location string
	JobTitle string
	Badges   []string
}

type Photo struct {
	ObjectKey string
	SignedUrl string
}

func (e Employee) HasBadge(badge string) bool {
	for _, b := range e.Badges {
		if b == badge {
			return true
		}
	}

	return false
}
