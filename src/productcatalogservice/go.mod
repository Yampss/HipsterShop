module github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice

go 1.25.0

toolchain go1.26.1

require (
	cloud.google.com/go/profiler v0.4.3
	github.com/golang/protobuf v1.5.4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.4
	go.mongodb.org/mongo-driver/v2 v2.2.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.67.0
	go.opentelemetry.io/otel v1.42.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.42.0
	go.opentelemetry.io/otel/sdk v1.42.0
	google.golang.org/grpc v1.79.2
	google.golang.org/protobuf v1.36.11
)
