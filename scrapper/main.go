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
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	wiki := client.NewWiki(sugar)
	broker := broker2.NewRabbitMQ(sugar)
	service := usecase.NewService(broker, wiki)
	cleanup := initTracer(sugar)
	defer cleanup(context.Background())
	service.StartListening()
}

func initTracer(logger *zap.SugaredLogger) func(context.Context) error {

	secureOption := otlptracegrpc.WithInsecure()

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint("localhost:4318"),
		),
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
