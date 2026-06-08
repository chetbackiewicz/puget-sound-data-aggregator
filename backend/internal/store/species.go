package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Species struct {
	ID               int     `json:"id"`
	ScientificName   string  `json:"scientific_name"`
	CommonName       string  `json:"common_name"`
	WormsAphiaID     *int    `json:"worms_aphia_id,omitempty"`
	GBIFTaxonKey     *int    `json:"gbif_taxon_key,omitempty"`
	INatTaxonID      *int    `json:"inat_taxon_id,omitempty"`
	WikidataQID      *string `json:"wikidata_qid,omitempty"`
	FishbaseSpecCode *int    `json:"fishbase_speccode,omitempty"`
	SpearLegal       *bool   `json:"spear_legal,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

type SpeciesStore struct{ DB *sql.DB }

func (s *SpeciesStore) List(ctx context.Context) ([]Species, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, scientific_name, common_name, worms_aphia_id, gbif_taxon_key,
		       inat_taxon_id, wikidata_qid, fishbase_speccode, spear_legal, notes
		  FROM species ORDER BY common_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Species{}
	for rows.Next() {
		sp, err := scanSpecies(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *SpeciesStore) Get(ctx context.Context, id int) (Species, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, scientific_name, common_name, worms_aphia_id, gbif_taxon_key,
		       inat_taxon_id, wikidata_qid, fishbase_speccode, spear_legal, notes
		  FROM species WHERE id = ?`, id)
	return scanSpecies(row)
}

func (s *SpeciesStore) Create(ctx context.Context, sp Species) (int, error) {
	now := time.Now().UTC()
	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO species (scientific_name, common_name, worms_aphia_id, gbif_taxon_key,
		                     inat_taxon_id, wikidata_qid, fishbase_speccode, spear_legal, notes,
		                     created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sp.ScientificName, sp.CommonName, sp.WormsAphiaID, sp.GBIFTaxonKey,
		sp.INatTaxonID, sp.WikidataQID, sp.FishbaseSpecCode, sp.SpearLegal, sp.Notes,
		now, now)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (s *SpeciesStore) Update(ctx context.Context, sp Species) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE species SET scientific_name=?, common_name=?, worms_aphia_id=?, gbif_taxon_key=?,
		                   inat_taxon_id=?, wikidata_qid=?, fishbase_speccode=?, spear_legal=?,
		                   notes=?, updated_at=?
		 WHERE id=?`,
		sp.ScientificName, sp.CommonName, sp.WormsAphiaID, sp.GBIFTaxonKey,
		sp.INatTaxonID, sp.WikidataQID, sp.FishbaseSpecCode, sp.SpearLegal,
		sp.Notes, time.Now().UTC(), sp.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// scanner accepts both *sql.Row and *sql.Rows.
type scanner interface{ Scan(dest ...any) error }

func scanSpecies(s scanner) (Species, error) {
	var sp Species
	err := s.Scan(&sp.ID, &sp.ScientificName, &sp.CommonName, &sp.WormsAphiaID,
		&sp.GBIFTaxonKey, &sp.INatTaxonID, &sp.WikidataQID,
		&sp.FishbaseSpecCode, &sp.SpearLegal, &sp.Notes)
	if errors.Is(err, sql.ErrNoRows) {
		return sp, sql.ErrNoRows
	}
	return sp, err
}
