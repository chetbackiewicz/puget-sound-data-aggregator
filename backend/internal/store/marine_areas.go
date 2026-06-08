package store

import (
	"context"
	"database/sql"
	"encoding/json"
)

type MarineArea struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Geometry json.RawMessage `json:"geometry,omitempty"`
}

type MarineAreaStore struct{ DB *sql.DB }

func (s *MarineAreaStore) List(ctx context.Context) ([]MarineArea, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, ST_AsGeoJSON(geometry) FROM marine_areas ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MarineArea{}
	for rows.Next() {
		var m MarineArea
		var geojson sql.NullString
		if err := rows.Scan(&m.ID, &m.Name, &geojson); err != nil {
			return nil, err
		}
		if geojson.Valid {
			m.Geometry = json.RawMessage(geojson.String)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *MarineAreaStore) Get(ctx context.Context, id int) (MarineArea, error) {
	var m MarineArea
	var geojson sql.NullString
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, ST_AsGeoJSON(geometry) FROM marine_areas WHERE id = ?`, id,
	).Scan(&m.ID, &m.Name, &geojson)
	if err != nil {
		return m, err
	}
	if geojson.Valid {
		m.Geometry = json.RawMessage(geojson.String)
	}
	return m, nil
}
