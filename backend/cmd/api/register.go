// Package main wires registered Source implementations.
// Each fetcher package contributes a register function.
package main

import (
	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

// registerAll is the single place each fetcher gets wired into the registry.
// New fetcher packages should add their Register call here.
func registerAll(reg *sources.Registry, cfg *config.Config) {
	// Fetchers register themselves through their own Register(reg, cfg) funcs.
	// Stubs below are filled in by individual fetcher packages.
	registerNWS(reg, cfg)
	registerNDBC(reg, cfg)
	registerCOOPS(reg, cfg)
	registerNANOOS(reg, cfg)
	registerUSGS(reg, cfg)
	registerUSNO(reg, cfg)
	registerSunriseSunset(reg, cfg)
	registerOpenMeteo(reg, cfg)
	registerWeatherAPI(reg, cfg)
	registerCoastWatch(reg, cfg)
	registerWDFWGIS(reg, cfg)
	registerDataWAGov(reg, cfg)
	registerWoRMS(reg, cfg)
	registerINaturalist(reg, cfg)
	registerGBIF(reg, cfg)
	registerOBIS(reg, cfg)
	registerWikipedia(reg, cfg)
	registerWikidata(reg, cfg)
	registerEPAWQP(reg, cfg)
	registerDART(reg, cfg)
}
