package clickhouse

type Config struct {
	Addr          string
	Table         string
	Username      string
	Password      string
	Auth_database string
	Database      string
}

type stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type queryResponse struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string   `json:"resultType"`
	Result     []stream `json:"result"`
}
