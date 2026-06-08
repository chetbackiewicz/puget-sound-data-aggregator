package store

import (
	"context"
	"database/sql"
	"time"
)

type Technique struct {
	ID                int     `json:"id"`
	SpeciesID         int     `json:"species_id"`
	Title             string  `json:"title"`
	Method            *string `json:"method,omitempty"`
	RecommendedTackle *string `json:"recommended_tackle,omitempty"`
	RecommendedBait   *string `json:"recommended_bait,omitempty"`
	BestSeason        *string `json:"best_season,omitempty"`
	BestTimeOfDay     *string `json:"best_time_of_day,omitempty"`
	BestConditions    *string `json:"best_conditions,omitempty"`
	MarineAreas       *string `json:"marine_areas,omitempty"`
	BodyMarkdown      *string `json:"body_markdown,omitempty"`
	Author            *string `json:"author,omitempty"`
	SourceAttribution *string `json:"source_attribution,omitempty"`
}

type TechniqueStore struct{ DB *sql.DB }

func (s *TechniqueStore) List(ctx context.Context, speciesID int, method string) ([]Technique, error) {
	q := `SELECT id, species_id, title, method, recommended_tackle, recommended_bait,
	             best_season, best_time_of_day, best_conditions, marine_areas, body_markdown,
	             author, source_attribution
	        FROM techniques WHERE 1=1`
	args := []any{}
	if speciesID > 0 {
		q += " AND species_id = ?"
		args = append(args, speciesID)
	}
	if method != "" {
		q += " AND method = ?"
		args = append(args, method)
	}
	q += " ORDER BY species_id, method, id"
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Technique{}
	for rows.Next() {
		var t Technique
		if err := rows.Scan(&t.ID, &t.SpeciesID, &t.Title, &t.Method, &t.RecommendedTackle,
			&t.RecommendedBait, &t.BestSeason, &t.BestTimeOfDay, &t.BestConditions,
			&t.MarineAreas, &t.BodyMarkdown, &t.Author, &t.SourceAttribution); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *TechniqueStore) Create(ctx context.Context, t Technique) (int, error) {
	now := time.Now().UTC()
	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO techniques (species_id, title, method, recommended_tackle, recommended_bait,
		                        best_season, best_time_of_day, best_conditions, marine_areas,
		                        body_markdown, author, source_attribution, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.SpeciesID, t.Title, t.Method, t.RecommendedTackle, t.RecommendedBait,
		t.BestSeason, t.BestTimeOfDay, t.BestConditions, t.MarineAreas,
		t.BodyMarkdown, t.Author, t.SourceAttribution, now, now)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (s *TechniqueStore) Update(ctx context.Context, t Technique) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE techniques SET species_id=?, title=?, method=?, recommended_tackle=?, recommended_bait=?,
		                      best_season=?, best_time_of_day=?, best_conditions=?, marine_areas=?,
		                      body_markdown=?, author=?, source_attribution=?, updated_at=?
		 WHERE id=?`,
		t.SpeciesID, t.Title, t.Method, t.RecommendedTackle, t.RecommendedBait,
		t.BestSeason, t.BestTimeOfDay, t.BestConditions, t.MarineAreas,
		t.BodyMarkdown, t.Author, t.SourceAttribution, time.Now().UTC(), t.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *TechniqueStore) Delete(ctx context.Context, id int) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM techniques WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
