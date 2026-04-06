package CSOutbountAdapter

import (
	"context"
	"database/sql"
	"gRPCbigapp/App/ClientService/CSDomain"
)

type CSPostgresAdapter struct {
	database *sql.DB
}

func (pa *CSPostgresAdapter) SaveUserInDB(ctx context.Context, user *CSDomain.User) error {
	_, err := pa.database.ExecContext(ctx,
		"INSERT INTO users_data (user_id, user_name, user_password, user_role), VALUES $1, $2, $3, $4",
		user.UserID, user.UserName, user.UserPassword, user.UserRole)
	if err != nil {
		return err // TODO formulate acid and add ZAP + recheck for make better
	}
	return err
}

func (pa *CSPostgresAdapter) GetUserFromDB(ctx context.Context, UserID string) (*CSDomain.User, error) {
	row := pa.database.QueryRowContext(ctx, "SELECT user_id, user_name FROM users_data WHERE user_id = $1", UserID)

	var user CSDomain.User
	err := row.Scan(&user.UserID, &user.UserName)
	if err != nil {
		return nil, err // TODO ZAP
	}
	return &user, nil
}

func (pa *CSPostgresAdapter) UpdateUserPlan(ctx context.Context, user *CSDomain.User) error {
	_, err := pa.database.ExecContext(ctx,
		"UPDATE user_data SET user_role = $1 WHERE user_id = $2", user.UserRole, user.UserID)
	if err != nil {
		return err // TODO ZAP + ROLLBACK
	}
	return nil
}

func (pa *CSPostgresAdapter) CheckIsAdmin(ctx context.Context, admin string) (*CSDomain.IsAdmin, error) {
	check := pa.database.QueryRowContext(ctx,
		"SELECT user_id FROM users_data WHERE user_id = $1", admin)

	var adm CSDomain.IsAdmin
	err := check.Scan(&adm.UserID)
	if err != nil {
		return nil, err // TODO ZAP
	}
	return &adm, nil
}
