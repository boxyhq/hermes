package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/boxyhq/hermes/auth"
	"github.com/boxyhq/hermes/db"
	"github.com/boxyhq/hermes/db/loki"
	"github.com/boxyhq/hermes/types"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var hdb db.DB

type configuration struct {
	backend    string
	apiBackend string
	loki       loki.Config
}

func parseArgs(args []string) configuration {
	var cfg configuration

	app := kingpin.New(filepath.Base(args[0]), "hermes")

	app.Flag("backend", "backend to use to store data").
		Envar("BACKEND").Default("loki").StringVar(&cfg.backend)
	app.Flag("loki-ingest-endpoint", "endpoint to ingest Loki logs").
		Envar("LOKI_INGEST_ENDPOINT").Default("http://localhost:3100").StringVar(&cfg.loki.IngestEndpoint)
	app.Flag("loki-query-endpoint", "endpoint to query Loki logs").
		Envar("LOKI_QUERY_ENDPOINT").Default("http://localhost:3100").StringVar(&cfg.loki.QueryEndpoint)

	app.Flag("api-backend", "backend to use to validate API keys").
		Envar("API_BACKEND").Default("demo").StringVar(&cfg.apiBackend)

	kingpin.MustParse(app.Parse(args[1:]))

	return cfg
}

func main() {
	cfg := parseArgs(os.Args)

	zapCfg := zap.NewProductionConfig()
	zapCfg.DisableStacktrace = true
	log, err := zapCfg.Build()
	if err != nil {

	}

	defer log.Sync()

	switch cfg.backend {
	case "loki":
		hdb, err = loki.New(cfg.loki)
		if err != nil {
			log.Error("Error initialising db:", zap.Error(err))
		}

		break
	default:
		log.Fatal("unknown backend, exiting", zap.String("backend", cfg.backend))

		break
	}

	defer hdb.Close()

	router := http.NewServeMux()
	router.HandleFunc("/ingest", ingestHandler)
	router.HandleFunc("/query", queryHandler)

	authMiddleware := auth.Authorize(auth.NewDemoValidator())
	authRouter := authMiddleware(router)

	log.Info("Hermes: Requisition me a beat!", zap.Int("port", 8080))

	err = http.ListenAndServe(":8080", authRouter)
	if err != nil {
		log.Fatal("Error serving:", zap.Error(err))
	}
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var logs []types.AuditLog
	err = json.Unmarshal(b, &logs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authMeta := auth.MetadataFromCtx(r.Context())
	if authMeta.Empty() || !authMeta.ValidateScope(auth.ScopeWriteEvents) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = hdb.Ingest(authMeta.TenantID, logs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	authMeta := auth.MetadataFromCtx(r.Context())
	if authMeta.Empty() || !authMeta.ValidateScope(auth.ScopeReadEvents) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logs, err := hdb.Query(authMeta.TenantID, map[string]string{}, 0, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(logs)
	w.Write(b)
}
