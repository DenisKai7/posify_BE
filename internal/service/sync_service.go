package service

import (
	"context"
	"fmt"
	"log"

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
func (s *SyncService) Push(ctx context.Context, tenantID string, req model.SyncPushRequest) (*model.SyncPushResponse, error) {
	resp := &model.SyncPushResponse{}

	for _, ot := range req.Transactions {
		// Each transaction in its own DB transaction for isolation
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

		// Both new inserts and skipped duplicates count as "synced"
		_ = inserted // new=true, dup=false — both are OK
		resp.SyncedIDs = append(resp.SyncedIDs, ot.OfflineTransactionID)
	}

	return resp, nil
}

// Pull returns current product catalog for a tenant
func (s *SyncService) Pull(ctx context.Context, tenantID string) (*model.SyncPullResponse, error) {
	products, err := s.repo.GetProducts(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if products == nil {
		products = []model.SyncProduct{} // never null in JSON
	}
	return &model.SyncPullResponse{Products: products}, nil
}
