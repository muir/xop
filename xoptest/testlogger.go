// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptest

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"
)

type testingT interface {
	Log(...interface{})
	Name() string
}

var (
	_ xopbase.Logger     = &TestLogger{}
	_ xopbase.Request    = &Span{}
	_ xopbase.Span       = &Span{}
	_ xopbase.Prefilling = &Prefilling{}
	_ xopbase.Prefilled  = &Prefilled{}
	_ xopbase.Line       = &Line{}
)

func New(t testingT) *TestLogger {
	return &TestLogger{
		t:        t,
		id:       t.Name() + "-" + uuid.New().String(),
		traceMap: make(map[string]*traceInfo),
	}
}

type TestLogger struct {
	lock       sync.Mutex
	t          testingT
	Requests   []*Span
	Spans      []*Span
	Lines      []*Line
	traceCount int
	traceMap   map[string]*traceInfo
	id         string
}

type traceInfo struct {
	spanCount int
	traceNum  int
	spans     map[string]int
}

type Span struct {
	Attributes   xoputil.AttributeBuilder
	lock         sync.Mutex
	testLogger   *TestLogger
	Trace        trace.Bundle
	IsRequest    bool
	Parent       *Span
	Spans        []*Span
	RequestLines []*Line
	Lines        []*Line
	short        string
	Metadata     map[string]interface{}
}

type Prefilling struct {
	Builder
}

type Builder struct {
	Data   map[string]interface{}
	Span   *Span
	kvText []string
}

type Prefilled struct {
	Data   map[string]interface{}
	Span   *Span
	Msg    string
	kvText []string
}

type Line struct {
	Builder
	Level     xopconst.Level
	Timestamp time.Time
	Message   string
	Text      string
	Tmpl      string
}

func (l *TestLogger) WithMe() xop.SeedModifier {
	return xop.WithBaseLogger("testing", l)
}

func (l *TestLogger) ID() string                   { return l.id }
func (l *TestLogger) Close()                       {}
func (l *TestLogger) Buffered() bool               { return false }
func (l *TestLogger) ReferencesKept() bool         { return true }
func (l *TestLogger) SetErrorReporter(func(error)) {}
func (l *TestLogger) Request(span trace.Bundle, name string) xopbase.Request {
	l.lock.Lock()
	defer l.lock.Unlock()
	s := &Span{
		testLogger: l,
		IsRequest:  true,
		Trace:      span,
		short:      l.setShort(span, name),
	}
	s.Attributes.Reset()
	return s
}

func (l *TestLogger) setShort(span trace.Bundle, name string) string {
	ts := span.Trace.GetTraceID().String()
	if ti, ok := l.traceMap[ts]; ok {
		ti.spanCount++
		ti.spans[span.Trace.GetSpanID().String()] = ti.spanCount
		short := fmt.Sprintf("T%d.%d", ti.traceNum, ti.spanCount)
		l.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
		return short
	}
	l.traceCount++
	l.traceMap[ts] = &traceInfo{
		spanCount: 1,
		traceNum:  l.traceCount,
		spans: map[string]int{
			span.Trace.GetSpanID().String(): 1,
		},
	}
	short := fmt.Sprintf("T%d.%d", l.traceCount, 1)
	l.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
	return short
}

func (s *Span) Flush()                       {}
func (s *Span) Boring(bool)                  {}
func (s *Span) ID() string                   { return s.testLogger.id }
func (s *Span) SetErrorReporter(func(error)) {}

func (s *Span) Span(span trace.Bundle, name string) xopbase.Span {
	s.testLogger.lock.Lock()
	defer s.testLogger.lock.Unlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	n := &Span{
		testLogger: s.testLogger,
		Trace:      span,
		short:      s.testLogger.setShort(span, name),
	}
	n.Attributes.Reset()
	s.Spans = append(s.Spans, n)
	s.testLogger.Spans = append(s.testLogger.Spans, n)
	return n
}

func (s *Span) NoPrefill() xopbase.Prefilled {
	return &Prefilled{
		Span: s,
	}
}

