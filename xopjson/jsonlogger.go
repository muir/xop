// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"

	"github.com/google/uuid"
	"github.com/phuslu/fasttime"
)

const (
	maxBufferToKeep = 1024 * 10
	minBuffer       = 1024
	lineChanDepth   = 32
)

func New(w xopbytes.BytesWriter, opts ...Option) *Logger {
	log := &Logger{
		writer:       w,
		id:           uuid.New(),
		timeDivisor:  time.Millisecond,
		closeRequest: make(chan struct{}),
	}
	for _, f := range opts {
		f(log)
	}
	if log.tagOption == DefaultTagOption {
		if log.perRequestBufferLimit > 0 {
			log.tagOption = OmitTagOption
		} else {
			log.tagOption = FullIDTagOption
		}
	}
	return log
}

func (logger *Logger) ID() string           { return logger.id.String() }
func (logger *Logger) Buffered() bool       { return logger.writer.Buffered() }
func (logger *Logger) ReferencesKept() bool { return false }

func (logger *Logger) Close() {
	logger.writer.Close()
}

func (logger *Logger) Request(trace trace.Bundle, name string) xopbase.Request {
	r := &request{
		span: span{
			logger: logger,
			writer: logger.writer.Request(trace),
			trace:  trace,
			name:   name,
		},
	}
	if logger.tagOption == TraceSequenceNumberTagOption {
		r.idNum = atomic.AddInt64(&logger.requestCount, 1)
	}
	r.attributes.Reset()
	r.request = r
	if logger.perRequestBufferLimit != 0 {
		r.maintainBuffer()
	}
	return r
}

func (r *request) maintainBuffer() {
	r.flushRequest = make(chan struct{})
	r.flushComplete = make(chan struct{})
	r.completedLines = make(chan *line, lineChanDepth)
	r.writeBuffer = make([]byte, 0, r.logger.perRequestBufferLimit/16)
	go func() {
		for {
			select {
			case <-r.logger.closeRequest:
				r.flushBuffer()
				// TODO: have logger wait for requests to complete
				// WaitGroup?
				return
			case <-r.flushRequest:
				r.flushBuffer()
				r.flushComplete <- struct{}{}
			case line := <-r.completedLines:
				if len(line.dataBuffer.B)+len(r.writeBuffer) > r.logger.perRequestBufferLimit {
					r.flushBuffer()
				}
				if len(line.dataBuffer.B) > r.logger.perRequestBufferLimit {
					// TODO: split into multiple writes
				}
				r.writeBuffer = append(r.writeBuffer, line.dataBuffer.B...)
				line.reclaimMemory()
			}
		}
	}()
}

func (r *request) flushBuffer() {
	// TODO: trigger spans to write their stuff
	if len(r.writeBuffer) == 0 {
		return
	}
	_, err := r.writer.Write(r.writeBuffer)
	if err != nil {
		r.errorFunc(err)
	}
	r.writer.Flush()
	r.writeBuffer = r.writeBuffer[:0]
}

func (r *request) Flush() {
	if r.logger.perRequestBufferLimit != 0 {
		// TODO: improve this a bit by using a WaitGroup or something
		r.flushRequest <- struct{}{}
		<-r.flushComplete
	} else {
		r.writer.Flush()
	}
}

func (r *request) SetErrorReporter(reporter func(error)) { r.errorFunc = reporter }

func (s *span) Span(ts time.Time, trace trace.Bundle, name string) xopbase.Span {
	n := &span{
		logger:    s.logger,
		writer:    s.writer,
		trace:     trace,
		name:      name,
		request:   s.request,
		startTime: ts,
	}
	n.attributes.Reset()
	return n
}

func (s *span) Done(t time.Time) { atomic.StoreInt64(&s.endTime, t.UnixNano()) }
func (s *span) Boring(bool)      {} // TODO
func (s *span) ID() string       { return s.logger.id.String() }

func (s *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		span: s,
	}
}

func (s *span) builder(attributesWanted bool) builder {
	b := builder{
		span: s,
		dataBuffer: xoputil.JBuilder{
			B:        make([]byte, 0, minBuffer),
			FastKeys: s.logger.fastKeys,
		},
		attributesWanted: attributesWanted,
	}
	b.encoder = json.NewEncoder(&b.dataBuffer)
	return b
}

