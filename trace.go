package gin_trace

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
)

type traceReader struct {
	io.ReadCloser
	body *bytes.Buffer
}

func (t *traceReader) Read(p []byte) (n int, err error) {
	n, err = t.ReadCloser.Read(p)
	t.body.Write(p)
	return
}

type traceWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *traceWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *traceWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

type Logger interface {
	Trace(src ...interface{})
}

func NewTrace(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reader := &traceReader{ReadCloser: c.Request.Body, body: bytes.NewBufferString("")}
		c.Request.Body = reader
		writer := &traceWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = writer
		c.Next()
		if logger != nil {
			logger.Trace(reader.body.String(), writer.body.String())
		} else {
			fmt.Println(reader.body.String(), writer.body.String())
		}
	}
}
