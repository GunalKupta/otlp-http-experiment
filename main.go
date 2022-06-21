package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric/export"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"math/rand"
	"os"
	"time"
)

var collectorEndpoint = "localhost:8080" // "collector.telemetry.confluent.cloud"

func main() {
	controller := initializeSdk()
	ctx := context.Background()
	defer controller.Stop(ctx)

	meter := global.Meter("kunal/example")
	counter, err := meter.SyncInt64().Counter("example.counter")
	if err != nil {
		panic(err)
	}

	colorKey := attribute.Key("color")
	countryKey := attribute.Key("country")
	colorLabelValues := []string{"red", "blue", "green"}
	countryLabelValues := []string{"usa", "china", "canada", "mexico", "france"}

	fmt.Println("Emitting metrics")

	for {
		time.Sleep(2 * time.Second)
		for i := 0; i < 5; i++ {
			counter.Add(ctx, rand.Int63n(10),
				colorKey.String(colorLabelValues[rand.Intn(len(colorLabelValues))]),
				countryKey.String(countryLabelValues[rand.Intn(len(countryLabelValues))]),
			)
		}
		fmt.Println("Added to counter")
	}

}

// Initializes the OpenTelemetry SDK by building the controller
// which contains configurations for the metrics exporter
func initializeSdk() *controller.Controller {
	ctx := context.Background()
	res, err := resource.New(ctx)
	if err != nil {
		panic(err)
	}

	exporter, exists := os.LookupEnv("METRICS_EXPORTER")

	var metricExporter sdkmetric.Exporter
	if exists && exporter == "http" {
		metricExporter, err = getHttpExporter(ctx)
		fmt.Printf("Using HTTP exporter with endpoint '%s'\n", collectorEndpoint)
	} else {
		metricExporter, err = getLoggingExporter(ctx)
		fmt.Println("Using stdout exporter")
	}
	if err != nil {
		panic(err)
	}

	controller := controller.New(
		processor.NewFactory(
			simple.NewWithInexpensiveDistribution(),
			metricExporter,
		),
		controller.WithResource(res),
		controller.WithExporter(metricExporter),
		controller.WithCollectPeriod(time.Second*5),
	)
	global.SetMeterProvider(controller)
	err = controller.Start(ctx)
	if err != nil {
		panic(err)
	}
	return controller
}

func getLoggingExporter(_ context.Context) (sdkmetric.Exporter, error) {
	return stdoutmetric.New(stdoutmetric.WithWriter(&logWriter{}))
}

func getHttpExporter(ctx context.Context) (sdkmetric.Exporter, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(collectorEndpoint),
		otlpmetrichttp.WithInsecure(),
	}
	metricClient := otlpmetrichttp.NewClient(opts...)
	return otlpmetric.New(ctx, metricClient)
}

type logWriter struct{}

func (l *logWriter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return len(p), nil
}
