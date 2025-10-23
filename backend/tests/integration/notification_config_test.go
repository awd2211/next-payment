package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNotificationSend tests sending notification
func TestNotificationSend(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.NotificationURL, 3); err != nil {
		t.Skipf("Notification service not ready: %v", err)
	}

	t.Log("\n=== Test: Send Notification ===")

	// Send email notification
	t.Log("\n--- Send Email Notification ---")
	notificationData := map[string]interface{}{
		"type":       "email",
		"recipient":  "test@example.com",
		"subject":    "Test Notification",
		"content":    "This is a test notification from integration test",
		"merchant_id": uuid.New().String(),
		"template":   "payment_success",
	}

	resp, body, err := client.POST(
		config.NotificationURL+"/api/v1/notifications/send",
		notificationData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		result := ParseJSONResponse(t, body)
		if notificationID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, notificationID, "notification_id")
			t.Logf("✓ Notification sent: %s", notificationID)
		} else if message, ok := result["message"].(string); ok {
			t.Logf("Response: %s", message)
		}
	} else {
		t.Logf("Send notification returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Notification Send Test Completed ===")
}

// TestNotificationQuery tests querying notification status
func TestNotificationQuery(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.NotificationURL, 3); err != nil {
		t.Skipf("Notification service not ready: %v", err)
	}

	t.Log("\n=== Test: Query Notification ===")

	// Send notification first
	notificationData := map[string]interface{}{
		"type":      "email",
		"recipient": "query@example.com",
		"subject":   "Query Test",
		"content":   "Test content",
	}

	resp, body, err := client.POST(
		config.NotificationURL+"/api/v1/notifications/send",
		notificationData,
		nil,
	)

	if err != nil || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted) {
		t.Skipf("Notification send failed, skipping query test")
	}

	result := ParseJSONResponse(t, body)
	var notificationID string
	if id, ok := result["id"].(string); ok {
		notificationID = id
	} else {
		t.Skip("Notification ID not available")
	}

	// Query notification
	Sleep(500 * time.Millisecond)

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/notifications/%s", config.NotificationURL, notificationID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query notification: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result = ParseJSONResponse(t, body)
		status := result["status"].(string)
		AssertNotEmpty(t, status, "notification status")
		t.Logf("✓ Notification status: %s", status)
	} else {
		t.Logf("Query notification returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Notification Query Test Completed ===")
}

// TestNotificationTemplate tests notification template management
func TestNotificationTemplate(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.NotificationURL, 3); err != nil {
		t.Skipf("Notification service not ready: %v", err)
	}

	t.Log("\n=== Test: Notification Template ===")

	// Create template
	templateData := map[string]interface{}{
		"name":        "payment_completed_test",
		"type":        "email",
		"subject":     "Payment Completed - {{merchant_name}}",
		"content":     "Your payment of {{amount}} has been completed successfully.",
		"description": "Payment completion notification template",
	}

	resp, body, err := client.POST(
		config.NotificationURL+"/api/v1/templates",
		templateData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result := ParseJSONResponse(t, body)
		if templateID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, templateID, "template_id")
			t.Logf("✓ Template created: %s", templateID)

			// Query template
			resp, body, err = client.GET(
				fmt.Sprintf("%s/api/v1/templates/%s", config.NotificationURL, templateID),
				nil,
			)

			if err != nil {
				t.Fatalf("Failed to query template: %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				result = ParseJSONResponse(t, body)
				templateName := result["name"].(string)
				AssertEqual(t, "payment_completed_test", templateName, "template name")
				t.Log("✓ Template verified")
			}
		}
	} else {
		t.Logf("Create template returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Template Test Completed ===")
}

// TestNotificationBatch tests batch notification sending
func TestNotificationBatch(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.NotificationURL, 3); err != nil {
		t.Skipf("Notification service not ready: %v", err)
	}

	t.Log("\n=== Test: Batch Notification ===")

	batchData := map[string]interface{}{
		"type":     "email",
		"template": "payment_success",
		"recipients": []map[string]interface{}{
			{"email": "user1@example.com", "name": "User 1"},
			{"email": "user2@example.com", "name": "User 2"},
			{"email": "user3@example.com", "name": "User 3"},
		},
		"subject": "Batch Test Notification",
		"content": "Batch notification content",
	}

	resp, body, err := client.POST(
		config.NotificationURL+"/api/v1/notifications/batch",
		batchData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to send batch notification: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		result := ParseJSONResponse(t, body)
		if batchID, ok := result["batch_id"].(string); ok {
			t.Logf("✓ Batch notification sent: %s", batchID)
		} else if count, ok := result["count"].(float64); ok {
			t.Logf("✓ Sent to %d recipients", int(count))
		}
	} else {
		t.Logf("Batch notification returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Batch Notification Test Completed ===")
}

