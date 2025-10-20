# Webhook Delivery Method

The webhook delivery method allows alarms to send HTTP POST requests to external services when weather conditions trigger an alarm. This enables integration with monitoring systems, notification services, automation platforms, and custom applications.

## Configuration

Webhooks are configured in the alarm JSON file as a channel with type `"webhook"`:

```json
{
 "name": "High Temperature Alert",
 "condition": "temperature > 85F",
 "channels": [
 {
 "type": "webhook",
 "webhook": {
 "url": "https://api.example.com/alerts",
 "method": "POST",
 "headers": {
 "Authorization": "Bearer your-token",
 "Content-Type": "application/json",
 "X-Source": "tempest-weather"
 },
 "body": "{\"alarm\":\"{{alarm_name}}\", \"condition\":\"{{alarm_condition}}\", \"temperature\":{{temperature}}, \"station\":\"{{station}}\", \"timestamp\":\"{{timestamp}}\"}",
 "content_type": "application/json"
 }
 }
 ]
}
```

### Configuration Fields

- **`url`** (required): The HTTP endpoint URL to send the webhook to
- **`method`** (optional): HTTP method to use. Defaults to `"POST"`
- **`headers`** (optional): Object containing HTTP headers to include in the request
- **`body`** (required): Template string for the request body. Supports all alarm and sensor variables
- **`content_type`** (optional): Content-Type header value. Defaults to `"application/json"`

## Template Variables

The webhook body supports all standard template variables:

### Alarm Information
- `{{alarm_name}}` - Name of the triggered alarm
- `{{alarm_description}}` - Description of the alarm
- `{{alarm_condition}}` - The condition that triggered the alarm
- `{{alarm_tags}}` - Comma-separated list of alarm tags

### Station Information
- `{{station}}` - Station name
- `{{timestamp}}` - Current timestamp (ISO format)

### Sensor Values
- `{{temperature}}` - Temperature in Celsius
- `{{temperature_f}}` - Temperature in Fahrenheit
- `{{humidity}}` - Relative humidity percentage
- `{{pressure}}` - Station pressure in millibars
- `{{wind_speed}}` - Wind speed in m/s
- `{{wind_gust}}` - Wind gust in m/s
- `{{wind_direction}}` - Wind direction in degrees
- `{{lux}}` - Illuminance in lux
- `{{uv}}` - UV index
- `{{rain_rate}}` - Rain rate in mm/hour
- `{{rain_daily}}` - Daily rain accumulation in mm
- `{{lightning_count}}` - Lightning strike count
- `{{lightning_distance}}` - Distance to last lightning strike in km

### Application Information
- `{{app_info}}` - Application version and uptime information
- `{{alarm_info}}` - Formatted alarm information block
- `{{sensor_info}}` - Formatted sensor readings block

## Testing Webhooks

Use the `--test-webhook` flag to test webhook delivery:

```bash
tempest-homekit-go --alarms @alarms.json --test-webhook https://api.example.com/webhook
```

This will send a test webhook with sample data and exit.

## Example Go Server

Here's a simple Go server that can receive and log webhook notifications:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type WebhookPayload struct {
	Alarm struct {
 Name string `json:"name"`
 Description string `json:"description"`
 Condition string `json:"condition"`
	} `json:"alarm"`
	Station string `json:"station"`
	Timestamp string `json:"timestamp"`
	Sensors map[string]interface{} `json:"sensors"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
 http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
 return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
 log.Printf("Error decoding webhook payload: %v", err)
 http.Error(w, "Bad request", http.StatusBadRequest)
 return
	}

	// Log the webhook
	fmt.Printf("[%s] Webhook received from station: %s\n",
 time.Now().Format("2006-01-02 15:04:05"), payload.Station)
	fmt.Printf(" Alarm: %s - %s\n", payload.Alarm.Name, payload.Alarm.Description)
	fmt.Printf(" Condition: %s\n", payload.Alarm.Condition)
	fmt.Printf(" Temperature: %.1f°C\n", payload.Sensors["temperature_c"])
	fmt.Printf(" Humidity: %.0f%%\n", payload.Sensors["humidity"])
	fmt.Println()

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
 "status": "received",
 "alarm": payload.Alarm.Name,
 "station": payload.Station,
	})
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)

 fmt.Println("Webhook test server listening on :8080")
 fmt.Println("Send test webhooks to: http://localhost:8080/webhook")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Running the Example Server

1. Save the code above as `webhook_server.go`
2. Run: `go run webhook_server.go`
3. Test with: `tempest-homekit-go --alarms @alarms.json --test-webhook http://localhost:8080/webhook`

## Integration Examples

### Slack Webhook

```json
{
 "type": "webhook",
 "webhook": {
 "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
 "method": "POST",
 "headers": {
 "Content-Type": "application/json"
 },
 "body": "{\"text\":\" Weather Alert: {{alarm_name}}\\nStation: {{station}}\\nTemperature: {{temperature_f}}°F\\nCondition: {{alarm_condition}}\\nTime: {{timestamp}}\"}"
 }
}
```

### Discord Webhook

```json
{
 "type": "webhook",
 "webhook": {
 "url": "https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK",
 "method": "POST",
 "headers": {
 "Content-Type": "application/json"
 },
 "body": "{\"content\":\" **Weather Alert**\\n**Alarm:** {{alarm_name}}\\n**Station:** {{station}}\\n**Temperature:** {{temperature_f}}°F\\n**Condition:** {{alarm_condition}}\\n**Time:** {{timestamp}}\"}"
 }
}
```

### Generic REST API

```json
{
 "type": "webhook",
 "webhook": {
 "url": "https://api.monitoring-service.com/alerts",
 "method": "POST",
 "headers": {
 "Authorization": "Bearer your-api-token",
 "Content-Type": "application/json"
 },
 "body": "{{sensor_info}}"
 }
}
```

## Error Handling

- Webhooks that return HTTP status codes 400-599 are considered failed
- Failed webhooks are logged as errors but don't prevent other delivery methods from executing
- Timeouts are set to 10 seconds
- Network errors are logged with details

## Security Considerations

- Use HTTPS URLs for production webhooks
- Include authentication headers (Bearer tokens, API keys) when required
- Validate webhook payloads on the receiving end
- Consider rate limiting to prevent abuse
- Log webhook delivery for audit purposes