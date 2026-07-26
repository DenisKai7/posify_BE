package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type SyncService struct {
	repo *repository.SyncRepo
	pool *pgxpool.Pool
}

func NewSyncService(repo *repository.SyncRepo, pool *pgxpool.Pool) *SyncService {
	return &SyncService{repo: repo, pool: pool}
}

// Push processes a batch of offline transactions idempotently.
// Never fails the entire batch for one bad transaction.
func (s *SyncService) Push(ctx context.Context, tenantID, userID string, req model.SyncPushRequest) (*model.SyncPushResponse, error) {
	resp := &model.SyncPushResponse{}
	var dupeCount int

	for _, ot := range req.Transactions {
		pgxTx, err := s.pool.Begin(ctx)
		if err != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: begin tx: %v", ot.OfflineTransactionID, err))
			continue
		}

		inserted, err := s.repo.SyncTransaction(ctx, pgxTx, tenantID, ot)
		if err != nil {
			_ = pgxTx.Rollback(ctx)
			log.Printf("sync push error [%s]: %v", ot.OfflineTransactionID, err)
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: %v", ot.OfflineTransactionID, err))
			continue
		}

		if err := pgxTx.Commit(ctx); err != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: commit: %v", ot.OfflineTransactionID, err))
			continue
		}

		if !inserted {
			dupeCount++
		}
		resp.SyncedIDs = append(resp.SyncedIDs, ot.OfflineTransactionID)
	}

	// Audit log — fire-and-forget
	go func() {
		if err := s.repo.InsertSyncLog(context.Background(), model.SyncLog{
			TenantID:       tenantID,
			UserID:         userID,
			BatchSize:      len(req.Transactions),
			SyncedCount:    len(resp.SyncedIDs),
			DuplicateCount: dupeCount,
			ErrorCount:     len(resp.Errors),
			Errors:         resp.Errors,
		}); err != nil {
			log.Printf("sync log insert error: %v", err)
		}
	}()

	return resp, nil
}

// Pull returns current product catalog for a tenant.
// If lastSyncedAt is provided, returns only products updated after that time (delta sync / LWW).
func (s *SyncService) Pull(ctx context.Context, tenantID string, lastSyncedAt *time.Time) (*model.SyncPullResponse, error) {
	products, err := s.repo.GetProducts(ctx, tenantID, lastSyncedAt)
	if err != nil {
		return nil, err
	}
	if products == nil {
		products = []model.SyncProduct{}
	}
	return &model.SyncPullResponse{Products: products}, nil
}
