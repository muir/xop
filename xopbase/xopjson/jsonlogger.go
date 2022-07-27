// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
	"github.com/muir/xoplog/xoputil"

	"github.com/google/uuid"
	"github.com/phuslu/fasttime"
)

const (
	maxBufferToKeep = 1024 * 10
	minBuffer       = 1024
)

var (
	_ xopbase.Logger  = &Logger{}
	_ xopbase.Request = &Span{}
	_ xopbase.Span    = &Span{}
	_ xopbase.Line    = &Line{}
)

type Option func(*Logger)

type timeOption int

const (
	epochTime timeOption = iota
	strftimeTime
	timeTime
	unixNano
)

type DurationOption int

const (
	AsNanos   DurationOption = iota // int64(duration)
	AsMillis                        // int64(duration / time.Milliscond)
	AsSeconds                       // int64(duration / time.Second)
	AsString                        // duration.String()
)

type AsynchronousWriter interface {
	Write([]byte) (int, error)
	Flush() error
	Close() error
	Buffered() bool
}

type Logger struct {
	writer           AsynchronousWriter
	timeOption       timeOption
	timeFormat       string
	framesAtLevelMap map[xopconst.Level]int
	framesAtLevel    [xopconst.AlertLevel]int
	withGoroutine    bool
	fastKeys         bool
	durationFormat   DurationOption
	id               uuid.UUID
}

type Request struct {
	*Span
}

type prefill struct {
	data []byte
	msg  string
}

type Span struct {
	attributes xoputil.AttributeBuilder
	trace      trace.Bundle
	logger     *Logger
	prefill    atomic.Value
	errorFunc  func(error)
}

type Line struct {
	dataBuffer xoputil.JBuilder
	level      xopconst.Level
	timestamp  time.Time
	span       *Span
	prefillLen int
	encoder    *json.Encoder
}

func WithUncheckedKeys(b bool) Option {
	return func(l *Logger) {
		l.fastKeys = b
	}
}

// TODO: allow custom error formats

// WithStrftime adds a timestamp to each log line.  See
// https://github.com/phuslu/fasttime for the supported
// formats.
func WithStrftime(format string) Option {
	return func(l *Logger) {
		l.timeOption = strftimeTime
		l.timeFormat = format
	}
}

func WithDuration(durationFormat DurationOption) Option {
	return func(l *Logger) {
		l.durationFormat = durationFormat
	}
}

func WithCallersAtLevel(logLevel xopconst.Level, framesWanted int) Option {
	return func(l *Logger) {
		l.framesAtLevelMap[logLevel] = framesWanted
		l.framesAtLevel[logLevel] = framesWanted
	}
}

func WithGoroutineID(b bool) Option {
	return func(l *Logger) {
		l.withGoroutine = b
	}
}

func New(w AsynchronousWriter, opts ...Option) *Logger {
	logger := &Logger{
		writer:           w,
		framesAtLevelMap: make(map[xopconst.Level]int),
		id:               uuid.New(),
	}
	for _, f := range opts {
		f(logger)
	}
	return logger
}

func (l *Logger) ID() string                                { return l.id.String() }
func (l *Logger) Buffered() bool                            { return l.writer.Buffered() }
func (l *Logger) ReferencesKept() bool                      { return false }
func (l *Logger) StackFramesWanted() map[xopconst.Level]int { return l.framesAtLevelMap }

func (l *Logger) Close() {
	// no place to report errors
	_ = l.writer.Close()
}

func (l *Logger) Request(span trace.Bundle, name string) xopbase.Request {
	s := &Span{
		logger: l,
	}
	s.attributes.Reset()
	return s
}

func (s *Span) Flush() {
	s.logger.writer.Flush()
}

func (s *Span) Boring(bool)                           {} // TODO
func (s *Span) ID() string                            { return s.logger.id.String() }
func (s *Span) SetErrorReporter(reporter func(error)) { s.errorFunc = reporter }

func (s *Span) Span(span trace.Bundle, name string) xopbase.Span {
	return s.logger.Request(span, name)
}

func (s *Span) getPrefill() *prefill {
	p := s.prefill.Load()
	if p == nil {
		return nil
	}
	return p.(*prefill)
}

func (s *Span) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
	l := &Line{
		level:     level,
		timestamp: t,
		span:      s,
		dataBuffer: xoputil.JBuilder{
			B:        make([]byte, 0, minBuffer),
			FastKeys: s.logger.fastKeys,
		},
	}
	l.encoder = json.NewEncoder(&l.dataBuffer)
	l.start(level, t, pc)
	return l
}

func (l *Line) Recycle(level xopconst.Level, t time.Time, pc []uintptr) {
	l.level = level
	l.timestamp = t
	l.dataBuffer.Reset()
	l.start(level, t, pc)
}

func (l *Line) start(level xopconst.Level, t time.Time, pc []uintptr) {
	l.dataBuffer.AppendByte('{') // }
	prefill := l.span.getPrefill()
	l.prefillLen = len(prefill.data)
	if prefill != nil {
		l.dataBuffer.AppendBytes(prefill.data)
	}
	l.dataBuffer.Comma()
	l.dataBuffer.AppendByte('{')
	l.Int("level", int64(level))
	l.Time("time", t)
	if l.span.logger.framesAtLevel[level] > 0 && len(pc) > 0 {
		n := l.span.logger.framesAtLevel[level]
		if n > len(pc) {
			n = len(pc)
		}
		frames := runtime.CallersFrames(pc[:n])
		l.dataBuffer.AppendBytes([]byte(`"stack":[`))
		for {
			frame, more := frames.Next()
			if !strings.Contains(frame.File, "runtime/") {
				break
			}
			l.dataBuffer.Comma()
			l.dataBuffer.AppendByte('"')
			l.dataBuffer.StringBody(frame.File)
			l.dataBuffer.AppendByte(':')
			l.dataBuffer.Int64(int64(frame.Line))
			l.dataBuffer.AppendByte('"')
			if !more {
				break
			}
		}
		l.dataBuffer.AppendByte(']')
	}
	l.dataBuffer.AppendByte('}')
}

