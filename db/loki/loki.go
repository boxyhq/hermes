package loki

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/boxyhq/hermes/db"
	"github.com/boxyhq/hermes/types"
)

const (
	ingestPath = "loki/api/v1/push"
	queryPath  = "loki/api/v1/query_range?"

	tsKey = "_ts"
)

var ErrNoEvents = errors.New("no events")

type loki struct {
	cfg       Config
	ingestURL string
	queryURL  string
	client    *http.Client
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

func (l loki) Ingest(tenantID string, logs []types.AuditLog) error {
	if len(logs) <= 0 {
		return ErrNoEvents
	}

	m := map[string][]stream{}
	a := make([]stream, 0, len(logs))

	for _, al := range logs {
		indexes, rest := al.Indexes()
		indexes["tenantID"] = tenantID
		indexes["when"] = al.When

		log, err := json.Marshal(rest)
		if err != nil {
			return err
		}

		ts, err := time.Parse(time.RFC3339Nano, al.When)
		if err != nil {
			return err
		}

		a = append(a, stream{
			Stream: indexes,
			Values: [][]string{
				{
					strconv.FormatInt(ts.UnixNano(), 10), // timestamp
					string(log),                          // log
				},
			},
		})
	}

	m["streams"] = a

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	body := bytes.NewReader(b)

	rsp, err := l.client.Post(l.ingestURL, "application/json", body)
	if err != nil {
		return err
	}

	r, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()
	if rsp.StatusCode >= 200 && rsp.StatusCode < 300 {
		return nil
	}

	return errors.New(string(r))
}

func (l loki) Query(tenantID string, indexes map[string]string, start, end int64) ([]map[string]interface{}, error) {
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

	rsp, err := l.client.Get(l.queryURL + enc)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	var qrsp queryResponse
	json.Unmarshal(b, &qrsp)

	if qrsp.Status != "success" || qrsp.Data.ResultType != "streams" {
		return nil, fmt.Errorf("error querying loki: status = %s, resultType = %s", qrsp.Status,
			qrsp.Data.ResultType)
	}

	if len(qrsp.Data.Result) <= 0 {
		return []map[string]interface{}{}, nil
	}

	logs := make([]map[string]interface{}, 0, len(qrsp.Data.Result))
	for _, s := range qrsp.Data.Result {
		for _, v := range s.Values {
			var l map[string]interface{}
			err := json.Unmarshal([]byte(v[1]), &l)
			if err != nil {
				return nil, err
			}

			ts, _ := strconv.ParseInt(v[0], 10, 0)
			l[tsKey] = ts

			for k, v := range s.Stream {
				l[k] = v
			}

			logs = append(logs, l)
		}
	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i][tsKey].(int64) > logs[j][tsKey].(int64)
	})

	for _, l := range logs {
		delete(l, tsKey)
	}

	return logs, nil
}

func (l loki) Close() {
	l.client.CloseIdleConnections()
}

func New(cfg Config) (db.DB, error) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	return loki{
		cfg:       cfg,
		ingestURL: cfg.IngestEndpoint + "/" + ingestPath,
		queryURL:  cfg.QueryEndpoint + "/" + queryPath,
		client:    httpClient,
	}, nil
}
