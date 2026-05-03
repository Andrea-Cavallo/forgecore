package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	roleUser           = "user"
	roleAdmin          = "admin"
	roleOwner          = "owner"
	roleBillingManager = "billing-manager"
	roleReadOnly       = "read-only"
)

type routePermission struct {
	Method string
	Prefix string
	Roles  []string
}

var protectedRoutePermissions = []routePermission{
	{Method: http.MethodGet, Prefix: "/v1/payments", Roles: []string{roleUser, roleBillingManager, roleAdmin, roleOwner}},
	{Method: http.MethodPost, Prefix: "/v1/payments", Roles: []string{roleBillingManager, roleAdmin, roleOwner}},
	{Method: http.MethodPost, Prefix: "/v1/payments/", Roles: []string{roleBillingManager, roleAdmin, roleOwner}},
	{Method: http.MethodGet, Prefix: "/v1/notifications", Roles: []string{roleUser, roleAdmin, roleOwner}},
	{Method: http.MethodPost, Prefix: "/v1/notifications", Roles: []string{roleAdmin, roleOwner}},
	{Method: "", Prefix: "/v1/admin", Roles: []string{roleAdmin, roleOwner}},
	{Method: http.MethodGet, Prefix: "/v1/audit", Roles: []string{roleAdmin, roleOwner, roleReadOnly}},
	{Method: http.MethodPost, Prefix: "/v1/permissions/check", Roles: []string{roleUser, roleBillingManager, roleAdmin, roleOwner, roleReadOnly}},
	{Method: "", Prefix: "/v1/permissions", Roles: []string{roleAdmin, roleOwner}},
	{Method: http.MethodGet, Prefix: "/v1/config", Roles: []string{roleAdmin, roleOwner, roleReadOnly}},
	{Method: "", Prefix: "/v1/config", Roles: []string{roleAdmin, roleOwner}},
	{Method: "", Prefix: "/v1/webhooks", Roles: []string{roleAdmin, roleOwner}},
	{Method: http.MethodGet, Prefix: "/v1/storage", Roles: []string{roleUser, roleAdmin, roleOwner}},
	{Method: http.MethodPost, Prefix: "/v1/storage", Roles: []string{roleUser, roleAdmin, roleOwner}},
	{Method: http.MethodPost, Prefix: "/v1/subscriptions", Roles: []string{roleUser, roleBillingManager, roleAdmin, roleOwner}},
	{Method: http.MethodDelete, Prefix: "/v1/subscriptions", Roles: []string{roleBillingManager, roleAdmin, roleOwner}},
}

func RBACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		permission, ok := matchRoutePermission(r.Method, r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		if hasAnyRole(parseRoles(r.Header.Get("X-User-Roles")), permission.Roles) {
			next.ServeHTTP(w, r)
			return
		}
		writeForbidden(w, r)
	})
}

func matchRoutePermission(method string, path string) (routePermission, bool) {
	for _, permission := range protectedRoutePermissions {
		if permission.Method != "" && permission.Method != method {
			continue
		}
		if strings.HasPrefix(path, permission.Prefix) {
			return permission, true
		}
	}
	return routePermission{}, false
}

func parseRoles(raw string) map[string]struct{} {
	roles := map[string]struct{}{}
	for _, role := range strings.Split(raw, ",") {
		role = strings.TrimSpace(role)
		if role != "" {
			roles[role] = struct{}{}
		}
	}
	return roles
}

func hasAnyRole(actual map[string]struct{}, required []string) bool {
	for _, role := range required {
		if _, ok := actual[role]; ok {
			return true
		}
	}
	return false
}

func writeForbidden(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Code:      "forbidden",
		Message:   "permessi insufficienti",
		RequestID: r.Header.Get(HeaderRequestID),
	})
}
