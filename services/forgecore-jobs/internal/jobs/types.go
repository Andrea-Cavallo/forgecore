package jobs

const (
	TypeRetryNotification = "notification:retry"
	TypeCleanupTokens     = "auth:cleanup_tokens"
	TypeDailyReport       = "admin:daily_report"
	TypeProcessWebhooks   = "webhook:process_pending"
)

type RetryNotificationPayload struct {
	NotificationID string `json:"notification_id"`
	TenantID       string `json:"tenant_id"`
}

type CleanupTokensPayload struct {
	TenantID string `json:"tenant_id"`
}

type DailyReportPayload struct {
	TenantID string `json:"tenant_id"`
	Date     string `json:"date"` // YYYY-MM-DD
}

type ProcessWebhooksPayload struct {
	TenantID string `json:"tenant_id"`
}
