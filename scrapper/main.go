package main

import (
	"context"
	broker2 "github.com/konradloch/distributed-scrapper/scrapper/site/broker"
	"github.com/konradloch/distributed-scrapper/scrapper/site/client"
	"github.com/konradloch/distributed-scrapper/scrapper/site/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	wiki := client.NewWiki(sugar)
	broker := broker2.NewRabbitMQ(sugar)
	service := usecase.NewService(broker, wiki)
	ctx := context.Background()
	cleanup := initTracer(ctx, sugar)
	defer cleanup(ctx)
	service.StartListening()
}

func initTracer(ctx context.Context, logger *zap.SugaredLogger) func(context.Context) error {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(secureOption, otlptracegrpc.WithEndpoint("otel-collector:4317")),
	)

	if err != nil {
		logger.Fatal(err)
	}
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(newResource()),
		),
	)

	return exporter.Shutdown
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("scrapper"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}