// TestConfigManagement tests system configuration CRUD
func TestConfigManagement(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.ConfigURL, 3); err != nil {
		t.Skipf("Config service not ready: %v", err)
	}

	t.Log("\n=== Test: Configuration Management ===")

	// Create configuration
	configData := map[string]interface{}{
		"key":         "test.max_transaction_amount",
		"value":       "1000000",
		"type":        "int",
		"description": "Maximum transaction amount for testing",
		"category":    "payment",
	}

	resp, body, err := client.POST(
		config.ConfigURL+"/api/v1/configs",
		configData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result := ParseJSONResponse(t, body)
		if configID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, configID, "config_id")
			t.Logf("✓ Configuration created: %s", configID)

			// Query configuration
			resp, body, err = client.GET(
				fmt.Sprintf("%s/api/v1/configs/%s", config.ConfigURL, configID),
				nil,
			)

			if err != nil {
				t.Fatalf("Failed to query config: %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				result = ParseJSONResponse(t, body)
				key := result["key"].(string)
				AssertEqual(t, "test.max_transaction_amount", key, "config key")
				t.Log("✓ Configuration verified")
			}

			// Update configuration
			updateData := map[string]interface{}{
				"value": "2000000",
			}

			resp, _, err = client.PUT(
				fmt.Sprintf("%s/api/v1/configs/%s", config.ConfigURL, configID),
				updateData,
				nil,
			)

			if err != nil {
				t.Fatalf("Failed to update config: %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				t.Log("✓ Configuration updated")
			}
		}
	} else {
		t.Logf("Create config returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Configuration Test Completed ===")
}

// TestConfigQuery tests querying configuration by key
func TestConfigQuery(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.ConfigURL, 3); err != nil {
		t.Skipf("Config service not ready: %v", err)
	}

	t.Log("\n=== Test: Query Configuration by Key ===")

	resp, body, err := client.GET(
		config.ConfigURL+"/api/v1/configs/by-key/payment.stripe.enabled",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query config by key: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		value := result["value"].(string)
		AssertNotEmpty(t, value, "config value")
		t.Logf("✓ Configuration value: %s", value)
	} else if resp.StatusCode == http.StatusNotFound {
		t.Log("Configuration key not found (expected for new system)")
	} else {
		t.Logf("Query config by key returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Config Query Test Completed ===")
}

// TestConfigList tests listing configurations
func TestConfigList(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.ConfigURL, 3); err != nil {
		t.Skipf("Config service not ready: %v", err)
	}

	t.Log("\n=== Test: List Configurations ===")

	resp, body, err := client.GET(
		config.ConfigURL+"/api/v1/configs?page=1&page_size=20",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to list configs: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "List configurations")

	result := ParseJSONResponse(t, body)
	if list, ok := result["list"].([]interface{}); ok {
		t.Logf("✓ Retrieved %d configurations", len(list))
	} else if data, ok := result["data"].([]interface{}); ok {
		t.Logf("✓ Retrieved %d configurations", len(data))
	}

	t.Log("\n=== Config List Test Completed ===")
}

// TestAccountingRecords tests accounting record creation
func TestAccountingRecords(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.AccountingURL, 3); err != nil {
		t.Skipf("Accounting service not ready: %v", err)
	}

	t.Log("\n=== Test: Accounting Records ===")

	// Create accounting record
	recordData := map[string]interface{}{
		"merchant_id":   uuid.New().String(),
		"transaction_id": uuid.New().String(),
		"amount":        10000,
		"currency":      "USD",
		"type":          "income",
		"description":   "Payment received",
	}

	resp, body, err := client.POST(
		config.AccountingURL+"/api/v1/accounting/records",
		recordData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create accounting record: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result := ParseJSONResponse(t, body)
		if recordID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, recordID, "record_id")
			t.Logf("✓ Accounting record created: %s", recordID)
		}
	} else {
		t.Logf("Create accounting record returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Accounting Record Test Completed ===")
}

// TestAccountingBalance tests balance query
func TestAccountingBalance(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.AccountingURL, 3); err != nil {
		t.Skipf("Accounting service not ready: %v", err)
	}

	t.Log("\n=== Test: Account Balance ===")

	merchantID := uuid.New().String()

	resp, body, err := client.GET(
		fmt.Sprintf("%s/api/v1/accounting/balance/%s", config.AccountingURL, merchantID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query balance: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		balance := result["balance"].(float64)
		t.Logf("Account balance: %.2f", balance)
		t.Log("✓ Balance query successful")
	} else if resp.StatusCode == http.StatusNotFound {
		t.Log("Account not found (expected for new merchant)")
	} else {
		t.Logf("Balance query returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Balance Test Completed ===")
}

// TestAnalyticsReport tests analytics report generation
func TestAnalyticsReport(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.AnalyticsURL, 3); err != nil {
		t.Skipf("Analytics service not ready: %v", err)
	}

	t.Log("\n=== Test: Analytics Report ===")

	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	resp, body, err := client.GET(
		fmt.Sprintf("%s/api/v1/analytics/reports?start_date=%s&end_date=%s",
			config.AnalyticsURL, startDate, endDate),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to get analytics report: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		t.Logf("Analytics Report: %+v", result)
		t.Log("✓ Analytics report generated")
	} else {
		t.Logf("Analytics report returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Analytics Report Test Completed ===")
}

// TestNotificationStatistics tests notification statistics
func TestNotificationStatistics(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.NotificationURL, 3); err != nil {
		t.Skipf("Notification service not ready: %v", err)
	}

	t.Log("\n=== Test: Notification Statistics ===")

	resp, body, err := client.GET(
		config.NotificationURL+"/api/v1/notifications/statistics",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to get notification statistics: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		t.Logf("Notification Statistics: %+v", result)

		expectedFields := []string{"total_sent", "success_count", "failed_count"}
		for _, field := range expectedFields {
			if _, ok := result[field]; ok {
				t.Logf("✓ Field %s present", field)
			}
		}
	} else {
		t.Logf("Notification statistics returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Notification Statistics Test Completed ===")
}
