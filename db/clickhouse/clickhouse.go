package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/boxyhq/hermes/db"
	"github.com/boxyhq/hermes/types"

	"github.com/boxyhq/hermes/db/clickhouse/insert"
	"go.temporal.io/sdk/client"
)

var ErrNoEvents = errors.New("no events")

type clickhouse struct {
	cfg       Config
	ingestURL string
	queryURL  string
	client    clickhousedb.Conn
}

func constructQuery(indexes map[string]string) string {
	var qb strings.Builder
	qb.WriteString("{")
	i := 0
	for k, v := range indexes {
		if i > 0 {
			qb.WriteString(",")
		}
		qb.WriteString(k)
		qb.WriteString("=")
		qb.WriteString(`"`)
		qb.WriteString(v)
		qb.WriteString(`"`)
		i++
	}
	qb.WriteString("}")

	return qb.String()
}

func (l clickhouse) Ingest(tenantID string, logs []types.AuditLog) error {
	if len(logs) <= 0 {
		return ErrNoEvents
	}
	log.Println("Got ", len(logs), "rows")
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "insert_clickhouse_workflow",
		TaskQueue: "insert-clickhouse",
	}

	for _, al := range logs {
		fmt.Print(al)
		now := time.Now().UnixNano()
		query := fmt.Sprintf(`INSERT INTO %s.%s (tenantId, timestamp, actor, actor_type, group, where, where_type, when, target, target_id, action, action_type, name, description) VALUES (%d, %d, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')`, l.cfg.Database, l.cfg.Table, 1, now, al.Where, al.WhereType, al.When, al.Target, al.TargetID, al.Action, al.ActionType, al.Name, al.Description, al.Actor, al.ActorType, al.Group)

		we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, insert.Workflow, query)
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}

		log.Println("Started workflow")
		log.Println("WorkflowID", we.GetID())
		log.Println("RunID", we.GetRunID())

		// To Inset in clickhouse

		// if errInsert := l.client.AsyncInsert(context.TODO(), query, false); errInsert != nil {
		// 	fmt.Print("Failed to save document", map[string]interface{}{
		// 		"query": query,
		// 		"error": errInsert,
		// 	})

		// 	return errInsert
		// }
	}

	return nil
}

func (l clickhouse) Query(tenantID string, indexes map[string]string, start, end int64) ([]map[string]interface{}, error) {
	if indexes == nil {
		indexes = map[string]string{}
	}
	indexes["tenantID"] = tenantID
	q := constructQuery(indexes)

	params := url.Values{}
	params.Add("query", q)
	if start > 0 {
		params.Add("start", strconv.Itoa(int(start)))
	}
	if end > 0 {
		params.Add("end", strconv.Itoa(int(end)))
	}

	enc := params.Encode()

	fmt.Print(enc)

	// rsp, err := l.client.Get(l.queryURL + enc)
	// if err != nil {
	// 	return nil, err
	// }

	// defer rsp.Body.Close()

	// b, err := ioutil.ReadAll(rsp.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// var qrsp queryResponse
	// json.Unmarshal(b, &qrsp)

	// if qrsp.Status != "success" || qrsp.Data.ResultType != "streams" {
	// 	return nil, fmt.Errorf("error querying loki: status = %s, resultType = %s", qrsp.Status,
	// 		qrsp.Data.ResultType)
	// }

	// if len(qrsp.Data.Result) <= 0 {
	// 	return []map[string]interface{}{}, nil
	// }

	logs := make([]map[string]interface{}, 0, 1)
	// for _, s := range qrsp.Data.Result {
	// 	for _, v := range s.Values {
	// 		var l map[string]interface{}
	// 		err := json.Unmarshal([]byte(v[1]), &l)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		for k, v := range s.Stream {
	// 			l[k] = v
	// 		}

	// 		logs = append(logs, l)
	// 	}
	// }

	return logs, nil
}

func (l clickhouse) Close() {
	l.client.Close()
}

func New(cfg Config) (db.DB, error) {
	_config := &clickhousedb.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhousedb.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		// Debug:           true,
		DialTimeout:     10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: time.Minute * 60,
	}

	session, _ := clickhousedb.Open(_config)

	return clickhouse{
		cfg:       cfg,
		ingestURL: "",
		queryURL:  "",
		client:    session,
	}, nil
}
