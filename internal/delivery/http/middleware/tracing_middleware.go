package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware starts an OpenTelemetry span for incoming requests
func TracingMiddleware(next httprouter.Handle, routeName string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		
		tracer := otel.Tracer("http-server")
		opts := []trace.SpanStartOption{
			trace.WithAttributes(
				semconv.HTTPMethod(r.Method),
				semconv.HTTPTarget(r.URL.Path),
				semconv.HTTPRoute(routeName),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		spanName := r.Method + " " + routeName
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		r = r.WithContext(ctx)

		var rw *responseWriterObserver
		var ok bool
		if rw, ok = w.(*responseWriterObserver); !ok {
			rw = &responseWriterObserver{ResponseWriter: w, status: 200}
		}

		next(rw, r, p)

		span.SetAttributes(semconv.HTTPStatusCode(rw.status))
		if rw.status >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}
