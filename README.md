# Puget Sound Fishing Data (v0)

A web application for spearfishing and kayak fishing in Puget Sound. v0 aggregates public marine, weather, and fisheries data sources into a single UI, with admin pages for manually entering WDFW regulations and fishing technique content.

- **Backend:** Go (chi router, `database/sql`, MySQL driver)
- **Frontend:** TypeScript + React + Vite + Mantine + TanStack Query
- **Database:** MySQL 8 (spatial + JSON columns)
- **Local dev:** Docker Compose for MySQL; backend + frontend run on host

## Quick start

Prerequisites:

- Go 1.23+
- Node.js 18+
- Docker Desktop (for the MySQL service) **or** a local MySQL 8 instance (`brew install mysql`)

```bash
# 1. Configure env
cp .env.example .env
# Edit .env — at minimum set NWS_USER_AGENT to identify your app

# 2. Start MySQL
make up

# 3. Run migrations + seed data
make migrate

# 4. In one terminal: run the Go API on :8080
make api

# 5. In another terminal: run the React app on :5173
cd frontend && npm install
make web
```

Open <http://localhost:5173>. The **Probes** page lists every upstream source — click *Run probe* on a card to make a live HTTP call and inspect the parsed + raw response.

## Manual setup checklist

These are one-time tasks. The app degrades gracefully when optional keys are missing.

| # | Item | Required? | How |
|---|---|---|---|
| 1 | `NWS_USER_AGENT` env var | **Yes** | NWS API requires a UA header identifying your app, e.g. `(myapp.com, contact@me.com)`. |
| 2 | `USNO_APP_ID` env var (≤ 8 chars) | Recommended | Polite identifier on US Naval Observatory calls. |
| 3 | MySQL credentials | **Yes** | Defaults match `docker-compose.yml`. Override in `.env` if using a host MySQL. |
| 4 | `WEATHERAPI_KEY` | Optional | Free at <https://www.weatherapi.com/signup.aspx>. Without it the WeatherAPI Marine probe returns a clear "missing key" error. |
| 5 | `OWM_API_KEY` | Optional | Free at <https://openweathermap.org/api>. Not wired into a v0 probe; reserved for v0.1. |
| 6 | `AISSTREAM_API_KEY` | Deferred | <https://www.aisstream.io/apikeys>. AIS vessel feed (WebSocket) is a v0.1 feature. |
| 7 | WDFW Sport Fishing Rules PDF | **Yes (content)** | Download the current pamphlet (e.g. `25WAFW_LR9.pdf`) from <https://wdfw.wa.gov/fishing/regulations>. Use it as the source-of-truth document to type entries into the **Regulations** page. Record the page reference in the *source_doc* field. |
| 8 | Bait/tackle/technique content | Optional | Enter via the **Techniques** page. Use the Markdown editor for long-form how-tos. Attribute external sources. |
| 9 | PSMFC RMIS bulk download | Optional | Email `nleonard@psmfc.org` for a data-use agreement. Out of scope for v0. |
| 10 | NASA Earthdata (MUR SST) | Optional | Register at <https://urs.earthdata.nasa.gov/>. Not wired into v0. |

## Data sources probed in v0

See `plan.md` for the full list (≈ 30 source keys across NWS, NDBC, NOAA CO-OPS, NANOOS ORCA, USGS NWIS, USNO, Sunrise-Sunset, Open-Meteo, WeatherAPI, NOAA CoastWatch SST, WDFW ArcGIS, data.wa.gov, WoRMS, iNaturalist, GBIF, OBIS, Wikipedia, Wikidata, EPA Water Quality Portal, Columbia River DART).

All sources are public except where flagged `auth_required: true` in `GET /api/sources`. v0 makes **one** smoke-test call per source on demand; nothing runs on a schedule.

## Project layout

```
backend/
  cmd/api/                 # main, subcommands (serve, migrate up)
  internal/
    config/                # env loading
    db/                    # MySQL pool + golang-migrate wrapper
    httpx/                 # chi router + middleware
    sources/               # one package per upstream API
    handlers/              # HTTP handlers (probes, species, regulations, techniques, marine-areas)
    store/                 # data access for MySQL tables
  migrations/              # golang-migrate SQL files
frontend/
  src/
    api/                   # axios client + TS types
    pages/                 # Dashboard, Probes, Species, Regulations, Techniques
docker-compose.yml         # MySQL 8 service
Makefile                   # operational entry points
.env.example               # configuration template
plan.md                    # implementation plan and schema
```

## REST API surface

```
GET    /api/health
GET    /api/sources                    list all probe-able sources + last result
POST   /api/probes/{key}               run a single probe (persists result)
GET    /api/probes/{key}/latest        last cached probe result

GET    /api/marine-areas               list WA Marine Areas (GeoJSON)
GET    /api/marine-areas/{id}

GET    /api/species
POST   /api/species
GET    /api/species/{id}
PUT    /api/species/{id}

GET    /api/regulations?marine_area_id=&species_id=
POST   /api/regulations
PUT    /api/regulations/{id}
DELETE /api/regulations/{id}

GET    /api/techniques?species_id=&method=
POST   /api/techniques
PUT    /api/techniques/{id}
DELETE /api/techniques/{id}
```

**No auth in v0** — CRUD endpoints are open and assume localhost use. Add JWT/session auth before deploying remotely.

## Data attribution & licensing

When displaying data in any public-facing build, observe these license terms:

- **NOAA NWS, NDBC, CO-OPS, CoastWatch, USGS NWIS, USNO, EPA, WDFW** — U.S. government public domain. Attribution is good practice but not required.
- **WoRMS** — CC BY 4.0. Cite "WoRMS Editorial Board (year). World Register of Marine Species. <https://www.marinespecies.org>".
- **GBIF, OBIS** — open with attribution; cite dataset DOIs returned per record.
- **iNaturalist** — observation metadata is CC0; photo licenses vary. **Filter to CC0/CC BY before commercial display.**
- **Wikipedia text** — CC BY-SA 4.0. ShareAlike applies — attribute and link source. Wikidata is CC0.
- **Open-Meteo, WeatherAPI** — see provider terms; free tiers allow non-commercial use.
- **NANOOS / ORCA** — academic + public good; cite providers.
- **FishBase** — **non-commercial** license; **not** ingested in v0. Do not add without negotiating a license.

## Development tasks

```bash
# Backend
cd backend && go build ./...        # compile everything
cd backend && go vet ./...          # static checks
cd backend && go test ./...         # tests (when added)

# Frontend
cd frontend && npm run typecheck    # tsc --noEmit
cd frontend && npm run build        # production build
```

## Known v0 limitations & roadmap

- **No auth.** v0.1: add JWT or session auth on write endpoints.
- **No background scheduler.** Probes are user-triggered. v0.1: cron worker for periodic refresh.
- **No spatial library on Go side.** Uses raw MySQL `ST_GeomFromGeoJSON` / `ST_AsGeoJSON`.
- **AISStream deferred.** Needs a WebSocket consumer + backend proxy (no browser CORS).
- **Bathymetry / charts deferred.** Static GEBCO download is a build-time concern.
- **HAB / biotoxin closures deferred.** No documented WA DOH API yet.
- **Open-Meteo subdomain ambiguity.** Fetcher tries `marine-api.open-meteo.com` first, falls back to `api.open-meteo.com` on 404.

## License

The application code is licensed under the MIT License. Data shown in the UI is subject to upstream providers' terms — see *Data attribution* above.
