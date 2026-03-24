package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aryan/apiproxy/internal/config"
	"github.com/aryan/apiproxy/internal/middleware"
)

func New(route config.Route) (http.Handler, error) {
	target, err := url.Parse(route.Upstream)
	if err != nil {
		return nil, fmt.Errorf("parse upstream for route %q: %w", route.Name, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.URL.Path = rewritePath(req.URL.Path, route.StripPrefix)
		if req.URL.RawPath != "" {
			req.URL.RawPath = rewritePath(req.URL.RawPath, route.StripPrefix)
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		writeProxyError(w, r, err)
	}

	if route.TimeoutMS <= 0 {
		return proxy, nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(route.TimeoutMS)*time.Millisecond)
		defer cancel()
		proxy.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}

func rewritePath(path string, stripPrefix string) string {
	if stripPrefix == "" || !strings.HasPrefix(path, stripPrefix) {
		return path
	}

	rewritten := strings.TrimPrefix(path, stripPrefix)
	if rewritten == "" {
		return "/"
	}
	if rewritten[0] != '/' {
		return "/" + rewritten
	}

	return rewritten
}

func writeProxyError(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := mapProxyError(err)
	requestID := r.Header.Get(middleware.RequestIDHeader)
	if requestID != "" {
		w.Header().Set(middleware.RequestIDHeader, requestID)
		msg = fmt.Sprintf("%s (request_id=%s)", msg, requestID)
	}
	http.Error(w, msg, status)
}

func mapProxyError(err error) (int, string) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "upstream timeout"
	case isTimeout(err):
		return http.StatusGatewayTimeout, "upstream timeout"
	default:
		return http.StatusBadGateway, "upstream request failed"
	}
}

func isTimeout(err error) bool {
	type timeout interface {
		Timeout() bool
	}

	if errors.Is(err, io.EOF) {
		return false
	}

	var te timeout
	return errors.As(err, &te) && te.Timeout()
}
