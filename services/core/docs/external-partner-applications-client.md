# External Partner Applications API: Client Integration Guide

This guide explains how an external client should call the partner application endpoints.

## Endpoints

Base path:

- /partners/external

Routes:

- GET /partners/external/applications/pending
- PATCH /partners/external/applications/:id/status

## Authentication and Authorization

These endpoints are machine-to-machine only.

They do not accept a user token obtained through the email/password login flow.

Both endpoints require:

- Authorization header with Bearer access token issued by POST /api/auth/v1/oauth/token
- Token must be a machine access token
- Token must be valid in the auth service
- Token must include permission entries:
  - Resource: partners_applications, Action: read (for GET)
  - Resource: partners_applications, Action: write (for PATCH)

Required header format:

- Authorization: Bearer <access_token>

## OAuth Client-Credentials Flow

### 1) Provision a machine client

A platform administrator must create an OAuth client in the auth service and assign the required scopes.

Current admin endpoints:

- GET /api/auth/v1/admin/oauth/clients
- POST /api/auth/v1/admin/oauth/clients
- PUT /api/auth/v1/admin/oauth/clients/:id/status
- POST /api/auth/v1/admin/oauth/clients/:id/rotate-secret

Create-client request example:

```json
{
  "client_id": "partner-sync-prod",
  "name": "Partner Sync Production",
  "description": "Machine client for partner application sync",
  "scopes": ["partners_applications:read", "partners_applications:write"]
}
```

Success response includes:

- client metadata
- client_secret
- scope string

Persist the returned client_secret securely. It is only returned at creation time and on secret rotation.

### 2) Request an access token

Request:

- Method: POST
- URL: /api/auth/v1/oauth/token
- Content-Type: application/x-www-form-urlencoded
- grant_type must be client_credentials

Form fields:

- grant_type=client_credentials
- client_id=<provisioned client id>
- client_secret=<provisioned client secret>

Success response:

```json
{
  "access_token": "<token>",
  "token_type": "Bearer",
  "expires_in": 900,
  "scope": "partners_applications:read partners_applications:write"
}
```

### 3) Validate granted access

Before calling the core endpoints, optionally validate the machine token against:

- GET /api/auth/v1/permissions/machine

Expected result:

- HTTP 200
- permission array containing the required resource/action pairs

If this request returns 401, the token is not a valid machine token.

If this request returns 200 but the expected permission is missing, the client was provisioned without sufficient access.

## Endpoint Details

### 1) Get pending applications

Request:

- Method: GET
- URL: /partners/external/applications/pending
- Optional query: limit
  - limit must be an integer
  - service bounds list size to a safe max

Example:
GET /partners/external/applications/pending?limit=25

Success response:

- HTTP 200
- Body: JSON array of partner applications

Common errors:

- HTTP 400: invalid limit
- HTTP 401: missing or invalid token
- HTTP 403: token valid but missing read permission
- HTTP 500: server/repository failure

### 2) Update application status

Request:

- Method: PATCH
- URL: /partners/external/applications/:id/status
- Body JSON:
  - status: string (required)

Allowed status values:

- pending
- approved
- rejected
- in_review

Example body:
{
"status": "approved"
}

Success response:

- HTTP 200
- Body: updated partner application object

Common errors:

- HTTP 400: invalid payload or invalid status
- HTTP 401: missing or invalid token
- HTTP 403: token valid but missing write permission
- HTTP 404: application id not found
- HTTP 500: server/repository failure

## cURL Examples

### Get pending applications

curl -X GET "https://partners.animalpride.com/api/core/v1/partners/external/applications/pending?limit=25" \
 -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

### Update status

curl -X PATCH "https://partners.animalpride.com/api/core/v1/partners/external/applications/APP_ID/status" \
 -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
 -H "Content-Type: application/json" \
 -d '{"status":"approved"}'

### Request machine access token

curl -X POST "https://partners.animalpride.com/api/auth/v1/oauth/token" \
 -H "Content-Type: application/x-www-form-urlencoded" \
 --data-urlencode "grant_type=client_credentials" \
 --data-urlencode "client_id=partner-sync-prod" \
 --data-urlencode "client_secret=YOUR_CLIENT_SECRET"

### Validate machine token permissions

curl -X GET "https://partners.animalpride.com/api/auth/v1/permissions/machine" \
 -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

## JavaScript Example (fetch)

const baseUrl = "https://partners.animalpride.com";

async function getMachineToken(clientId, clientSecret) {
const body = new URLSearchParams({
grant_type: "client_credentials",
client_id: clientId,
client_secret: clientSecret,
});

const response = await fetch(`${baseUrl}/api/auth/v1/oauth/token`, {
method: "POST",
headers: {
"Content-Type": "application/x-www-form-urlencoded",
},
body,
});

if (!response.ok) {
throw new Error(`Failed to acquire machine token: ${response.status}`);
}

return response.json();
}

async function getPendingApplications(token, limit = 50) {
const response = await fetch(`${baseUrl}/api/core/v1/partners/external/applications/pending?limit=${limit}`, {
method: "GET",
headers: {
Authorization: `Bearer ${token}`,
},
});

if (!response.ok) {
throw new Error(`Failed to fetch pending applications: ${response.status}`);
}

return response.json();
}

async function updateApplicationStatus(token, id, status) {
const response = await fetch(`${baseUrl}/api/core/v1/partners/external/applications/${id}/status`, {
method: "PATCH",
headers: {
Authorization: `Bearer ${token}`,
"Content-Type": "application/json",
},
body: JSON.stringify({ status }),
});

if (!response.ok) {
throw new Error(`Failed to update status: ${response.status}`);
}

return response.json();
}

## Client-side Implementation Checklist

1. Have an administrator provision an OAuth client with partners_applications:read and or partners_applications:write scopes.
2. Store the returned client_id and client_secret in your secret manager.
3. Acquire a machine access token from POST /api/auth/v1/oauth/token using client_credentials.
4. Optionally validate the token against GET /api/auth/v1/permissions/machine before first use.
5. Send Authorization Bearer header on every core request.
6. Restrict status updates in client UI or service logic to allowed values only.
7. Handle 401 by reacquiring a token or alerting on invalid client credentials.
8. Handle 403 by surfacing insufficient client permissions to the operator.
9. Handle 404 by removing stale application ids from the sync queue.
10. Retry only idempotent GET calls on transient network failures.
11. Rotate client secrets periodically and immediately after suspected exposure.

## Suggested QA Test Cases (Client)

1. Request token with valid client credentials: expect 200 and Bearer token response.
2. Request token with invalid client_secret: expect 401.
3. Call GET /api/auth/v1/permissions/machine with valid machine token: expect 200 and permission array.
4. Call GET pending with valid token and read permission: expect 200 and array.
5. Call GET pending with user login token: expect 401.
6. Call GET pending without token: expect 401.
7. Call GET pending with machine token missing read permission: expect 403.
8. Call PATCH with valid token and write permission: expect 200 and updated status.
9. Call PATCH with invalid status value: expect 400.
10. Call PATCH with unknown id: expect 404.
11. Call PATCH with machine token missing write permission: expect 403.
