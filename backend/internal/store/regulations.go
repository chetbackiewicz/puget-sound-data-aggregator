package store

import (
	"context"
	"database/sql"
	"time"
)

type Regulation struct {
	ID                   int      `json:"id"`
	MarineAreaID         int      `json:"marine_area_id"`
	SpeciesID            int      `json:"species_id"`
	SeasonOpen           *string  `json:"season_open,omitempty"`
	SeasonClose          *string  `json:"season_close,omitempty"`
	DailyLimit           *int     `json:"daily_limit,omitempty"`
	SizeMinInches        *float64 `json:"size_min_inches,omitempty"`
	SizeMaxInches        *float64 `json:"size_max_inches,omitempty"`
	GearRestrictions     *string  `json:"gear_restrictions,omitempty"`
	IsEmergencyRule      bool     `json:"is_emergency_rule"`
	EmergencyRuleExpires *string  `json:"emergency_rule_expires,omitempty"`
	WACCitation          *string  `json:"wac_citation,omitempty"`
	SourceDoc            *string  `json:"source_doc,omitempty"`
	Notes                *string  `json:"notes,omitempty"`
}

type RegulationStore struct{ DB *sql.DB }

func (s *RegulationStore) List(ctx context.Context, marineAreaID, speciesID int) ([]Regulation, error) {
	q := `SELECT id, marine_area_id, species_id, season_open, season_close, daily_limit,
	             size_min_inches, size_max_inches, gear_restrictions, is_emergency_rule,
	             emergency_rule_expires, wac_citation, source_doc, notes
	        FROM regulations WHERE 1=1`
	args := []any{}
	if marineAreaID > 0 {
		q += " AND marine_area_id = ?"
		args = append(args, marineAreaID)
	}
	if speciesID > 0 {
		q += " AND species_id = ?"
		args = append(args, speciesID)
	}
	q += " ORDER BY marine_area_id, species_id, id"
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Regulation{}
	for rows.Next() {
		var r Regulation
		var so, sc, ere sql.NullString
		if err := rows.Scan(&r.ID, &r.MarineAreaID, &r.SpeciesID, &so, &sc, &r.DailyLimit,
			&r.SizeMinInches, &r.SizeMaxInches, &r.GearRestrictions, &r.IsEmergencyRule,
			&ere, &r.WACCitation, &r.SourceDoc, &r.Notes); err != nil {
			return nil, err
		}
		if so.Valid {
			v := so.String
			r.SeasonOpen = &v
		}
		if sc.Valid {
			v := sc.String
			r.SeasonClose = &v
		}
		if ere.Valid {
			v := ere.String
			r.EmergencyRuleExpires = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *RegulationStore) Create(ctx context.Context, r Regulation) (int, error) {
	now := time.Now().UTC()
	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO regulations (marine_area_id, species_id, season_open, season_close, daily_limit,
		                         size_min_inches, size_max_inches, gear_restrictions, is_emergency_rule,
		                         emergency_rule_expires, wac_citation, source_doc, notes,
		                         created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.MarineAreaID, r.SpeciesID, r.SeasonOpen, r.SeasonClose, r.DailyLimit,
		r.SizeMinInches, r.SizeMaxInches, r.GearRestrictions, r.IsEmergencyRule,
		r.EmergencyRuleExpires, r.WACCitation, r.SourceDoc, r.Notes, now, now)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (s *RegulationStore) Update(ctx context.Context, r Regulation) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE regulations SET marine_area_id=?, species_id=?, season_open=?, season_close=?, daily_limit=?,
		                       size_min_inches=?, size_max_inches=?, gear_restrictions=?, is_emergency_rule=?,
		                       emergency_rule_expires=?, wac_citation=?, source_doc=?, notes=?, updated_at=?
		 WHERE id=?`,
		r.MarineAreaID, r.SpeciesID, r.SeasonOpen, r.SeasonClose, r.DailyLimit,
		r.SizeMinInches, r.SizeMaxInches, r.GearRestrictions, r.IsEmergencyRule,
		r.EmergencyRuleExpires, r.WACCitation, r.SourceDoc, r.Notes, time.Now().UTC(), r.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *RegulationStore) Delete(ctx context.Context, id int) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM regulations WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
