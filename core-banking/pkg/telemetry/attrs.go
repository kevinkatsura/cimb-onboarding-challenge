package telemetry

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

// --- Attribute Key Constants ---
// Grouped by layer to avoid hardcoded strings across the codebase.

// Handler layer keys (HTTP)
const (
	KeyHTTPMethod    = attribute.Key("http.method")
	KeyHTTPRoute     = attribute.Key("http.route")
	KeyHTTPStatus    = attribute.Key("http.status_code")
	KeyHTTPUserAgent = attribute.Key("http.user_agent")
	KeyNetPeerIP     = attribute.Key("net.peer.ip")
	KeyRequestSize   = attribute.Key("request.size")
	KeyResponseSize  = attribute.Key("response.size")
	KeyRequestID     = attribute.Key("request.id")
	KeyUserID        = attribute.Key("user.id")
	KeyTenantID      = attribute.Key("tenant.id")
	KeySpanKind      = attribute.Key("span.kind")
)

// Service layer keys (Business)
const (
	KeyOperationName   = attribute.Key("operation.name")
	KeyBusinessDomain  = attribute.Key("business.domain")
	KeyBusinessUseCase = attribute.Key("business.use_case")
	KeyIdempotencyKey  = attribute.Key("idempotency.key")
	KeyRetryCount      = attribute.Key("retry.count")
)

// Repository layer keys
const (
	KeyRepoName   = attribute.Key("repository.name")
	KeyRepoOp     = attribute.Key("repository.operation")
	KeyEntityName = attribute.Key("entity.name")
	KeyEntityID   = attribute.Key("entity.id")
)

// Database layer keys
const (
	KeyDBSystem       = attribute.Key("db.system")
	KeyDBName         = attribute.Key("db.name")
	KeyDBOperation    = attribute.Key("db.operation")
	KeyDBStatement    = attribute.Key("db.statement")
	KeyDBRowsAffected = attribute.Key("db.rows_affected")
	KeyDBConnString   = attribute.Key("db.connection_string")
)

// Cache layer keys
const (
	KeyCacheSystem    = attribute.Key("cache.system")
	KeyCacheOperation = attribute.Key("cache.operation")
	KeyCacheKey       = attribute.Key("cache.key")
	KeyCacheHit       = attribute.Key("cache.hit")
	KeyCacheTTL       = attribute.Key("cache.ttl")
)

// --- Attribute Builders ---
// Each returns []attribute.KeyValue for direct use with span.SetAttributes(...).

// HandlerAttrs builds HTTP handler-level attributes from the request.
func HandlerAttrs(r *http.Request, route string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		KeyHTTPMethod.String(r.Method),
		KeyHTTPRoute.String(route),
		KeyHTTPUserAgent.String(r.UserAgent()),
		KeyNetPeerIP.String(r.RemoteAddr),
		KeyRequestSize.Int64(r.ContentLength),
		KeySpanKind.String("server"),
	}

	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		attrs = append(attrs, KeyRequestID.String(reqID))
	}
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		attrs = append(attrs, KeyUserID.String(userID))
	}
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		attrs = append(attrs, KeyTenantID.String(tenantID))
	}

	return attrs
}

// ResponseAttrs builds response-level attributes set after the handler completes.
func ResponseAttrs(statusCode int, responseSize int) []attribute.KeyValue {
	return []attribute.KeyValue{
		KeyHTTPStatus.Int(statusCode),
		KeyResponseSize.Int(responseSize),
	}
}

// ServiceAttrs builds service/business-layer span attributes.
func ServiceAttrs(operation, domain, useCase string) []attribute.KeyValue {
	return []attribute.KeyValue{
		KeyOperationName.String(operation),
		KeyBusinessDomain.String(domain),
		KeyBusinessUseCase.String(useCase),
	}
}

// ServiceAttrsWithIdempotency extends ServiceAttrs with idempotency and retry info.
func ServiceAttrsWithIdempotency(operation, domain, useCase, idempotencyKey string, retryCount int) []attribute.KeyValue {
	attrs := ServiceAttrs(operation, domain, useCase)
	if idempotencyKey != "" {
		attrs = append(attrs, KeyIdempotencyKey.String(idempotencyKey))
	}
	if retryCount > 0 {
		attrs = append(attrs, KeyRetryCount.Int(retryCount))
	}
	return attrs
}

// RepoAttrs builds repository-layer span attributes.
func RepoAttrs(repoName, operation, entityName, entityID string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		KeyRepoName.String(repoName),
		KeyRepoOp.String(operation),
		KeyEntityName.String(entityName),
	}
	if entityID != "" {
		attrs = append(attrs, KeyEntityID.String(entityID))
	}
	return attrs
}

// DBAttrs builds database-layer span attributes.
func DBAttrs(system, dbName, operation, statement string, rowsAffected int64) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		KeyDBSystem.String(system),
		KeyDBName.String(dbName),
		KeyDBOperation.String(operation),
	}
	if statement != "" {
		attrs = append(attrs, KeyDBStatement.String(statement))
	}
	if rowsAffected >= 0 {
		attrs = append(attrs, KeyDBRowsAffected.Int64(rowsAffected))
	}
	return attrs
}

// CacheAttrs builds cache/Redis-layer span attributes.
func CacheAttrs(system, operation, key string, hit bool, ttlSeconds int) []attribute.KeyValue {
	return []attribute.KeyValue{
		KeyCacheSystem.String(system),
		KeyCacheOperation.String(operation),
		KeyCacheKey.String(key),
		KeyCacheHit.Bool(hit),
		KeyCacheTTL.Int(ttlSeconds),
	}
}

// CacheAttrsNoHit builds cache attributes for write operations where hit is irrelevant.
func CacheAttrsNoHit(system, operation, key string, ttlSeconds int) []attribute.KeyValue {
	return []attribute.KeyValue{
		KeyCacheSystem.String(system),
		KeyCacheOperation.String(operation),
		KeyCacheKey.String(key),
		KeyCacheTTL.Int(ttlSeconds),
	}
}
