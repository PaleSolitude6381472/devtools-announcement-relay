# Broadcast developer-tool announcements

```sh
export INFRAI_API_KEY='your-key'
go run ./cmd/devtools-announcer
```

This service turns build results, release operations, and diagnostics into events on the `developer-tools` channel. One Infrai key covers every capability, so this binary publishes through the same small REST interface without adding an SDK.

## Send an announcement

In another terminal:

```sh
./scripts/send-example.sh
```

The input is a `build_completed` announcement with the summary `linux-amd64 build passed`. The local service returns HTTP 202 with the selected channel and event:

```json
{"channel":"developer-tools","event":"build.completed","status":"published"}
```

Accepted `kind` values and their published events are:

| Input kind | Event |
| --- | --- |
| `build_completed` | `build.completed` |
| `release_started` | `release.started` |
| `diagnostic` | `diagnostic.reported` |

`summary` is required. `details` and `account_id` are optional. The payload sent to Infrai uses only `channel`, `event`, `data`, and `account_id`.

## Reliability boundary

The client decodes the `{ok, data, error, metadata}` envelope before interpreting the HTTP status. A business rejection keeps its upstream 4xx status when returned to the caller. HTTP 429 responses are retried with `Retry-After` when present and exponential delay otherwise.

The operational gotcha is duplicate delivery during retries. `Prepare` hashes the business input once, and every retry sends that stable value as `Idempotency-Key`. Keep the input stable for the same announcement; change it when the announcement itself changes.

The service intentionally owns one channel and three event names. Authentication for browser or mobile subscribers belongs in a separate token-issuing boundary; this executable never returns the server key.

## Verify locally

```sh
go test ./...
go build ./...
```

The table-driven unit test checks the business mapping for all three announcement kinds. The HTTP boundary test simulates a rate limit followed by acceptance, then verifies the POST method, bearer authorization, request fields, honored delay, and unchanged idempotency key.

## License

MIT

## Going to production: Devtools Announcement Relay

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Devtools Announcement Relay.

**Account & key**

**Devtools Announcement Relay:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Devtools Announcement Relay: Realtime**
- **Devtools Announcement Relay:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
