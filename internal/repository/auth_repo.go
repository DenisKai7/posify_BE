package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) CreateTenant(ctx context.Context, t *model.Tenant) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO tenants (name, address, phone) VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		t.Name, t.Address, t.Phone,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *AuthRepo) CreateUser(ctx context.Context, u *model.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (tenant_id, name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		u.TenantID, u.Name, u.Email, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *AuthRepo) GetUserByEmail(ctx context.Context, tenantID, email string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, email, password_hash, role, created_at, updated_at
		 FROM users WHERE tenant_id = $1 AND email = $2`,
		tenantID, email,
	).Scan(&u.ID, &u.TenantID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *AuthRepo) GetTenantByName(ctx context.Context, name string) (*model.Tenant, error) {
	t := &model.Tenant{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, address, phone, created_at, updated_at
		 FROM tenants WHERE name = $1`,
		name,
	).Scan(&t.ID, &t.Name, &t.Address, &t.Phone, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return t, nil
}
