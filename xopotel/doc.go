/*
Package xopotel provides a gateways between XOP and Open Telemetry.

There is a mismatch of features between Open Telemetry and XOP. Open Telemetry
supports only a very limited set of attribute types. When gatewaying from
XOP into Open Telemetry the richer set of types are almost always converted
to string slices.

There are several integration points.

# BaseLogger

The BaseLogger() function returns a xopbase.Logger that can be used like
any other base logger to configure XOP output. In this case, the XOP logs
and traces will be output through the Open Telemtry system using the
primary interfaces of TracerProvider, Tracer, Span, etc.  There is a
restriction though: to use this you MUST create the TracerProvider with
the xopotel IDGenerator:

	import (
		"github.com/xoplog/xop-go/xopotel"
		sdktrace "go.opentelemetry.io/otel/sdk/trace"
	)

	tracerProvider := NewTraceProvider(xopotel.IDGenerator(), sdktrace.WithBatcher(...))

This allows the TraceIDs and SpanIDs created by XOP to be used by
Open Telemetry.

# SeedModifier

If for some reason, you do not have control over the creation of your TracerProvider,
you can use SeedModifer() modify your xop.Seed so that it delgates SpanID and TraceID
creation to Open Telemetry.

# SpanToLog

If you don't have access to a TracerProvider at all and instead have
a "go.opentelemetry.io/otel/trace".Span, you can use that as the basis for generating
logs with XOP by converting it directly to a *xop.Log.

# ExportToXOP

Integration can go the other direction. You can flow traces from Open Telemetry to
XOP base loggers. Use ExportToXOP() to wrap a xopbase.Logger so that it can be used
as a SpanExporter.

# ExporterAsLogger

# ExporterAsLogger is NOT YET IMPLEMENTED

ExporterAsLogger wraps an Open Telemetry SpanExporter so that it can be used
as a xopbase.Logger. This bypasses the TracerProvider, Tracer, and Span APIs.
*/
package xopotel