func (l *Line) SetAsPrefill(m string) {
	skip := 1 + l.prefillLen
	prefill := prefill{
		msg:  m,
		data: make([]byte, len(l.dataBuffer.B)-skip),
	}
	copy(prefill.data, l.dataBuffer.B[skip:])
	l.span.prefill.Store(prefill)
	// this Line will not be recycled so destory its buffers
	l.reclaimMemory()
}

func (l *Line) Static(m string) {
	l.Msg(m) // TODO
}

func (l *Line) Msg(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.AppendBytes([]byte(`"msg":`))
	l.dataBuffer.String(m)
	// {
	l.dataBuffer.AppendByte('}')
	_, err := l.span.logger.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.errorFunc(err)
	}
	l.reclaimMemory()
}

func (l *Line) reclaimMemory() {
	if len(l.dataBuffer.B) > maxBufferToKeep {
		l.dataBuffer = xoputil.JBuilder{
			B:        make([]byte, 0, minBuffer),
			FastKeys: l.span.logger.fastKeys,
		}
		l.encoder = json.NewEncoder(&l.dataBuffer)
	}
}

func (l *Line) Template(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.AppendString(`"xop":"template","msg":`)
	l.dataBuffer.String(m)
	// {
	l.dataBuffer.AppendByte('}')
	_, err := l.span.logger.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.errorFunc(err)
	}
	l.reclaimMemory()
}

func (l *Line) Any(k string, v interface{}) {
	l.dataBuffer.Key(k)
	before := len(l.dataBuffer.B)
	err := l.encoder.Encode(v)
	if err != nil {
		l.dataBuffer.B = l.dataBuffer.B[:before]
		l.span.errorFunc(err)
		l.Error("encode:"+k, err)
	} else {
		// remove \n added by json.Encoder.Encode.  So helpful!
		if l.dataBuffer.B[len(l.dataBuffer.B)-1] == '\n' {
			l.dataBuffer.B = l.dataBuffer.B[:len(l.dataBuffer.B)-1]
		}
	}
}

func (l *Line) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	// TODO: send dictionary and numbers
	l.Int(k.Key(), v.Int64())
}

func (l *Line) Time(k string, t time.Time) {
	switch l.span.logger.timeOption {
	case strftimeTime:
		l.dataBuffer.Key(k)
		l.dataBuffer.AppendByte('"')
		l.dataBuffer.B = fasttime.AppendStrftime(l.dataBuffer.B, l.span.logger.timeFormat, t)
		l.dataBuffer.AppendByte('"')
	case timeTime:
		l.dataBuffer.Key(k)
		l.dataBuffer.AppendByte('"')
		l.dataBuffer.B = t.AppendFormat(l.dataBuffer.B, l.span.logger.timeFormat)
		l.dataBuffer.AppendByte('"')
	case epochTime:
		l.dataBuffer.Key(k)
		l.dataBuffer.Float64(float64(t.UnixNano()) / 1000000000.0) // TODO good enough?
	case unixNano:
		l.dataBuffer.Key(k)
		l.dataBuffer.Int64(t.UnixNano())
	}
}

func (l *Line) Link(k string, v trace.Trace) {
	// TODO: is this the right format for links?
	l.dataBuffer.Key(k)
	l.dataBuffer.AppendBytes([]byte(`{"xop.link":"`))
	l.dataBuffer.AppendString(v.HeaderString())
	l.dataBuffer.AppendBytes([]byte(`"}`))
}

func (l *Line) Bool(k string, v bool) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Bool(v)
}

func (l *Line) Int(k string, v int64) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Int64(v)
}

func (l *Line) Uint(k string, v uint64) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Uint64(v)
}

func (l *Line) Str(k string, v string) {
	l.dataBuffer.Key(k)
	l.dataBuffer.String(v)
}

func (l *Line) Number(k string, v float64) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Float64(v)
}

func (l *Line) Duration(k string, v time.Duration) {
	l.dataBuffer.Key(k)
	switch l.span.logger.durationFormat {
	case AsNanos:
		l.dataBuffer.Int64(int64(v / time.Nanosecond))
	case AsMillis:
		l.dataBuffer.Int64(int64(v / time.Millisecond))
	case AsSeconds:
		l.dataBuffer.Int64(int64(v / time.Second))
	case AsString:
		fallthrough
	default:
		l.dataBuffer.UncheckedString(v.String())
	}
}

// TODO: allow custom formats
func (l *Line) Error(k string, v error) {
	l.dataBuffer.Key(k)
	l.dataBuffer.String(v.Error())
}

func (s *Span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) { s.attributes.MetadataAny(k, v) }
func (s *Span) MetadataBool(k *xopconst.BoolAttribute, v bool)      { s.attributes.MetadataBool(k, v) }
func (s *Span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	s.attributes.MetadataEnum(k, v)
}
func (s *Span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.attributes.MetadataInt64(k, v) }
func (s *Span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.attributes.MetadataLink(k, v)
}
func (s *Span) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	s.attributes.MetadataNumber(k, v)
}
func (s *Span) MetadataStr(k *xopconst.StrAttribute, v string)      { s.attributes.MetadataStr(k, v) }
func (s *Span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.attributes.MetadataTime(k, v) }

// end