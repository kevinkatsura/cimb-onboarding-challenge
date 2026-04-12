package transaction

import (
	"context"
	"core-banking/pkg/logging"
)

type AuditService struct {
	repo Repository
}

func NewAuditService(repo Repository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, action string, entityType string, entityID string, actorID *string) {
	// In a real system, this would write to a specialized audit_logs table
	// For now, we utilize the repository (if supported) or structured logging

	logging.Ctx(ctx).Infow("audit_log",
		"action", action,
		"entity_type", entityType,
		"entity_id", entityID,
		"actor_id", actorID,
	)

	// Persist if model allows (using existing AuditLog model)
	// s.repo.SaveAuditLog(ctx, AuditLog{...})
	// (Note: we haven't added SaveAuditLog to repository interface yet,
	// but it's defined in models.go)
}
