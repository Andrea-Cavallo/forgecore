# RBAC Endpoint Matrix

This matrix is the verified authorization contract for frontend-facing ForgeCore routes.

| Method | Endpoint prefix | Required roles | Audit required |
| --- | --- | --- | --- |
| `GET` | `/v1/payments` | `user`, `billing-manager`, `admin`, `owner` | no |
| `POST` | `/v1/payments` | `billing-manager`, `admin`, `owner` | yes |
| `POST` | `/v1/payments/` | `billing-manager`, `admin`, `owner` | yes |
| `GET` | `/v1/notifications` | `user`, `admin`, `owner` | no |
| `POST` | `/v1/notifications` | `admin`, `owner` | yes |
| `ANY` | `/v1/admin` | `admin`, `owner` | yes |
| `GET` | `/v1/audit` | `admin`, `owner`, `read-only` | no |
| `POST` | `/v1/permissions/check` | `user`, `billing-manager`, `admin`, `owner`, `read-only` | no |
| `ANY` | `/v1/permissions` | `admin`, `owner` | yes |
| `GET` | `/v1/config` | `admin`, `owner`, `read-only` | no |
| `ANY` | `/v1/config` | `admin`, `owner` | yes |
| `ANY` | `/v1/webhooks` | `admin`, `owner` | yes |
| `GET` | `/v1/storage` | `user`, `admin`, `owner` | no |
| `POST` | `/v1/storage` | `user`, `admin`, `owner` | yes |
| `POST` | `/v1/subscriptions` | `user`, `billing-manager`, `admin`, `owner` | yes |
| `DELETE` | `/v1/subscriptions` | `billing-manager`, `admin`, `owner` | yes |

Public auth and health endpoints bypass RBAC. Every protected route is authenticated first, then checked by `forgecore-gateway/internal/middleware.RBACMiddleware`.
