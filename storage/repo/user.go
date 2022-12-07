package repo

import "time"

const (
	UserTypeSuperAdmin = "superadmin"
	UserTypeUser       = "user"
)

type User struct {
	ID              int64     `db:"id"`
	FirstName       string    `db:"first_name"`
	LastName        string    `db:"last_name"`
	PhoneNumber     *string   `db:"phone_number"`
	Email           string    `db:"email"`
	Gender          *string   `db:"gender"`
	Password        string    `db:"password"`
	Username        *string   `db:"username"`
	ProfileImageUrl *string   `db:"profile_image_url"`
	Type            string    `db:"type"`
	CreatedAt       time.Time `db:"created_at"`
}

type GetUsersParams struct {
	Limit  int32  `db:"limit"`
	Page   int32  `db:"page"`
	Search string `db:"search"`
}

type GetUsersResult struct {
	Users []*User `db:"users"`
	Count int32   `db:"count"`
}

type UpdatePassword struct {
	UserID   int64  `db:"user_id"`
	Password string `db:"password"`
}

type UserStorageI interface {
	Create(user *User) (*User, error)
	Get(id int64) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll(params *GetUsersParams) (*GetUsersResult, error)
	Update(user *User) error
	Delete(id int64) error
	UpdatePassword(req *UpdatePassword) error
}
