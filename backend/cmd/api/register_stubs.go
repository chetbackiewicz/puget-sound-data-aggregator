package main

import (
	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/coastwatch"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/coops"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/dart"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/datawagov"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/epawqp"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/gbif"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/inaturalist"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/nanoos"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/ndbc"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/nws"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/obis"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/openmeteo"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/sunrisesunset"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/usgs"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/usno"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/wdfwgis"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/weatherapi"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/wikidata"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/wikipedia"
	"github.com/chetbackiewicz/water-data/backend/internal/sources/worms"
)

// These stubs are replaced by each fetcher package's real Register function.
// Until a fetcher is implemented, its stub is a no-op so the build stays green.

func registerNWS(reg *sources.Registry, cfg *config.Config)    { nws.Register(reg, cfg) }
func registerNDBC(reg *sources.Registry, cfg *config.Config)   { ndbc.Register(reg, cfg) }
func registerCOOPS(reg *sources.Registry, cfg *config.Config)  { coops.Register(reg, cfg) }
func registerNANOOS(reg *sources.Registry, cfg *config.Config) { nanoos.Register(reg, cfg) }
func registerUSGS(reg *sources.Registry, cfg *config.Config)   { usgs.Register(reg, cfg) }
func registerUSNO(reg *sources.Registry, cfg *config.Config)   { usno.Register(reg, cfg) }
func registerSunriseSunset(reg *sources.Registry, cfg *config.Config) {
	sunrisesunset.Register(reg, cfg)
}
func registerOpenMeteo(reg *sources.Registry, cfg *config.Config)   { openmeteo.Register(reg, cfg) }
func registerWeatherAPI(reg *sources.Registry, cfg *config.Config)  { weatherapi.Register(reg, cfg) }
func registerCoastWatch(reg *sources.Registry, cfg *config.Config)  { coastwatch.Register(reg, cfg) }
func registerWDFWGIS(reg *sources.Registry, cfg *config.Config)     { wdfwgis.Register(reg, cfg) }
func registerDataWAGov(reg *sources.Registry, cfg *config.Config)   { datawagov.Register(reg, cfg) }
func registerWoRMS(reg *sources.Registry, cfg *config.Config)       { worms.Register(reg, cfg) }
func registerINaturalist(reg *sources.Registry, cfg *config.Config) { inaturalist.Register(reg, cfg) }
func registerGBIF(reg *sources.Registry, cfg *config.Config)        { gbif.Register(reg, cfg) }
func registerOBIS(reg *sources.Registry, cfg *config.Config)        { obis.Register(reg, cfg) }
func registerWikipedia(reg *sources.Registry, cfg *config.Config)   { wikipedia.Register(reg, cfg) }
func registerWikidata(reg *sources.Registry, cfg *config.Config)    { wikidata.Register(reg, cfg) }
func registerEPAWQP(reg *sources.Registry, cfg *config.Config)      { epawqp.Register(reg, cfg) }
func registerDART(reg *sources.Registry, cfg *config.Config)        { dart.Register(reg, cfg) }
