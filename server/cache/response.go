package cache

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ResponseContent struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`
}

func (c ResponseContent) MarshalJSON() ([]byte, error) {
	type marshalingAlias ResponseContent
	return json.Marshal(marshalingAlias(c))
}

func (c *ResponseContent) UnmarshalJSON(data []byte) error {
	type marshalingAlias ResponseContent
	return json.Unmarshal(data, (*marshalingAlias)(c))
}

func (c ResponseContent) copyTo(w http.ResponseWriter) error {
	for key, values := range c.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(c.StatusCode)
	_, err := w.Write(c.Body)
	return err
}

type bufferedResponseWriter struct {
	statusCode int
	header     http.Header
	bodyBuffer *bytes.Buffer
}

func newBufferedResponseWriter(w http.ResponseWriter) *bufferedResponseWriter {
	header := w.Header().Clone()
	if header == nil {
		header = make(http.Header)
	}
	return &bufferedResponseWriter{
		header:     header,
		bodyBuffer: &bytes.Buffer{},
	}
}

func (bw *bufferedResponseWriter) Header() http.Header  { return bw.header }
func (bw *bufferedResponseWriter) WriteHeader(code int) { bw.statusCode = code }

func (bw *bufferedResponseWriter) Write(b []byte) (int, error) {
	if _, err := bw.bodyBuffer.Write(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (bw *bufferedResponseWriter) toResponseContent() *ResponseContent {
	return &ResponseContent{
		StatusCode: bw.statusCode,
		Headers:    bw.header.Clone(),
		Body:       bw.bodyBuffer.Bytes(),
	}
}
