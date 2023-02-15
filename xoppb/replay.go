// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoppb

import (
	"bytes"
	"context"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

func (log *Logger) Replay(ctx context.Context, input any, logger xopbase.Logger) error {
	return log.LosslessReplay(ctx, input, logger)
}

func (_ *Logger) LosslessReplay(ctx context.Context, input any, logger xopbase.Logger) error {
	trace, ok := input.(*xopproto.Trace)
	if !ok {
		return errors.Errorf("expected *xopproto.Trace for xoppb.Replay, got %T", input)
	}
	for _, request := range trace.Requests {
		err := replayRequest{
			logger:       logger,
			requestInput: request,
			traceID:      xoptrace.NewHexBytes16FromSlice(trace.TraceID),
			registry:     xopat.NewRegistry(false),
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type replayRequest struct {
	logger       xopbase.Logger
	requestInput *xopproto.Request
	traceID      xoptrace.HexBytes16
	request      xopbase.Request
	spansSeen    map[xoptrace.HexBytes8]spanData
	registry     *xopat.Registry
}

type spanData struct {
	version int32
	span    xopbase.Span
}

func (x replayRequest) Replay(ctx context.Context) error {
	requestID := xoptrace.NewHexBytes8FromSlice(x.requestInput.RequestID)
	if len(x.requestInput.Spans) == 0 {
		return errors.Errorf("expected >0 spans in request (%s-%s)", x.traceID, requestID)
	}
	if !bytes.Equal(x.requestInput.RequestID, x.requestInput.Spans[0].SpanID) {
		return errors.Errorf("expected first span of request (%s-%s) to be the request (got %s)", x.traceID, requestID, x.requestInput.Spans[0].SpanID)
	}
	requestSpan := x.requestInput.Spans[0]
	var bundle xoptrace.Bundle
	bundle.Trace.TraceID().Set(x.traceID)
	bundle.Trace.SpanID().Set(requestID)
	bundle.Parent.SpanID().SetBytes(requestSpan.ParentID)
	if len(x.requestInput.ParentTraceID) != 0 {
		bundle.Parent.TraceID().SetBytes(x.requestInput.ParentTraceID)
	} else {
		bundle.Parent.TraceID().Set(x.traceID)
	}
	bundle.State.SetString(requestSpan.TraceState)
	bundle.Baggage.SetString(requestSpan.Baggage)

	sourceInfo := xopbase.SourceInfo{
		Source:    x.requestInput.SourceID,
		Namespace: x.requestInput.SourceNamespace,
	}
	var err error
	sourceInfo.SourceVersion, err = semver.StrictNewVersion(x.requestInput.SourceVersion)
	if err != nil {
		return errors.Wrapf(err, "invalid source version in request (%s-%s)", x.traceID, requestID)
	}
	sourceInfo.NamespaceVersion, err = semver.StrictNewVersion(x.requestInput.SourceNamespaceVersion)
	if err != nil {
		return errors.Wrapf(err, "invalid namespace version in request (%s-%s)", x.traceID, requestID)
	}

	x.request = x.logger.Request(ctx,
		time.Unix(0, requestSpan.StartTime),
		bundle,
		requestSpan.Name,
		sourceInfo)
	x.spansSeen = make(map[xoptrace.HexBytes8]spanData)
	for i := len(x.requestInput.Spans) - 1; i > 0; i-- { // 0 is processed above
		err = replaySpan{
			replayRequest: x,
			spanInput:     x.requestInput.Spans[i],
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}

	for _, line := range x.requestInput.Lines {
		spanID := xoptrace.NewHexBytes8FromSlice(line.SpanID)
		span, ok := x.spansSeen[spanID]
		if !ok {
			return errors.Errorf("line references spanID (%s) that does not exist", spanID)
		}
		err := replayLine{
			replayRequest: x,
			span:          span,
			lineInput:     line,
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type replaySpan struct {
	replayRequest
	spanInput *xopproto.Span
	span      xopbase.Span
}

func (x replaySpan) Replay(ctx context.Context) error {
	spanID := xoptrace.NewHexBytes8FromSlice(x.spanInput.SpanID)
	if sd, ok := x.spansSeen[spanID]; ok && sd.version > x.spanInput.Version {
		return nil
	}
	var bundle xoptrace.Bundle
	bundle.Trace.TraceID().Set(x.traceID)
	bundle.Trace.SpanID().Set(spanID)
	bundle.Parent.SpanID().SetBytes(x.spanInput.ParentID)
	bundle.Parent.TraceID().Set(x.traceID)
	x.span = x.request.Span(ctx,
		time.Unix(0, x.spanInput.StartTime),
		bundle,
		x.spanInput.Name,
		x.spanInput.SequenceCode)
	x.spansSeen[spanID] = spanData{
		version: x.spanInput.Version,
		span:    x.span,
	}
	for _, attribute := range x.spanInput.Attributes {
		err := x.replayAttribute(attribute)
		if err != nil {
			return err
		}
	}
	if x.spanInput.EndTime != nil {
		x.span.Done(time.Unix(0, *x.spanInput.EndTime), false)
	}
	return nil
}

func (x replaySpan) replayAttribute(attribute *xopproto.SpanAttribute) error {
	def := x.requestInput.AttributeDefinitions[attribute.AttributeDefinitionSequenceNumber]
	m := xopat.Make{
		Key:         def.Key,
		Description: def.Description,
		Namespace:   def.Namespace + " " + def.NamespaceSemver,
		Indexed:     def.ShouldIndex,
		Prominence:  int(def.Prominence),
		Multiple:    def.Multiple,
		Distinct:    def.Distinct,
		// XXX Ranged      :
		Locked: def.Locked,
	}
	switch xopat.AttributeType(def.Type) {
	case xopat.AttributeTypeAny:
		registeredAttribute, err := x.registry.ConstructAnyAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataAny(registeredAttribute, xopbase.ModelArg{
				TypeName: v.StringValue,
				Encoded:  v.BytesValue,
				Encoding: xopproto.Encoding(v.IntValue),
			})
		}
		return nil
	case xopat.AttributeTypeBool:
		registeredAttribute, err := x.registry.ConstructBoolAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			var b bool
			if v.IntValue != 0 {
				b = true
			}
			x.span.MetadataBool(registeredAttribute, b)
		}
		return nil
	case xopat.AttributeTypeEnum:
		registeredAttribute, err := x.registry.ConstructEnumAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			enum := registeredAttribute.Add64(v.IntValue, v.StringValue)
			x.span.MetadataEnum(&registeredAttribute.EnumAttribute, enum)
		}
		return nil
	case xopat.AttributeTypeFloat64:
		registeredAttribute, err := x.registry.ConstructFloat64Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataFloat64(registeredAttribute, v.FloatValue)
		}
		return nil
	case xopat.AttributeTypeInt64:
		registeredAttribute, err := x.registry.ConstructInt64Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataInt64(registeredAttribute, v.IntValue)
		}
		return nil
	case xopat.AttributeTypeLink:
		registeredAttribute, err := x.registry.ConstructLinkAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			t, ok := xoptrace.TraceFromString(v.StringValue)
			if !ok {
				return errors.Errorf("invalid trace attribute '%s'", v.StringValue)
			}
			x.span.MetadataLink(registeredAttribute, t)
		}
		return nil
	case xopat.AttributeTypeString:
		registeredAttribute, err := x.registry.ConstructStringAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataString(registeredAttribute, v.StringValue)
		}
		return nil
	case xopat.AttributeTypeTime:
		registeredAttribute, err := x.registry.ConstructTimeAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataTime(registeredAttribute, time.Unix(0, v.IntValue))
		}
		return nil

	default:
		return errors.Errorf("unexpected attribute type %s", def.Type)
	}
}

