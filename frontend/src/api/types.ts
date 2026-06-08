export type ProbeResult = {
  source_key: string;
  endpoint_url: string;
  http_status: number;
  duration_ms: number;
  ok: boolean;
  error_message?: string;
  raw_response_snippet?: string;
  parsed_summary?: unknown;
  fetched_at: string;
};

export type SourceInfo = {
  key: string;
  description: string;
  auth_required: boolean;
  last_probe?: ProbeResult;
};

export type MarineArea = { id: number; name: string };

export type Species = {
  id: number;
  scientific_name: string;
  common_name: string;
  worms_aphia_id?: number;
  gbif_taxon_key?: number;
  inat_taxon_id?: number;
  wikidata_qid?: string;
  fishbase_speccode?: number;
  spear_legal?: boolean;
  notes?: string;
};

export type Regulation = {
  id: number;
  marine_area_id: number;
  species_id: number;
  season_open?: string;
  season_close?: string;
  daily_limit?: number;
  size_min_inches?: number;
  size_max_inches?: number;
  gear_restrictions?: string;
  is_emergency_rule?: boolean;
  emergency_rule_expires?: string;
  wac_citation?: string;
  source_doc?: string;
  notes?: string;
};

export type Technique = {
  id: number;
  species_id: number;
  title: string;
  method?: string;
  recommended_tackle?: string;
  recommended_bait?: string;
  best_season?: string;
  best_time_of_day?: string;
  best_conditions?: string;
  marine_areas?: string;
  body_markdown?: string;
  author?: string;
  source_attribution?: string;
};
