// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// SpanLog allows xop to add logs to an existing OTEL span.  log.Done() will be
// ignored for this span.
func SpanLog(ctx context.Context, name string, extraModifiers ...xop.SeedModifier) *xop.Log {
	span := oteltrace.SpanFromContext(ctx)
	var xoptrace xoptrace.Trace
	xoptrace.TraceID().SetArray(span.SpanContext().TraceID())
	xoptrace.SpanID().SetArray(span.SpanContext().SpanID())
	xoptrace.Flags().SetArray([1]byte{byte(span.SpanContext().TraceFlags())})
	xoptrace.Version().SetArray([1]byte{1})
	log := xop.NewSeed(
		xop.CombineSeedModifiers(extraModifiers...),
		xop.WithContext(ctx),
		xop.WithTrace(xoptrace),
		xop.WithBase(&logger{
			id:         "otel-" + uuid.New().String(),
			doLogging:  true,
			ignoreDone: span,
			tracer:     span.TracerProvider().Tracer(""),
		}),
		// The first time through, we do not want to change the spanID,
		// but on subsequent calls, we do so the outer reactive function
		// just sets the future function.
		xop.WithReactive(func(ctx context.Context, seed xop.Seed, nameOrDescription string, isChildSpan bool) []xop.SeedModifier {
			return []xop.SeedModifier{
				xop.WithTrace(xoptrace),
				xop.WithReactiveReplaced(
					func(ctx context.Context, seed xop.Seed, nameOrDescription string, isChildSpan bool) []xop.SeedModifier {
						var newSpan oteltrace.Span
						// XXX add WithTimestamp
						// XXX add WithAttributes
						// XXX add TraceState
						// XXX add Bundle
						if isChildSpan {
							ctx, newSpan = span.TracerProvider().Tracer("").Start(ctx, nameOrDescription, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
						} else {
							ctx, newSpan = span.TracerProvider().Tracer("").Start(ctx, nameOrDescription, oteltrace.WithSpanKind(oteltrace.SpanKindServer))
						}
						return []xop.SeedModifier{
							xop.WithContext(ctx),
							xop.WithSpan(newSpan.SpanContext().SpanID()),
						}
					}),
			}
		}),
	).SubSpan(name)
	go func() {
		<-ctx.Done()
		log.Done()
	}()
	return log
}

// BaseLogger provides SeedModifiers to set up an OTEL Tracer as a xopbase.Logger
// so that xop logs are output through the OTEL Tracer.
func BaseLogger(ctx context.Context, tracer oteltrace.Tracer, doLogging bool) xop.SeedModifier {
	return xop.CombineSeedModifiers(
		xop.WithBase(&logger{
			id:        "otel-" + uuid.New().String(),
			doLogging: doLogging,
			tracer:    tracer,
		}),
		xop.WithContext(ctx),
		xop.WithReactive(func(ctx context.Context, seed xop.Seed, nameOrDescription string, isChildSpan bool) []xop.SeedModifier {
			if isChildSpan {
				ctx, span := tracer.Start(ctx, nameOrDescription, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
				// XXX add WithTimestamp
				// XXX add WithAttributes -- from seed
				// XXX add TraceState
				// XXX add Bundle
				return []xop.SeedModifier{
					xop.WithContext(ctx),
					xop.WithSpan(span.SpanContext().SpanID()),
				}
			}
			si := seed.SourceInfo()
			ctx, span := tracer.Start(ctx, nameOrDescription,
				oteltrace.WithNewRoot(),
				// TODO: use runtime/debug ReadBuildInfo to get the version of xoputil
				// XXX add WithTimestamp
				// XXX add WithAttributes -- from seed
				oteltrace.WithAttributes(
					xopVersion.String("0.0.1"),
					xopOTELVersion.String("0.0.1"),
					xopSource.String(si.Source+" "+si.SourceVersion.String()),
					xopNamespace.String(si.Namespace+" "+si.NamespaceVersion.String()),
				),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			)
			bundle := seed.Bundle()
			if bundle.Parent.IsZero() {
				bundle.State.SetString(span.SpanContext().TraceState().String())
				bundle.Trace.Flags().SetArray([1]byte{byte(span.SpanContext().TraceFlags())})
				bundle.Trace.Version().SetArray([1]byte{1})
				bundle.Trace.TraceID().SetArray(span.SpanContext().TraceID())
			}
			bundle.Trace.SpanID().SetArray(span.SpanContext().SpanID())
			return []xop.SeedModifier{
				xop.WithContext(ctx),
				xop.WithBundle(bundle),
			}
		}),
	)
}

func (logger *logger) ID() string           { return logger.id }
func (logger *logger) ReferencesKept() bool { return true }
func (logger *logger) Buffered() bool       { return false }

func (logger *logger) Request(ctx context.Context, ts time.Time, _ xoptrace.Bundle, description string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	// we ignore the Bundle because we've already recorded that information
	// in the OTEL span that we've already created.
	s := logger.span(ctx, ts, description, "")
	s.span.SetAttributes(
		xopSource.String(sourceInfo.Source+" "+sourceInfo.SourceVersion),
		xopNamespace.String(sourceInfo.Namespace+" "+sourceInfo.NamespaceVersion),
	)
	s.request = &request{
		span:              s,
		attributesDefined: make(map[string]struct{}),
	}
	return s.request
}

func (span *span) Flush()                         {}
func (span *span) Final()                         {}
func (span *span) SetErrorReporter(f func(error)) {}
func (span *span) Boring(_ bool)                  {}
func (span *span) ID() string                     { return span.logger.id }
func (span *span) Done(endTime time.Time, final bool) {
	if !final {
		return
	}
	if span.logger.ignoreDone == span.span {
		// skip Done for spans passed in to SpanLog()
		return
	}
	span.span.End()
}

func (span *span) Span(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, spanSequenceCode string) xopbase.Span {
	s := span.logger.span(ctx, ts, description, spanSequenceCode)
	s.request = span.request
	if spanSequence != "" {
		s.span.SetAttributes(logSpanSequence.String(spanSequence))
	}
	return s
}

func (logger *logger) span(ctx context.Context, ts time.Time, description string, spanSequence string) *span {
	otelSpan := oteltrace.SpanFromContext(ctx)
	return &span{
		logger: logger,
		span:   otelSpan,
		ctx:    ctx,
	}
}

func (span *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		builder: builder{
			span: span,
		},
	}
}

func (span *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builder: builder{
			span: span,
		},
	}
}

