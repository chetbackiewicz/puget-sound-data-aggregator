package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/db"
	"github.com/chetbackiewicz/water-data/backend/internal/handlers"
	"github.com/chetbackiewicz/water-data/backend/internal/httpx"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
	"github.com/chetbackiewicz/water-data/backend/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		if err := serve(); err != nil {
			slog.Error("serve failed", "err", err)
			os.Exit(1)
		}
	case "migrate":
		if err := migrateCmd(os.Args[2:]); err != nil {
			slog.Error("migrate failed", "err", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: api [serve|migrate up]")
}

func serve() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	sqlDB, err := db.Open(cfg.DSN())
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if err := db.Ping(sqlDB); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	reg := sources.NewRegistry()
	registerAll(reg, cfg)

	probeStore := &store.ProbeStore{DB: sqlDB}
	sourcesH := handlers.NewSourcesHandler(reg, probeStore)

	speciesH := &handlers.SpeciesHandler{Store: &store.SpeciesStore{DB: sqlDB}}
	regsH := &handlers.RegulationsHandler{Store: &store.RegulationStore{DB: sqlDB}}
	techH := &handlers.TechniquesHandler{Store: &store.TechniqueStore{DB: sqlDB}}
	maH := &handlers.MarineAreasHandler{Store: &store.MarineAreaStore{DB: sqlDB}}

	r := httpx.NewRouter(cfg.CORSOrigins)
	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, 200, map[string]string{"status": "ok"})
	})
	r.Get("/api/sources", sourcesH.List)
	r.Post("/api/probes/{key}", sourcesH.Run)
	r.Get("/api/probes/{key}/latest", sourcesH.Latest)

	r.Get("/api/marine-areas", maH.List)
	r.Get("/api/marine-areas/{id}", maH.Get)

	r.Get("/api/species", speciesH.List)
	r.Post("/api/species", speciesH.Create)
	r.Get("/api/species/{id}", speciesH.Get)
	r.Put("/api/species/{id}", speciesH.Update)

	r.Get("/api/regulations", regsH.List)
	r.Post("/api/regulations", regsH.Create)
	r.Put("/api/regulations/{id}", regsH.Update)
	r.Delete("/api/regulations/{id}", regsH.Delete)

	r.Get("/api/techniques", techH.List)
	r.Post("/api/techniques", techH.Create)
	r.Put("/api/techniques/{id}", techH.Update)
	r.Delete("/api/techniques/{id}", techH.Delete)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr, "sources", len(reg.All()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func migrateCmd(args []string) error {
	if len(args) < 1 || args[0] != "up" {
		return fmt.Errorf("usage: api migrate up")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	sqlDB, err := db.Open(cfg.DSN())
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if err := db.Ping(sqlDB); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	cwd, _ := os.Getwd()
	dir, err := findMigrationsDir(cwd)
	if err != nil {
		return err
	}
	slog.Info("running migrations", "dir", dir)
	return db.MigrateUp(sqlDB, dir)
}

// findMigrationsDir walks up from cwd looking for a backend/migrations directory.
func findMigrationsDir(start string) (string, error) {
	dir := start
	for i := 0; i < 5; i++ {
		candidate := filepath.Join(dir, "backend", "migrations")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}
		candidate2 := filepath.Join(dir, "migrations")
		if st, err := os.Stat(candidate2); err == nil && st.IsDir() {
			return candidate2, nil
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("could not locate migrations dir from %s", start)
}
