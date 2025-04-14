package models

type Role string

var ValidRoles = map[string]bool{
	"employee":  true,
	"moderator": true,
}

const (
	Moderator Role = "moderator"
	Employee  Role = "employee"
)
