package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupOtelSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}
	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// propagators are used to propagate the trace context and baggage across the different services.
	initPropagators()

	// initialize the gRPC connection
	conn, err := initConn()
	if err != nil {
		return nil, err
	}

	// tracer provider is used to create and manage the tracers.
	tp, err := newTracerProvider(ctx, conn)
	if err != nil {
		handleErr(err)
	}
	otel.SetTracerProvider(tp)
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)

	// metric provider is used to create and manage the metrics.
	mp, err := newMetricProvider(ctx, conn)
	if err != nil {
		handleErr(err)
	}
	otel.SetMeterProvider(mp)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)

	// logger provider is used to create and manage the loggers.
	lp, err := newLoggerProvider(ctx, conn)
	if err != nil {
		handleErr(err)
	}
	global.SetLoggerProvider(lp)
	shutdownFuncs = append(shutdownFuncs, lp.Shutdown)

	return shutdown, nil
}

func initPropagators() {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)
}

// Initialize a gRPC connection to be used by both the tracer and meter
// providers.
func initConn() (*grpc.ClientConn, error) {
	endpoint := viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:4317"
	}

	fmt.Println("connecting to endpoint: ", endpoint)

	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `127.0.0.1:4317` with your endpoint.
	conn, err := grpc.NewClient(endpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

func newTracerProvider(ctx context.Context, conn *grpc.ClientConn) (*sdktrace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service name and version
	res, err := newResource()
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	return tp, nil
}

func newMetricProvider(ctx context.Context, conn *grpc.ClientConn) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service name and version
	res, err := newResource()
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	return mp, nil
}

func newLoggerProvider(ctx context.Context, conn *grpc.ClientConn) (*sdklog.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithGRPCConn(conn),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service name and version
	res, err := newResource()
	if err != nil {
		return nil, err
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(logExporter),
		),
		sdklog.WithResource(res),
	)

	return loggerProvider, nil
}

// newResource creates a new OpenTelemetry resource with service name and version
func newResource() (*resource.Resource, error) {
	return resource.New(
		context.Background(),
		resource.WithAttributes(
			// Service name - this will replace "unknown_service:main"
			semconv.ServiceNameKey.String("go_starter-api"),
			// Service version
			semconv.ServiceVersionKey.String("1.0.0"),
			// Service namespace (optional)
			semconv.ServiceNamespaceKey.String("knullsoft"),
			// Additional attributes
			semconv.DeploymentEnvironmentKey.String("development"),
		),
	)
}
