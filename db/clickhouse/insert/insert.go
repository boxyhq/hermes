package insert

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	"go.temporal.io/sdk/workflow"
	"gopkg.in/alecthomas/kingpin.v2"
)

type configuration struct {
	backend       string
	apiBackend    string
	Addr          string
	Database      string
	Table         string
	Username      string
	Auth_database string
	Password      string
}

func parseArgs(args []string) configuration {
	var cfg configuration

	app := kingpin.New(filepath.Base(args[0]), "hermes")

	app.Flag("backend", "backend to use to store data").
		Envar("BACKEND").Default("clickhouse").StringVar(&cfg.backend)

	app.Flag("clickhouse-endpoint", "endpoint to query Loki logs").
		Envar("CLICKHOUSE_ENDPOINT").Default("http://localhost:9000").StringVar(&cfg.Addr)
	app.Flag("clickhouse-database", "endpoint to query Loki logs").
		Envar("CLICKHOUSE_DATABASE").Default("hermes").StringVar(&cfg.Database)
	app.Flag("clickhouse-table", "endpoint to query Loki logs").
		Envar("CLICKHOUSE_TABLE").Default("auditlogs").StringVar(&cfg.Table)
	app.Flag("clickhouse-username", "endpoint to query Loki logs").
		Envar("CLICKHOUSE_USERNAME").Default("default").StringVar(&cfg.Username)
	app.Flag("clickhouse-auth-database", "endpoint to query Loki logs").
		Envar("CLICKHOUSE_AUTH_DATABASE").Default("default").StringVar(&cfg.Auth_database)

	app.Flag("api-backend", "backend to use to validate API keys").
		Envar("API_BACKEND").Default("demo").StringVar(&cfg.apiBackend)

	kingpin.MustParse(app.Parse(args[1:]))
	return cfg
}

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Insert Log workflow started")

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("Insert Log workflow completed.")

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	cfg := parseArgs(os.Args)
	_config := &clickhousedb.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhousedb.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		// Debug:           true,
		DialTimeout:     10 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Minute * 5,
	}
	session, _ := clickhousedb.Open(_config)

	if errInsert := session.AsyncInsert(context.TODO(), name, false); errInsert != nil {
		fmt.Print("Failed to save document", map[string]interface{}{
			"query": name,
			"error": errInsert,
		})
		return "", errInsert
	}
	return "", nil
}
