// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package usefunctionsyntaxforexecutioncontext

import (
	"fmt"
	"io"
	"strconv"
)

type Entity interface {
	IsEntity()
	GetID() string
	GetCreatedAt() *string
}

type Admin struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	CreatedAt   *string  `json:"createdAt,omitempty"`
}

func (Admin) IsEntity()                  {}
func (this Admin) GetID() string         { return this.ID }
func (this Admin) GetCreatedAt() *string { return this.CreatedAt }

type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   *int   `json:"age,omitempty"`
	Role  *Role  `json:"role,omitempty"`
}

type Mutation struct {
}

type MutationResponse struct {
	Success bool    `json:"success"`
	Message *string `json:"message,omitempty"`
}

type Query struct {
}

type Subscription struct {
}

type User struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Age       *int    `json:"age,omitempty"`
	Role      Role    `json:"role"`
	CreatedAt *string `json:"createdAt,omitempty"`
}

func (User) IsEntity()                  {}
func (this User) GetID() string         { return this.ID }
func (this User) GetCreatedAt() *string { return this.CreatedAt }

type UserFilter struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Age      *int    `json:"age,omitempty"`
	Roles    []Role  `json:"roles,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
}

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
	RoleGuest Role = "GUEST"
)

var AllRole = []Role{
	RoleAdmin,
	RoleUser,
	RoleGuest,
}

func (e Role) IsValid() bool {
	switch e {
	case RoleAdmin, RoleUser, RoleGuest:
		return true
	}
	return false
}

func (e Role) String() string {
	return string(e)
}

func (e *Role) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Role(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Role", str)
	}
	return nil
}

func (e Role) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