type replayLine struct {
	replayRequest
	lineInput *xopproto.Line
	span      xopbase.Span
}

func (x replayLine) Replay(ctx context.Context) error {
	line := span.NoPrefill().Line(
		xopnum.Level(x.lineInput.LogLevel),
		time.Unix(0, x.lineInput.Timestamp),
		nil, // XXX todo
	)
	for _, attribute := range line.Attributes {
		switch attribute.Type {
		case xopproto.AttributeType_Enum:
			m := xopat.Make{
				Key: attribute.Key,
			}
			ea, err := x.registry.ConstructEnumAttribute(m)
			enum := ea.Add64(attribute.Value.IntValue, attribute.Value.StringValue)
			line = line.Enum(&ea.EnumAttribute, enum)
		case xopproto.AttributeType_Float64:
			line = line.Float64(attribute.Key, attribute.Value.FloatValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_Int64:
			line = line.Int64(attribute.Key, attribute.Value.IntValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_String:
			line = line.String(attribute.Key, attribute.Value.StringValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_Uint64:
			line = line.Uint64(attribute.Key, attribute.Value.UintValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_Any:
			line = line.Any(attribute.Key, models.ModelArg{
				TypeName: attribute.Value.StringValue,
				Encoded:  attribute.Value.BytesValue,
				Encoding: xopproto.Encoding(attribute.Value.IntValue),
			})
		case xopproto.AttributeType_Bool:
			var b bool
			if attribute.Value.IntValue != 0 {
				b = true
			}
			line = line.Any(attribute.Key, b)
		case xopproto.AttributeType_Duration:
			line = line.Duration(attribute.Key, time.Duration(attribute.Value.IntValue))
		case xopproto.AttributeType_Time:
			line = line.Time(attribute.Key, time.Unix(0, attribute.Value.IntValue))
		default:
			return errors.Errorf("unknown data type %s", attribute.Type)
		}
	}
	switch {
	case x.lineInput.Model != nil:
		line.Model(x.lineInput.Message, xopbase.ModelArg{
			TypeName: x.lineInput.Model.Type,
			Encoded:  x.lineInput.Model.Encoded,
			Encoding: x.lineInput.Model.Encoding,
		})
	case x.lineInput.Link != "":
		trace, ok := xoptrace.TraceFromString(x.lineInput.Link)
		if !ok {
			return errors.Errorf("invalid trace (%s)", x.lineInput.Link)
		}
		line.Link(x.lineInput.Message, trace)
	case x.lineInput.MessageTemplate != "":
		line.Template(x.lineInput.MessageTemplate)
	default:
		line.Msg(x.lineInput.Message)
	}
	return nil
}