func (s *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builder: s.builder(false),
	}
}

func (p *prefilling) PrefillComplete(m string) xopbase.Prefilled {
	prefilled := &prefilled{
		data: make([]byte, len(p.builder.dataBuffer.B)),
		span: p.builder.span,
	}
	copy(prefilled.data, p.builder.dataBuffer.B)
	if len(m) > 0 {
		msgBuffer := xoputil.JBuilder{
			B: make([]byte, len(m)), // alloc-per-prefill
		}
		msgBuffer.StringBody(m)
		prefilled.preEncodedMsg = msgBuffer.B
	}
	return prefilled
}

func (p *prefilled) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
	atomic.StoreInt64(&p.span.endTime, t.UnixNano())
	l := &line{
		builder:              p.span.builder(p.span.logger.attributesObject),
		level:                level,
		timestamp:            t,
		prefillMsgPreEncoded: p.preEncodedMsg,
	}
	l.dataBuffer.AppendByte('{') // }
	if !l.attributesWanted {
		l.dataBuffer.AppendBytes([]byte(`"zop":{`)) // }
	}
	l.Int("lvl", int64(level))
	l.Time("ts", t)
	if len(pc) > 0 {
		frames := runtime.CallersFrames(pc)
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
	switch l.span.logger.tagOption {
	case SpanIDTagOption:
		l.dataBuffer.Comma()
		l.dataBuffer.AppendBytes([]byte(`"span_id":"`))
		l.dataBuffer.AppendString(l.span.trace.Trace.SpanIDString())
		l.dataBuffer.AppendByte('"')
	case FullIDTagOption:
		l.dataBuffer.Comma()
		l.dataBuffer.AppendBytes([]byte(`"trace_header":"`))
		l.dataBuffer.AppendString(l.span.trace.Trace.HeaderString())
		l.dataBuffer.AppendByte('"')
	case TraceIDTagOption:
		l.dataBuffer.Comma()
		l.dataBuffer.AppendBytes([]byte(`"trace_id":"`))
		l.dataBuffer.AppendString(l.span.trace.Trace.TraceIDString())
		l.dataBuffer.AppendByte('"')
	case TraceSequenceNumberTagOption:
		l.dataBuffer.Key("trace_num")
		l.dataBuffer.Int64(l.span.request.idNum)
	case OmitTagOption:
		// yay!
	}
	if !l.attributesWanted {
		// {
		l.dataBuffer.AppendByte('}')
	}
	if len(p.data) != 0 {
		if l.attributesWanted {
			l.attributesStarted = true
			l.dataBuffer.AppendBytes([]byte(`"attributes":{`)) // }
		} else {
			l.dataBuffer.AppendByte(',')
		}
		l.dataBuffer.AppendBytes(p.data)
	}
	return l
}

func (l *line) Msg(m string) {
	if l.attributesStarted {
		// {
		l.dataBuffer.AppendByte('}')
	}
	l.dataBuffer.AppendBytes([]byte(`,"msg":"`))
	if len(l.prefillMsgPreEncoded) != 0 {
		l.dataBuffer.AppendBytes(l.prefillMsgPreEncoded)
	}
	l.dataBuffer.StringBody(m)
	// {
	l.dataBuffer.AppendBytes([]byte{'"', '}'})
	if l.span.logger.perRequestBufferLimit != 0 {
		l.span.request.completedLines <- l
	} else {
		_, err := l.span.writer.Write(l.dataBuffer.B)
		if err != nil {
			l.span.request.errorFunc(err)
		}
		l.reclaimMemory()
	}
}

func (l *line) Static(m string) {
	l.Msg(m) // TODO
}

func (l *line) reclaimMemory() {
	if len(l.dataBuffer.B) > maxBufferToKeep {
		l.dataBuffer = xoputil.JBuilder{
			B:        make([]byte, 0, minBuffer),
			FastKeys: l.span.logger.fastKeys,
		}
		l.encoder = json.NewEncoder(&l.dataBuffer)
	}
	// TODO have pool of Lines & Buffers
}

