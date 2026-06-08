-- Initial schema for Puget Sound Fishing App

CREATE TABLE IF NOT EXISTS api_probes (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  source_key VARCHAR(64) NOT NULL,
  endpoint_url TEXT NOT NULL,
  http_status INT,
  duration_ms BIGINT,
  ok BOOLEAN NOT NULL,
  error_message TEXT,
  raw_response_snippet MEDIUMTEXT,
  parsed_summary JSON,
  fetched_at DATETIME(6) NOT NULL,
  INDEX idx_source_fetched (source_key, fetched_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS marine_areas (
  id INT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  geometry GEOMETRY NOT NULL SRID 4326,
  updated_at DATETIME(6) NOT NULL,
  SPATIAL INDEX idx_geom (geometry)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS species (
  id INT AUTO_INCREMENT PRIMARY KEY,
  scientific_name VARCHAR(128) NOT NULL UNIQUE,
  common_name VARCHAR(128) NOT NULL,
  worms_aphia_id INT,
  gbif_taxon_key INT,
  inat_taxon_id INT,
  wikidata_qid VARCHAR(16),
  fishbase_speccode INT,
  spear_legal BOOLEAN,
  notes TEXT,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS regulations (
  id INT AUTO_INCREMENT PRIMARY KEY,
  marine_area_id INT NOT NULL,
  species_id INT NOT NULL,
  season_open DATE,
  season_close DATE,
  daily_limit INT,
  size_min_inches DECIMAL(5,2),
  size_max_inches DECIMAL(5,2),
  gear_restrictions TEXT,
  is_emergency_rule BOOLEAN NOT NULL DEFAULT FALSE,
  emergency_rule_expires DATE,
  wac_citation VARCHAR(64),
  source_doc VARCHAR(256),
  notes TEXT,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  CONSTRAINT fk_regulations_marine_area FOREIGN KEY (marine_area_id) REFERENCES marine_areas(id),
  CONSTRAINT fk_regulations_species FOREIGN KEY (species_id) REFERENCES species(id),
  INDEX idx_regulations_area_species (marine_area_id, species_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS techniques (
  id INT AUTO_INCREMENT PRIMARY KEY,
  species_id INT NOT NULL,
  title VARCHAR(256) NOT NULL,
  method VARCHAR(64),
  recommended_tackle TEXT,
  recommended_bait TEXT,
  best_season VARCHAR(64),
  best_time_of_day VARCHAR(64),
  best_conditions TEXT,
  marine_areas VARCHAR(64),
  body_markdown MEDIUMTEXT,
  author VARCHAR(128),
  source_attribution TEXT,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  CONSTRAINT fk_techniques_species FOREIGN KEY (species_id) REFERENCES species(id),
  INDEX idx_techniques_species_method (species_id, method)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS stations (
  id INT AUTO_INCREMENT PRIMARY KEY,
  provider VARCHAR(32) NOT NULL,
  external_id VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  kind VARCHAR(32),
  lat DECIMAL(9,6),
  lon DECIMAL(9,6),
  UNIQUE KEY uq_provider_id (provider, external_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
