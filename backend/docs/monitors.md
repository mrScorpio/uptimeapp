# Monitors API

Monitor endpoints require an access token in the `Authorization: Bearer <token>` header.

## Create a monitor

`POST /api/v1/monitors`

```json
{
  "url": "https://example.com",
  "interval_seconds": 60
}
```

The URL must use HTTP or HTTPS. The interval must be between 1 and 86,400 seconds. A successful request returns `201 Created` and the saved monitor.

## List monitors

`GET /api/v1/monitors` returns the authenticated user's monitors, newest first. Creating a monitor stores its URL and interval; it does not start website checks.
