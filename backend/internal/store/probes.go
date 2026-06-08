// Package store provides persistence helpers for the various tables.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type ProbeStore struct{ DB *sql.DB }

// Save inserts a ProbeResult into api_probes.
func (s *ProbeStore) Save(ctx context.Context, r sources.ProbeResult) (int64, error) {
	var parsed any
	if len(r.ParsedSummary) > 0 {
		parsed = string(r.ParsedSummary)
	}
	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO api_probes
		  (source_key, endpoint_url, http_status, duration_ms, ok,
		   error_message, raw_response_snippet, parsed_summary, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.SourceKey, r.EndpointURL, r.HTTPStatus, r.DurationMS, r.OK,
		nullStr(r.ErrorMessage), nullStr(r.RawResponseSnippet), parsed, r.FetchedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Latest returns the most recent probe for a source_key, or sql.ErrNoRows.
func (s *ProbeStore) Latest(ctx context.Context, key string) (sources.ProbeResult, error) {
	var r sources.ProbeResult
	var (
		errMsg, raw sql.NullString
		parsed      sql.NullString
		fetchedAt   time.Time
	)
	err := s.DB.QueryRowContext(ctx, `
		SELECT source_key, endpoint_url, http_status, duration_ms, ok,
		       error_message, raw_response_snippet, parsed_summary, fetched_at
		  FROM api_probes
		 WHERE source_key = ?
		 ORDER BY fetched_at DESC
		 LIMIT 1`, key,
	).Scan(&r.SourceKey, &r.EndpointURL, &r.HTTPStatus, &r.DurationMS, &r.OK,
		&errMsg, &raw, &parsed, &fetchedAt)
	if err != nil {
		return r, err
	}
	r.ErrorMessage = errMsg.String
	r.RawResponseSnippet = raw.String
	if parsed.Valid && parsed.String != "" {
		r.ParsedSummary = json.RawMessage(parsed.String)
	}
	r.FetchedAt = fetchedAt
	return r, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
