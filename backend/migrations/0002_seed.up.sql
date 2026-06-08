-- Seed Washington Marine Areas (placeholder bounding-box polygons; refined when
-- the WDFW ArcGIS fetcher runs and replaces these with authoritative geometry).
-- Boxes are intentionally rough; exact boundaries come from WDFW.

SET @now = UTC_TIMESTAMP(6);

INSERT INTO marine_areas (id, name, geometry, updated_at) VALUES
 (5,  'Sekiu-Pillar Point',         ST_GeomFromText('POLYGON((-124.5 48.2,-123.9 48.2,-123.9 48.5,-124.5 48.5,-124.5 48.2))', 4326, 'axis-order=long-lat'), @now),
 (6,  'East Juan de Fuca',          ST_GeomFromText('POLYGON((-123.9 48.0,-123.0 48.0,-123.0 48.4,-123.9 48.4,-123.9 48.0))', 4326, 'axis-order=long-lat'), @now),
 (7,  'San Juan Islands',           ST_GeomFromText('POLYGON((-123.3 48.4,-122.6 48.4,-122.6 48.8,-123.3 48.8,-123.3 48.4))', 4326, 'axis-order=long-lat'), @now),
 (8,  'Deception Pass-Camano',      ST_GeomFromText('POLYGON((-122.8 48.0,-122.2 48.0,-122.2 48.5,-122.8 48.5,-122.8 48.0))', 4326, 'axis-order=long-lat'), @now),
 (9,  'Admiralty Inlet',            ST_GeomFromText('POLYGON((-122.8 47.7,-122.3 47.7,-122.3 48.2,-122.8 48.2,-122.8 47.7))', 4326, 'axis-order=long-lat'), @now),
 (10, 'Seattle-Bremerton',          ST_GeomFromText('POLYGON((-122.7 47.4,-122.2 47.4,-122.2 47.8,-122.7 47.8,-122.7 47.4))', 4326, 'axis-order=long-lat'), @now),
 (11, 'Tacoma-Vashon',              ST_GeomFromText('POLYGON((-122.7 47.1,-122.2 47.1,-122.2 47.5,-122.7 47.5,-122.7 47.1))', 4326, 'axis-order=long-lat'), @now),
 (12, 'Hood Canal',                 ST_GeomFromText('POLYGON((-123.2 47.3,-122.6 47.3,-122.6 47.9,-123.2 47.9,-123.2 47.3))', 4326, 'axis-order=long-lat'), @now),
 (13, 'South Puget Sound',          ST_GeomFromText('POLYGON((-123.0 47.0,-122.4 47.0,-122.4 47.3,-123.0 47.3,-123.0 47.0))', 4326, 'axis-order=long-lat'), @now);

INSERT INTO species
 (scientific_name, common_name, worms_aphia_id, gbif_taxon_key, inat_taxon_id, wikidata_qid, fishbase_speccode, spear_legal, notes, created_at, updated_at)
