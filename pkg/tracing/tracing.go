package tracing

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/prometheus/common/version"
	jaegerprop "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/samplers/jaegerremote"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// InitTraceExporter creates a new OTLP trace exporter and configures it as the
// global trace provider.
//
// Use environment variables to configure the exporter, such as
// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT.
func InitTraceExporter(
	ctx context.Context,
	serviceName string,
	batchOptions ...sdktrace.BatchSpanProcessorOption) (closer func(context.Context) error, err error) {
	var exporter sdktrace.SpanExporter

	exporter, err = otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init OTLP exporter: %w", err)
	}

	res, err := traceResource(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	sampler := traceSampler(serviceName)

	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exporter, batchOptions...),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	// set the global tracer provider
	otel.SetTracerProvider(tp)

	// configure propagation for both W3C and Jaeger
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		jaegerprop.Jaeger{},
	))

	return exporter.Shutdown, nil
}

type otelErrHandler func(err error)

func (o otelErrHandler) Handle(err error) {
	o(err)
}

func traceSampler(serviceName string) sdktrace.Sampler {
	// for now, just support the JAEGER_SAMPLER_MANAGER_HOST_PORT env var
	// to configure the remote sampler
	samplerURL := "http://localhost:5778/sampling"
	if v := os.Getenv("JAEGER_SAMPLER_MANAGER_HOST_PORT"); v != "" {
		samplerURL = v
	}

	return jaegerremote.New(
		serviceName,
		jaegerremote.WithSamplingServerURL(samplerURL),
		jaegerremote.WithSamplingRefreshInterval(10*time.Second),
		jaegerremote.WithInitialSampler(sdktrace.AlwaysSample()),
	)
}

func traceResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	module := "unknown"
	if bi, ok := debug.ReadBuildInfo(); ok {
		module = bi.Main.Path
	}

	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcessPID(),
		resource.WithProcessExecutableName(),
		resource.WithProcessExecutablePath(),
		resource.WithProcessOwner(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeDescription(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version.Version),
			attribute.String("service.revision", version.Revision),
			attribute.String("module.path", module),
		),
	)
}
