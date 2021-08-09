package db

import "github.com/boxyhq/hermes/types"

type DB interface {
	Ingest(tenantID string, logs []types.AuditLog) error
	Query(tenantID string, indexes map[string]string, start, end int64) ([]map[string]interface{}, error)
	Close()
}
