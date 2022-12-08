package repo

import "time"

const (
	UserTypeSuperAdmin = "superadmin"
	UserTypeUser       = "user"
)

type User struct {
	ID              int64
	FirstName       string
	LastName        string
	PhoneNumber     string
	Email           string
	Gender          string
	Password        string
	Username        string
	ProfileImageUrl string
	Type            string
	CreatedAt       time.Time
}

type GetUsersParams struct {
	Limit  int32
	Page   int32
	Search string
}

type GetUsersResult struct {
	Users []*User
	Count int32
}

type UpdatePassword struct {
	UserID   int64
	Password string
}

type UserStorageI interface {
	Create(user *User) (*User, error)
	Get(id int64) (*User, error)
	GetAll(params *GetUsersParams) (*GetUsersResult, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) (*User, error)
	UpdatePassword(req *UpdatePassword) error
	Delete(id int64) error
}