func (prefill *prefilling) PrefillComplete(msg string) xopbase.Prefilled {
	prefill.builder.prefillMsg = msg
	return &prefilled{
		builder: prefill.builder,
	}
}

func (prefilled *prefilled) Line(level xopnum.Level, _ time.Time, pc []uintptr) xopbase.Line {
	if !prefilled.span.logger.doLogging || !prefilled.span.span.IsRecording() {
		return xopbase.SkipLine
	}
	// PERFORMANCE: get line from a pool
	line := &line{}
	line.level = level
	line.span = prefilled.span
	line.attributes = line.prealloc[:0]
	line.attributes = append(line.attributes, prefilled.span.spanPrefill...)
	line.attributes = append(line.attributes, prefilled.attributes...)
	line.prefillMsg = prefilled.prefillMsg
	line.linkKey = prefilled.linkKey
	line.linkValue = prefilled.linkValue
	if len(pc) > 0 {
		var b strings.Builder
		frames := runtime.CallersFrames(pc)
		for {
			frame, more := frames.Next()
			if strings.Contains(frame.File, "runtime/") {
				break
			}
			b.WriteString(frame.File)
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(frame.Line))
			b.WriteByte('\n')
			if !more {
				break
			}
		}
		line.attributes = append(line.attributes, semconv.ExceptionStacktraceKey.String(b.String()))
	}
	return line
}

func (line *line) Link(k string, v xoptrace.Trace) {
	line.attributes = append(line.attributes,
		logMessageKey.String(line.prefillMsg+k),
		typeKey.String("link"),
		xopLinkData.String(v.String()),
	)
	line.done()
	_, tmpSpan := line.span.logger.tracer.Start(line.span.ctx, k,
		oteltrace.WithLinks(
			oteltrace.Link{
				SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
					TraceID:    v.TraceID().Array(),
					SpanID:     v.SpanID().Array(),
					TraceFlags: oteltrace.TraceFlags(v.Flags().Array()[0]),
					TraceState: emptyTraceState, // TODO: is this right?
					Remote:     true,            // information not available
				}),
			}),
		oteltrace.WithAttributes(
			spanIsLinkEventKey.Bool(true),
		),
	)
	tmpSpan.AddEvent(line.level.String(), oteltrace.WithAttributes(line.attributes...))
	tmpSpan.SetAttributes(typeKey.String("link-event"))
	tmpSpan.End()
}

func (line *line) Model(msg string, v xopbase.ModelArg) {
	v.Encode()
	line.attributes = append(line.attributes,
		logMessageKey.String(line.prefillMsg+msg),
		typeKey.String("model"),
		xopModelType.String(v.TypeName),
		xopEncoding.String(v.Encoding.String()),
		xopModel.String(string(v.Encoded)),
	)
	line.done()
}

func (line *line) Msg(msg string) {
	line.attributes = append(line.attributes, logMessageKey.String(line.prefillMsg+msg), typeKey.String("line"))
	line.done()
	// PERFORMANCE: return line to pool
}

