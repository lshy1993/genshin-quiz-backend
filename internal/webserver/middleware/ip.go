package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type contextKey string

const RealIPKey contextKey = "real_ip"
const UserAgentKey contextKey = "user_agent"

// SecureRealIP 返回一个安全的 IP 提取中间件.
func SecureRealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var clientIP string

		// 场景 A：如果你的应用部署在 Cloudflare 后面
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			clientIP = cfIP
		}

		// 场景 B：如果使用了 Nginx 且配置了 proxy_set_header X-Real-IP $remote_addr;
		if clientIP == "" {
			if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
				clientIP = xRealIP
			}
		}

		// 场景 C：处理 X-Forwarded-For (反向代理会将真实客户端 IP 放在最左边)
		if clientIP == "" {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// X-Forwarded-For: client, proxy1, proxy2
				ips := strings.Split(xff, ",")
				if len(ips) > 0 {
					clientIP = strings.TrimSpace(ips[0])
				}
			}
		}

		// 场景 D：无代理直连，直接提取 TCP 连接的 IP
		if clientIP == "" {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				clientIP = ip
			} else {
				clientIP = r.RemoteAddr
			}
		}

		// 将提取到的真实 IP 写入 context
		ctx := context.WithValue(r.Context(), RealIPKey, clientIP)
		// 自动记录浏览器 User-Agent Header
		ctx = context.WithValue(ctx, UserAgentKey, r.UserAgent())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
