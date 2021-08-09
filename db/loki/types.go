package loki

type Config struct {
	IngestEndpoint string
	QueryEndpoint  string
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
