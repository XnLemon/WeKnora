package middleware

import (
	"net/http"
	"strings"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	secutils "github.com/Tencent/WeKnora/internal/utils"
	"github.com/gin-gonic/gin"
)

// ModelHeaderForwarding captures an operator-approved subset of inbound
// request headers and stores them on the request context for downstream model
// calls. It intentionally does not forward anything unless explicitly enabled.
func ModelHeaderForwarding(cfg *config.ModelHeaderForwardingConfig) gin.HandlerFunc {
	allowed := headerNameSet(nil)
	reservedAllowed := headerNameSet(nil)
	enabled := cfg != nil && cfg.Enabled
	if enabled {
		allowed = headerNameSet(cfg.Allow)
		reservedAllowed = headerNameSet(cfg.ReservedAllow)
	}

	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}
		headers := collectModelForwardHeaders(c.Request.Header, allowed, reservedAllowed)
		if len(headers) > 0 {
			ctx := types.WithModelForwardHeaders(c.Request.Context(), headers)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

func collectModelForwardHeaders(h http.Header, allowed, reservedAllowed map[string]struct{}) map[string]string {
	if len(allowed) == 0 && len(reservedAllowed) == 0 {
		return nil
	}
	out := make(map[string]string)
	for rawName, values := range h {
		name := http.CanonicalHeaderKey(strings.TrimSpace(rawName))
		if name == "" || len(values) == 0 {
			continue
		}
		canonical := strings.ToLower(name)
		_, allowNormal := allowed[canonical]
		_, allowReserved := reservedAllowed[canonical]
		if secutils.IsReservedHeader(name) {
			if !allowReserved {
				continue
			}
		} else if !allowNormal {
			continue
		}
		value := strings.TrimSpace(values[0])
		if value == "" {
			continue
		}
		out[name] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func headerNameSet(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		key := strings.ToLower(strings.TrimSpace(name))
		if key != "" {
			out[key] = struct{}{}
		}
	}
	return out
}