VALUES
 ('Ophiodon elongatus',       'Lingcod',          240745, 2336521, 52539, 'Q81964', 167116, TRUE,
  'Top spearfishing target. Season + area restrictions apply per WAC.', @now, @now),
 ('Scorpaenichthys marmoratus','Cabezon',         282726, NULL,    47638, NULL,     NULL,   TRUE,
  'Spear-legal; check size limit per WAC 220-316.', @now, @now),
 ('Hexagrammos decagrammus',  'Kelp greenling',   240732, NULL,    NULL,  NULL,     NULL,   TRUE,
  'Common in shallow kelp; spear-legal.', @now, @now),
 ('Sebastes caurinus',        'Copper rockfish',  274780, NULL,    NULL,  NULL,     NULL,   FALSE,
  'Most PS rockfish under permanent or emergency closures (WAC 220-338).', @now, @now),
 ('Sebastes melanops',        'Black rockfish',   NULL,   NULL,    NULL,  NULL,     NULL,   FALSE,
  'Closed in much of inner Puget Sound.', @now, @now),
 ('Oncorhynchus tshawytscha', 'Chinook salmon',   158075, NULL,    NULL,  NULL,     NULL,   FALSE,
  'Salmon are NOT spear-legal (WAC 220-310-130).', @now, @now),
 ('Oncorhynchus kisutch',     'Coho salmon',      NULL,   NULL,    NULL,  NULL,     NULL,   FALSE,
  'Salmon are NOT spear-legal.', @now, @now),
 ('Oncorhynchus gorbuscha',   'Pink salmon',      NULL,   NULL,    NULL,  NULL,     NULL,   FALSE,
  'Odd-year runs in PS; not spear-legal.', @now, @now),
 ('Oncorhynchus keta',        'Chum salmon',      NULL,   NULL,    NULL,  NULL,     NULL,   FALSE,
  'Hood Canal summer + fall runs; not spear-legal.', @now, @now),
 ('Oncorhynchus nerka',       'Sockeye salmon',   NULL,   NULL,    NULL,  NULL,     NULL,   FALSE,
  'Baker Lake / Lake Wenatchee stocks; not spear-legal.', @now, @now),
 ('Hippoglossus stenolepis',  'Pacific halibut',  NULL,   NULL,    NULL,  NULL,     NULL,   TRUE,
  'IPHC-managed quotas; short seasons in PS marine areas.', @now, @now);

INSERT INTO stations (provider, external_id, name, kind, lat, lon) VALUES
 ('coops', '9447130',          'Seattle',                'tide',     47.602600, -122.339300),
 ('coops', '9446484',          'Tacoma',                 'tide',     47.266700, -122.413300),
 ('coops', '9444900',          'Port Townsend',          'tide',     48.111200, -122.759700),
 ('coops', '9449880',          'Friday Harbor',          'tide',     48.545300, -123.012500),
 ('coops', '9444090',          'Port Angeles',           'tide',     48.125000, -123.440000),
 ('coops', '9443090',          'Neah Bay',               'tide',     48.370700, -124.601600),
 ('coops', 'PUG1501',          'Agate Passage (S end)',  'current',  47.711000, -122.567100),
 ('coops', 'PUG1510',          'Port Washington Narrows','current',  47.579600, -122.630700),
 ('coops', 'PUG1508',          'Liberty Bay entrance',   'current',  47.706800, -122.628300),
 ('ndbc',  'WPOW1',            'West Point',             'wind',     47.660000, -122.440000),
 ('ndbc',  'SISW1',            'Smith Island',           'wind',     48.320000, -122.840000),
 ('ndbc',  'PTAW1',            'Port Angeles Harbor',    'wind',     48.120000, -123.440000),
 ('ndbc',  '46087',            'Neah Bay offshore',      'buoy',     48.490000, -124.730000),
 ('ndbc',  '46088',            'New Dungeness',          'buoy',     48.330000, -123.170000),
 ('usgs',  '12101500',         'Puyallup R at Puyallup', 'river',    47.208000, -122.327000),
 ('usgs',  '12089500',         'Nisqually R at McKenna', 'river',    46.933000, -122.561000),
 ('usgs',  '12061500',         'Skokomish R nr Potlatch','river',    47.310000, -123.177000),
 ('usgs',  '12150800',         'Snohomish R nr Monroe',  'river',    47.831000, -122.048000),
 ('usgs',  '12167000',         'NF Stillaguamish',       'river',    48.261000, -122.048000),
 ('usgs',  '12181000',         'Skagit R at Marblemount','river',    48.534000, -121.430000),
 ('usgs',  '12200500',         'Skagit R nr Mt. Vernon', 'river',    48.445000, -122.335000),
 ('usgs',  '12048000',         'Dungeness R nr Sequim',  'river',    48.014000, -123.133000),
 ('orca',  'orca_hydro_carrinlet','Carr Inlet (ORCA)',   'buoy',     47.280000, -122.730000),
 ('orca',  'orca_met_twanoh',  'Twanoh (ORCA)',          'buoy',     47.375000, -123.008000),
 ('orca',  'orca_hydro_hansville','Hansville (ORCA)',    'buoy',     47.908000, -122.628000);