func (s *Span) StartPrefill() xopbase.Prefilling {
	return &Prefilling{
		Builder: Builder{
			Data: make(map[string]interface{}),
			Span: s,
		},
	}
}

func (p *Prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return &Prefilled{
		Data:   p.Data,
		Span:   p.Span,
		kvText: p.kvText,
		Msg:    m,
	}
}

func (p *Prefilled) Line(level xopconst.Level, t time.Time, _ []uintptr) xopbase.Line {
	// TODO: stack traces
	line := &Line{
		Builder: Builder{
			Data: make(map[string]interface{}),
			Span: p.Span,
		},
		Level:     level,
		Timestamp: t,
	}
	for k, v := range p.Data {
		line.Data[k] = v
	}
	if len(p.kvText) != 0 {
		line.kvText = make([]string, len(p.kvText), len(p.kvText)+5)
		copy(line.kvText, p.kvText)
	}
	line.Tmpl = p.Msg
	line.Message = p.Msg
	return line
}

func (l *Line) Static(m string) {
	l.Msg(m)
}

func (l *Line) Msg(m string) {
	l.Message += m
	text := l.Span.short + ": " + l.Message
	if len(l.kvText) > 0 {
		text += " " + strings.Join(l.kvText, " ")
		l.kvText = nil
	}
	l.Text = text
	l.send(text)
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

func (l *Line) Template(m string) {
	l.Tmpl += m
	used := make(map[string]struct{})
	text := l.Span.short + ": " +
		templateRE.ReplaceAllStringFunc(l.Tmpl, func(k string) string {
			k = k[1 : len(k)-1]
			if v, ok := l.Data[k]; ok {
				used[k] = struct{}{}
				return fmt.Sprint(v)
			}
			return "''"
		})
	for k, v := range l.Data {
		if _, ok := used[k]; !ok {
			text += " " + k + "=" + fmt.Sprint(v)
		}
	}
	l.Text = text
	l.send(text)
}

func (l Line) send(text string) {
	l.Span.testLogger.t.Log(text)
	l.Span.testLogger.lock.Lock()
	defer l.Span.testLogger.lock.Unlock()
	l.Span.lock.Lock()
	defer l.Span.lock.Unlock()
	l.Span.testLogger.Lines = append(l.Span.testLogger.Lines, &l)
	l.Span.Lines = append(l.Span.Lines, &l)
}

func (b *Builder) Any(k string, v interface{}) {
	b.Data[k] = v
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%+v", k, v))
}

func (b *Builder) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	b.Data[k.Key()] = v.String()
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%s(%d)", k.Key(), v.String(), v.Int64()))
}

func (b *Builder) Bool(k string, v bool)              { b.Any(k, v) }
func (b *Builder) Duration(k string, v time.Duration) { b.Any(k, v) }
func (b *Builder) Error(k string, v error)            { b.Any(k, v) }
func (b *Builder) Int(k string, v int64)              { b.Any(k, v) }
func (b *Builder) Link(k string, v trace.Trace)       { b.Any(k, v) }
func (b *Builder) Str(k string, v string)             { b.Any(k, v) }
func (b *Builder) Time(k string, v time.Time)         { b.Any(k, v) }
func (b *Builder) Uint(k string, v uint64)            { b.Any(k, v) }

func (s *Span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) { s.Attributes.MetadataAny(k, v) }
func (s *Span) MetadataBool(k *xopconst.BoolAttribute, v bool)      { s.Attributes.MetadataBool(k, v) }
func (s *Span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	s.Attributes.MetadataEnum(k, v)
}
func (s *Span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.Attributes.MetadataInt64(k, v) }
func (s *Span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.Attributes.MetadataLink(k, v)
}
func (s *Span) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	s.Attributes.MetadataNumber(k, v)
}
func (s *Span) MetadataStr(k *xopconst.StrAttribute, v string)      { s.Attributes.MetadataStr(k, v) }
func (s *Span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.Attributes.MetadataTime(k, v) }

// end