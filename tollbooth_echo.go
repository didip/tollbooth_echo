package tollbooth_echo

import (
	"net/http"
	"strings"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/errors"
	"github.com/didip/tollbooth/libstring"
	"github.com/didip/tollbooth/limiter"
	"github.com/labstack/echo"
)

func LimitMiddleware(limiter *limiter.Limiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return echo.HandlerFunc(func(c echo.Context) error {
			httpError := LimitByRequest(limiter, c.Request())
			if httpError != nil {
				return c.String(httpError.StatusCode, httpError.Message)
			}
			return next(c)
		})
	}
}

func LimitHandler(limiter *limiter.Limiter) echo.MiddlewareFunc {
	return LimitMiddleware(limiter)
}

// LimitByRequest builds keys based on http.Request struct,
// loops through all the keys, and check if any one of them returns HTTPError.
func LimitByRequest(limiter *limiter.Limiter, r *http.Request) *errors.HTTPError {
	sliceKeys := BuildKeys(limiter, r)

	// Loop sliceKeys and check if one of them has error.
	for _, keys := range sliceKeys {
		httpError := tollbooth.LimitByKeys(limiter, keys)
		if httpError != nil {
			return httpError
		}
	}

	return nil
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// RemoteIP finds IP Address given http.Request struct.
func RemoteIP(ipLookups []string, r *http.Request) string {
	realIP := r.Header.Get("X-Real-IP")
	forwardedFor := r.Header.Get("X-Forwarded-For")

	for _, lookup := range ipLookups {
		if lookup == "RemoteAddr" {
			return ipAddrFromRemoteAddr(r.RemoteAddr)
		}
		if lookup == "X-Forwarded-For" && forwardedFor != "" {
			// X-Forwarded-For is potentially a list of addresses separated with ","
			parts := strings.Split(forwardedFor, ",")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}
			return parts[0]
		}
		if lookup == "X-Real-IP" && realIP != "" {
			return realIP
		}
	}

	return ""
}

// BuildKeys generates a slice of keys to rate-limit by given limiter and request structs.
func BuildKeys(limiter *limiter.Limiter, r *http.Request) [][]string {
	remoteIP := RemoteIP(limiter.GetIPLookups(), r)
	path := r.URL.Path
	sliceKeys := make([][]string, 0)

	// Don't BuildKeys if remoteIP is blank.
	if remoteIP == "" {
		return sliceKeys
	}

	if limiter.GetMethods() != nil && limiter.GetHeaders() != nil && limiter.GetBasicAuthUsers() != nil {
		// Limit by HTTP methods and HTTP headers+values and Basic Auth credentials.
		if libstring.StringInSlice(limiter.GetMethods(), r.Method) {
			for headerKey, headerValues := range limiter.GetHeaders() {
				if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
					// If header values are empty, rate-limit all request with headerKey.
					username, _, ok := r.BasicAuth()
					if ok && libstring.StringInSlice(limiter.GetBasicAuthUsers(), username) {
						sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey, username})
					}

				} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
					// If header values are not empty, rate-limit all request with headerKey and headerValues.
					for _, headerValue := range headerValues {
						username, _, ok := r.BasicAuth()
						if ok && libstring.StringInSlice(limiter.GetBasicAuthUsers(), username) {
							sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey, headerValue, username})
						}
					}
				}
			}
		}

	} else if limiter.GetMethods() != nil && limiter.GetHeaders() != nil {
		// Limit by HTTP methods and HTTP headers+values.
		if libstring.StringInSlice(limiter.GetMethods(), r.Method) {
			for headerKey, headerValues := range limiter.GetHeaders() {
				if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
					// If header values are empty, rate-limit all request with headerKey.
					sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey})

				} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
					// If header values are not empty, rate-limit all request with headerKey and headerValues.
					for _, headerValue := range headerValues {
						sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey, headerValue})
					}
				}
			}
		}

	} else if limiter.GetMethods() != nil && limiter.GetBasicAuthUsers() != nil {
		// Limit by HTTP methods and Basic Auth credentials.
		if libstring.StringInSlice(limiter.GetMethods(), r.Method) {
			username, _, ok := r.BasicAuth()
			if ok && libstring.StringInSlice(limiter.GetBasicAuthUsers(), username) {
				sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, username})
			}
		}

	} else if limiter.GetMethods() != nil {
		// Limit by HTTP methods.
		if libstring.StringInSlice(limiter.GetMethods(), r.Method) {
			sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method})
		}

	} else if limiter.GetHeaders() != nil {
		// Limit by HTTP headers+values.
		for headerKey, headerValues := range limiter.GetHeaders() {
			if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
				// If header values are empty, rate-limit all request with headerKey.
				sliceKeys = append(sliceKeys, []string{remoteIP, path, headerKey})

			} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
				// If header values are not empty, rate-limit all request with headerKey and headerValues.
				for _, headerValue := range headerValues {
					sliceKeys = append(sliceKeys, []string{remoteIP, path, headerKey, headerValue})
				}
			}
		}

	} else if limiter.GetBasicAuthUsers() != nil {
		// Limit by Basic Auth credentials.
		username, _, ok := r.BasicAuth()
		if ok && libstring.StringInSlice(limiter.GetBasicAuthUsers(), username) {
			sliceKeys = append(sliceKeys, []string{remoteIP, path, username})
		}
	} else {
		// Default: Limit by remoteIP and path.
		sliceKeys = append(sliceKeys, []string{remoteIP, path})
	}

	return sliceKeys
}
