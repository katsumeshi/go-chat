package trace

import (
	"fmt"
	"io"
)

type tracer struct {
	out io.Writer
}

type nilTracer struct{}

type Tracer interface {
	Trace(...interface{})
}

func (t *nilTracer) Trace(...interface{}) {}

func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprint(t.out, a...)
	fmt.Fprintln(t.out)
}

func Off() Tracer {
	return &nilTracer{}
}

func New(w io.Writer) Tracer {
	return &tracer{out: w}
}
