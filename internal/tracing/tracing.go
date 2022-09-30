package tracing

import "go.opentelemetry.io/otel"

var Tracer = otel.Tracer("artforintrovert-test")