func (l *line) Template(m string) {
	if l.attributesStarted {
		// {
		l.dataBuffer.AppendByte('}')
	}
	l.dataBuffer.AppendString(`,"fmt":"tmpl","msg":`)
	l.dataBuffer.String(m)
	// {
	l.dataBuffer.AppendByte('}')
	_, err := l.span.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.request.errorFunc(err)
	}
	l.reclaimMemory()
}

func (b *builder) startAttributes() {
	if b.attributesWanted && !b.attributesStarted {
		b.attributesStarted = true
		b.dataBuffer.AppendBytes([]byte(`"attributes":{`)) // }
	}
}

func (b *builder) Any(k string, v interface{}) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	before := len(b.dataBuffer.B)
	err := b.encoder.Encode(v)
	if err != nil {
		b.dataBuffer.B = b.dataBuffer.B[:before]
		b.span.request.errorFunc(err)
		b.Error("encode:"+k, err)
	} else {
		// remove \n added by json.Encoder.Encode.  So helpful!
		if b.dataBuffer.B[len(b.dataBuffer.B)-1] == '\n' {
			b.dataBuffer.B = b.dataBuffer.B[:len(b.dataBuffer.B)-1]
		}
	}
}

func (b *builder) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	b.startAttributes()
	// TODO: send dictionary and numbers
	b.Int(k.Key(), v.Int64())
}

func (b *builder) Time(k string, t time.Time) {
	b.startAttributes()
	switch b.span.logger.timeOption {
	case strftimeTime:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.B = fasttime.AppendStrftime(b.dataBuffer.B, b.span.logger.timeFormat, t)
		b.dataBuffer.AppendByte('"')
	case timeTimeFormat:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.B = t.AppendFormat(b.dataBuffer.B, b.span.logger.timeFormat)
		b.dataBuffer.AppendByte('"')
	case epochTime:
		b.dataBuffer.Key(k)
		b.dataBuffer.Int64(t.UnixNano() / int64(b.span.logger.timeDivisor))
	case epochQuoted:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.Int64(t.UnixNano() / int64(b.span.logger.timeDivisor))
		b.dataBuffer.AppendByte('"')
	}
}

func (b *builder) Link(k string, v trace.Trace) {
	b.startAttributes()
	// TODO: is this the right format for links?
	b.dataBuffer.Key(k)
	b.dataBuffer.AppendBytes([]byte(`{"xop.link":"`))
	b.dataBuffer.AppendString(v.HeaderString())
	b.dataBuffer.AppendBytes([]byte(`"}`))
}

func (b *builder) Bool(k string, v bool) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	b.dataBuffer.Bool(v)
}

func (b *builder) Int(k string, v int64) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	b.dataBuffer.Int64(v)
}

func (b *builder) Uint(k string, v uint64) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	b.dataBuffer.Uint64(v)
}

func (b *builder) Str(k string, v string) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	b.dataBuffer.String(v)
}

func (b *builder) Float64(k string, v float64) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	b.dataBuffer.Float64(v)
}

func (b *builder) Duration(k string, v time.Duration) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	switch b.span.logger.durationFormat {
	case AsNanos:
		b.dataBuffer.Int64(int64(v / time.Nanosecond))
	case AsMillis:
		b.dataBuffer.Int64(int64(v / time.Millisecond))
	case AsSeconds:
		b.dataBuffer.Int64(int64(v / time.Second))
	case AsString:
		fallthrough
	default:
		b.dataBuffer.UncheckedString(v.String())
	}
}

// TODO: allow custom formats
func (b *builder) Error(k string, v error) {
	b.startAttributes()
	b.dataBuffer.Key(k)
	b.dataBuffer.String(v.Error())
}

func (s *span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) { s.attributes.MetadataAny(k, v) }
func (s *span) MetadataBool(k *xopconst.BoolAttribute, v bool)      { s.attributes.MetadataBool(k, v) }
func (s *span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	s.attributes.MetadataEnum(k, v)
}
func (s *span) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	s.attributes.MetadataFloat64(k, v)
}
func (s *span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.attributes.MetadataInt64(k, v) }
func (s *span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.attributes.MetadataLink(k, v)
}
func (s *span) MetadataStr(k *xopconst.StrAttribute, v string)      { s.attributes.MetadataStr(k, v) }
func (s *span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.attributes.MetadataTime(k, v) }

// end
