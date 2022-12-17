package postgres

import (
	"database/sql"
	"errors"

	"github.com/ibrat-muslim/blog_app_user_service/storage/repo"
	"github.com/jmoiron/sqlx"
)

type permissionRepo struct {
	db *sqlx.DB
}

func NewPermission(db *sqlx.DB) repo.PermissionStorageI {
	return &permissionRepo{
		db: db,
	}
}

func (pr *permissionRepo) CheckPermission(userType, resource, action string) (bool, error) {
	query := `
		SELECT id FROM permissions 
		WHERE user_type = $1 AND resource = $2 AND action = $3
	`

	var id int64
	err := pr.db.QueryRow(query, userType, resource, action).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
