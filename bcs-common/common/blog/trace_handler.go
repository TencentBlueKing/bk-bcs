package blog

import (
	"context"
	"net/http"
	"strings"
)

type key int

const (
	tracerLogHandlerID key = 32702 // random key
	realIPValueID      key = 16221
)

// Handler wrap a trace handler outer the original http.Handler
func Handler(name string, handler http.Handler) http.Handler {
	return http.HandlerFunc(HandleFunc(name, handler.ServeHTTP))
}

// HandleFunc wrap a trace handle func outer the original http handle func
func HandleFunc(name string, handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var tracer Trace
		if id := r.Header.Get("x-request-id"); len(id) > 0 {
			tracer = WithID(name, id)
		} else {
			tracer = New(name)
		}
		w.Header().Set("x-request-id", tracer.ID())

		lastRoute, ip := func(r *http.Request) (string, string) {
			lastRoute := strings.Split(r.RemoteAddr, ":")[0]
			if ip, exists := r.Header["X-Real-IP"]; exists && len(ip) > 0 {
				return lastRoute, ip[0]
			}
			if ips, exists := r.Header["X-Forwarded-For"]; exists && len(ips) > 0 {
				return lastRoute, ips[0]
			}
			return lastRoute, lastRoute
		}(r)

		tracer.Infof("event=[request-in] remote=[%s] route=[%s] method=[%s] url=[%s]", ip, lastRoute, r.Method, r.URL.String())
		defer tracer.Info("event=[request-out]")

		ctx := context.WithValue(r.Context(), tracerLogHandlerID, tracer)
		ctx = context.WithValue(ctx, realIPValueID, ip)

		handler(w, r.WithContext(ctx))
	}
}

// GetTraceFromRequest get the Trace var from the req context, if there is no such a trace utility, return nil
func GetTraceFromRequest(r *http.Request) Trace {
	return GetTraceFromContext(r.Context())
}

// GetTraceFromContext get the Trace var from the context, if there is no such a trace utility, return nil
func GetTraceFromContext(ctx context.Context) Trace {
	if tracer, ok := ctx.Value(tracerLogHandlerID).(Trace); ok {
		return tracer
	}
	return New("default-trace")
}

// GetRealIPFromContext get the remote endpoint from request, if not found, return an empty string
func GetRealIPFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(realIPValueID).(string); ok {
		return ip
	}
	return ""
}

// WithTraceForContext will return a new context wrapped a trace handler around the original ctx
func WithTraceForContext(ctx context.Context, traceName string, traceID ...string) context.Context {
	return context.WithValue(ctx, tracerLogHandlerID, New(traceName, traceID...))
}

// WithTraceForContext2 will return a new context wrapped a trace handler around the original ctx
func WithTraceForContext2(ctx context.Context, tracer Trace) context.Context {
	if tracer == nil {
		return ctx
	}
	return context.WithValue(ctx, tracerLogHandlerID, tracer)
}