func (line *line) done() {
	line.span.span.AddEvent(line.level.String(), oteltrace.WithAttributes(line.attributes...))
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

func (line *line) Template(template string) {
	kv := make(map[string]int)
	for i, a := range line.attributes {
		kv[string(a.Key)] = i
	}
	msg := templateRE.ReplaceAllStringFunc(template, func(k string) string {
		k = k[1 : len(k)-1]
		if i, ok := kv[k]; ok {
			a := line.attributes[i]
			switch a.Value.Type() {
			case attribute.BOOL:
				return strconv.FormatBool(a.Value.AsBool())
			case attribute.INT64:
				return strconv.FormatInt(a.Value.AsInt64(), 10)
			case attribute.FLOAT64:
				return strconv.FormatFloat(a.Value.AsFloat64(), 'g', -1, 64)
			case attribute.STRING:
				return a.Value.AsString()
			case attribute.BOOLSLICE:
				return fmt.Sprint(a.Value.AsBoolSlice())
			case attribute.INT64SLICE:
				return fmt.Sprint(a.Value.AsInt64Slice())
			case attribute.FLOAT64SLICE:
				return fmt.Sprint(a.Value.AsFloat64Slice())
			case attribute.STRINGSLICE:
				return fmt.Sprint(a.Value.AsStringSlice())
			default:
				return "{" + k + "}"
			}
		}
		return "''"
	})
	line.attributes = append(line.attributes,
		logMessageKey.String(line.prefillMsg+msg),
		typeKey.String("line"),
		xopLineFormat.String("tmpl"),
		xopTemplate.String(template),
	)
	line.done()
}

func (builder *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.Key(), []string{v.String(), "enum", strconv.FormatInt(v.Int64(), 10)}))
}

func (builder *builder) Any(k string, v xopbase.ModelArg) {
	v.Encode()
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{string(v.Encoded), "any", v.Encoding.String(), v.TypeName}))
}

func (builder *builder) Time(k string, v time.Time) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{v.Format(time.RFC3339Nano), "time"}))
}

func (builder *builder) Duration(k string, v time.Duration) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{v.String(), "dur"}))
}

func (builder *builder) Uint64(k string, v uint64, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{strconv.FormatUint(v, 10), xopbase.DataTypeToString[dt]}))
}

func (builder *builder) Int64(k string, v int64, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{strconv.FormatInt(v, 10), xopbase.DataTypeToString[dt]}))
}

func (builder *builder) Float64(k string, v float64, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{strconv.FormatFloat(v, 'g', -1, 64), xopbase.DataTypeToString[dt]}))
}

func (builder *builder) String(k string, v string, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k, []string{v, xopbase.DataTypeToString[dt]}))
}

func (builder *builder) Bool(k string, v bool) {
	builder.attributes = append(builder.attributes, attribute.Bool(k, v))
}

func (span *span) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	enc, err := v.MarshalJSON()
	var value string
	if err != nil {
		value = fmt.Sprintf("[zopotel] could not marshal %T value: %s", v, err)
	} else {
		value = string(enc)
	}
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataBool(k *xopat.BoolAttribute, v bool) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.Bool(key, values))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[bool]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[bool]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorBoolSlices == nil {
		span.priorBoolSlices = make(map[string][]bool)
	}
	s := span.priorBoolSlices[key]
	s = append(s, value)
	span.priorBoolSlices[key] = s
	span.span.SetAttributes(attribute.BoolSlice(key, s))
}

func (span *span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v.String() + "/" + strconv.FormatInt(v.Int64(), 10)
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.Float64(key, values))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[float64]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[float64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorFloat64Slices == nil {
		span.priorFloat64Slices = make(map[string][]float64)
	}
	s := span.priorFloat64Slices[key]
	s = append(s, value)
	span.priorFloat64Slices[key] = s
	span.span.SetAttributes(attribute.Float64Slice(key, s))
}

func (span *span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.Int64(key, values))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[int64]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[int64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorInt64Slices == nil {
		span.priorInt64Slices = make(map[string][]int64)
	}
	s := span.priorInt64Slices[key]
	s = append(s, value)
	span.priorInt64Slices[key] = s
	span.span.SetAttributes(attribute.Int64Slice(key, s))
}

func (span *span) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v.String()
	_, tmpSpan := span.logger.tracer.Start(span.ctx, k.Key(),
		oteltrace.WithLinks(
			oteltrace.Link{
				SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
					TraceID:    v.TraceID().Array(),
					SpanID:     v.SpanID().Array(),
					TraceFlags: oteltrace.TraceFlags(v.Flags().Array()[0]),
					TraceState: emptyTraceState, // TODO: is this right?
					Remote:     true,            // information not available
				}),
			}),
		oteltrace.WithAttributes(
			spanIsLinkAttributeKey.Bool(true),
		),
	)
	tmpSpan.SetAttributes(spanIsLinkAttributeKey.Bool(true))
	tmpSpan.End()
	value := v.String() + "/" + strconv.FormatInt(v.Int64(), 10)
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataString(k *xopat.StringAttribute, v string) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	key := k.Key()
	if _, ok := span.request.attributesDefined[key]; !ok {
		if k.Description() != xopSynthesizedForOTEL {
			span.span.SetAttributes(attribute.String(attributeDefinitionPrefix+key, k.DefinitionJSONString()))
			span.request.attributesDefined[key] = struct{}{}
		}
	}
	value := v.Format(time.RFC3339Nano)
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}
