package PostgresAdapter

import (
	"context"
	"errors"
	"fmt"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	tracing "gRPCbigapp/Shared/Tracing"
	"gRPCbigapp/Shared/Txmanager"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/codes"
)

var trace = tracing.Tracer("db.user_repo")

var _ CSPorts.CSOutboundPorts = (*UserRepo)(nil)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (rc *UserRepo) connection(ctx context.Context) dbExecutor {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return rc.pool
}

func (rc *UserRepo) SaveUser(ctx context.Context, user *CSDomain.User) error {
	const query = `
		INSERT INTO users_data (user_id, user_name, user_password, user_role)
		VALUES ($1, $2, $3, $4)`

	ctx, span := trace.Start(ctx, "db.SaveUser", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	_, err := rc.connection(ctx).Exec(ctx, query, user.UserID, user.UserName, user.UserPassword, string(user.UserRole))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db.SaveUser failed")
		return fmt.Errorf("postgres, error saving user: %w", err)
	}
	return nil
}

func (rc *UserRepo) GetUser(ctx context.Context, userID string) (*CSDomain.User, error) {
	const query = `
		SELECT user_id, user_name, user_password, user_role
		FROM users_data WHERE user_id = $1`
	row := rc.connection(ctx).QueryRow(ctx, query, userID)

	ctx, span := trace.Start(ctx, "db.GetUser", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	var user CSDomain.User
	var role string
	if err := row.Scan(&user.UserID, &user.UserName, &user.UserPassword, &role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, CSDomain.ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db.GetUser failed")
		return nil, fmt.Errorf("postgres, error find user: %w", err)
	}
	user.UserRole = CSDomain.UserPlan(user.UserRole)
	return &user, nil
}

func (rc *UserRepo) UpdateUserPlan(ctx context.Context, userID string, userPlan CSDomain.UserPlan) error {
	const query = `UPDATE users_data SET user_role = $1 WHERE user_id = $2`

	ctx, span := trace.Start(ctx, "db.UpdateUserPlan", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	_, err := rc.connection(ctx).Exec(ctx, query, userID, string(userPlan))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db.UpdateUserPlan failed")
		return fmt.Errorf("postgres, error updating user plan: %w", err)
	}
	return nil
}

func (rc *UserRepo) IsAdmin(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT user_role FROM users_data WHERE user_id = $1`

	ctx, span := trace.Start(ctx, "db.IsAdmin", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	row := rc.connection(ctx).QueryRow(ctx, query, userID)

	var role string
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.AddEvent("user_not_found")
			return false, CSDomain.ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db.IsAdmin failed")
		return false, fmt.Errorf("postgres, error isAdmin: %w", err)
	}
	return CSDomain.UserPlan(role) == CSDomain.AdminRole, nil
}
