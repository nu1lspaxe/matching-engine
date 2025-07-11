package http

import (
	"bytes"
	"encoding/json"
	"io"
	"matching-engine/utils/logger"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/schema"
	"go.uber.org/zap"
)

var queryEncoder = schema.NewEncoder()

var pool = sync.Pool{
	New: func() interface{} {
		return &Agent{
			header: make(map[string]string),
		}
	},
}

type Agent struct {
	client  *http.Client
	host    *url.URL
	method  string
	header  map[string]string
	body    []byte
	start   time.Time
	latency time.Duration
	debug   bool
}

func NewAgent(host string) (*Agent, error) {
	agent := pool.Get().(*Agent).reset()
	if err := agent.initHost(host); err != nil {
		return nil, err
	}
	return agent, nil
}

func (a *Agent) initHost(host string) error {
	u, err := url.Parse(host)
	if err != nil {
		return err
	}
	a.host = u
	return nil
}

func (a *Agent) reset() *Agent {
	a.client = nil
	a.host = nil
	a.method = http.MethodGet
	a.header = make(map[string]string)
	a.body = nil
	a.latency = 0
	a.debug = false
	return a
}

func (a *Agent) Debug() *Agent {
	a.debug = true
	return a
}

func (a *Agent) Use(client *http.Client) *Agent {
	a.client = client
	return a
}

func (a *Agent) Get() *Agent {
	a.method = http.MethodGet
	return a
}

func (a *Agent) Post() *Agent {
	a.method = http.MethodPost
	return a
}

func (a *Agent) Patch() *Agent {
	a.method = http.MethodPatch
	return a
}

func (a *Agent) Put() *Agent {
	a.method = http.MethodPut
	return a
}

func (a *Agent) Delete() *Agent {
	a.method = http.MethodDelete
	return a
}

func (a *Agent) Method(method string) *Agent {
	a.method = method
	return a
}

func (a *Agent) SetHeader(key, value string) *Agent {
	a.header[key] = value
	return a
}

func (a *Agent) JSON(data interface{}) *Agent {
	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal data", err.Error())
		return a
	}

	a.header["Content-Type"] = "application/json"
	a.body = bytes
	return a
}

func (a *Agent) PathJoin(paths ...string) *Agent {
	newPath := append([]string{a.host.Path}, paths...)
	a.host.Path = path.Join(newPath...)
	return a
}

func (a *Agent) Send() (code int, raw []byte, err error) {
	defer pool.Put(a)

	var req *http.Request
	if len(a.body) > 0 {
		req, err = http.NewRequest(a.method, a.host.String(), bytes.NewBuffer(a.body))
	} else {
		req, err = http.NewRequest(a.method, a.host.String(), nil)
	}

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	for k, v := range a.header {
		req.Header.Set(k, v)
	}

	if a.client == nil {
		a.client = &client
	}

	a.start = time.Now()
	resp, err := a.client.Do(req)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	defer resp.Body.Close()

	a.latency = time.Since(a.start)

	code = resp.StatusCode
	raw, err = io.ReadAll(resp.Body)

	if a.debug {
		logger.DebugWith("HTTP Request",
			zap.String("method", a.method),
			zap.String("url", a.host.String()),
			zap.Any("header", a.header),
			zap.Any("body", a.body),
			zap.Int("status", code),
			zap.Duration("latency", a.latency),
			zap.String("response", string(raw)),
			zap.Error(err),
		)
	}

	return code, raw, err
}

type QueryFunc func(query url.Values) bool

func (a *Agent) Query(queries ...QueryFunc) *Agent {
	q := a.host.Query()
	for _, f := range queries {
		if !f(q) {
			return a
		}
	}
	a.host.RawQuery = q.Encode()
	return a
}

func QueryKeyValue(key, value string) QueryFunc {
	return func(query url.Values) bool {
		query.Add(key, value)
		return true
	}
}

func QueryValues(values url.Values) QueryFunc {
	return func(query url.Values) bool {
		for k, v := range values {
			query[k] = v
		}
		return true
	}
}

func QueryStruct(data interface{}) QueryFunc {
	query, err := queryMarshaler(data)
	if err != nil {
		logger.Error("Failed to marshal data", err.Error())
		return nil
	}
	return QueryValues(query)
}

func queryMarshaler(data interface{}) (url.Values, error) {
	query := make(url.Values)
	if err := queryEncoder.Encode(data, query); err != nil {
		return nil, err
	}

	for k, v := range query {
		if len(v) > 0 {
			query[k] = []string{strings.Join(v, ",")}
		}
	}
	return query, nil
}
