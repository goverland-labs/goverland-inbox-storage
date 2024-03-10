package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type RequestWatcher struct {
	name string
}

func NewRequestWatcher(name string) *RequestWatcher {
	return &RequestWatcher{
		name: name,
	}
}

func (m *RequestWatcher) RoundTrip(r *http.Request) (*http.Response, error) {
	var err error
	defer func(start time.Time) {
		CollectRequestsMetric(m.name, r.Header.Get("alias"), err, start)
	}(time.Now())

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	if data, ok := resp.Header["Ratelimit-Remaining"]; ok && len(data) > 0 {
		if val, err := strconv.ParseFloat(data[0], 64); err == nil {
			CollectKeyState(m.name, "remaining_value", val)
		}
	}

	return resp, nil
}
