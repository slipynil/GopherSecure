# AWG Service HTTP API

This document describes all HTTP endpoints provided by the AWG (AmneziaWireGuard) service.

## Base URL

```
http://<HTTP_ENDPOINT>
```

## Response Format

All responses are in JSON format with the following structure:

```json
{
  "data": {},
  "error": ""
}
```

- **data**: Response payload (omitted if empty or on error)
- **error**: Error message (omitted if no error)

---

## Endpoints

### 1. Add Peer

**Create a new WireGuard peer and generate its configuration.**

#### Request

- **Method**: `POST`
- **Path**: `/peers`
- **Content-Type**: `application/json`

#### Request Body

```json
{
  "id": 123,
  "virtual_endpoint": "10.0.0.1",
  "dns": "8.8.8.8"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | integer | Yes | Unique peer identifier |
| `virtual_endpoint` | string | Yes | Virtual IP address for the peer (e.g., "10.0.0.1") |
| `dns` | string | No | DNS server address for the peer (e.g., "8.8.8.8") |

#### Response

**Success (201 Created)**

```json
{
  "data": {
    "public_key": "WKNwjBYSFXX6NvLhc/OaC1Vxz3DxShF2C1Bz4dE5+0w=",
    "preshared_key": "gI+VqaLCzN9P5K8dR2E3L0M7N2E1D8Q5T4U9X8A7V6="
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `data.public_key` | string | Generated public key for the peer |
| `data.preshared_key` | string | Generated preshared key for enhanced security |

**Error Cases**

- **400 Bad Request** - Missing `id` or `virtual_endpoint`
  ```json
  {
    "error": "id and virtual endpoint are required"
  }
  ```

- **400 Bad Request** - Invalid JSON
  ```json
  {
    "error": "invalid character 'x' looking for beginning of value"
  }
  ```

- **500 Internal Server Error** - AWG service error
  ```json
  {
    "error": "failed to add peer: <error details>"
  }
  ```

#### Example

```bash
curl -X POST http://localhost:8080/peers \
  -H "Content-Type: application/json" \
  -d '{
    "id": 123,
    "virtual_endpoint": "10.0.0.1",
    "dns": "8.8.8.8"
  }'
```

---

### 2. Delete Peer

**Remove an existing WireGuard peer.**

#### Request

- **Method**: `DELETE`
- **Path**: `/peers`
- **Content-Type**: `application/json`

#### Request Body

```json
{
  "public_key": "WKNwjBYSFXX6NvLhc/OaC1Vxz3DxShF2C1Bz4dE5+0w="
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `public_key` | string | Yes | Public key of the peer to delete |

#### Response

**Success (200 OK)**

```json
{}
```

**Error Cases**

- **400 Bad Request** - Invalid JSON
  ```json
  {
    "error": "invalid character 'x' looking for beginning of value"
  }
  ```

- **500 Internal Server Error** - AWG service error
  ```json
  {
    "error": "failed to delete peer: <error details>"
  }
  ```

#### Example

```bash
curl -X DELETE http://localhost:8080/peers \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "WKNwjBYSFXX6NvLhc/OaC1Vxz3DxShF2C1Bz4dE5+0w="
  }'
```

---

### 3. Get Peer Config

**Retrieve the WireGuard configuration file for a peer.**

#### Request

- **Method**: `GET`
- **Path**: `/peers/{id}/config`

#### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Peer identifier (used to locate config file) |

#### Response

**Success (200 OK)**

Returns the raw configuration file content with `Content-Type: text/plain` or appropriate file type.

Example config file:
```ini
[Interface]
PrivateKey = SGoVrRdrYSVX8N2iJQQn2pFf8Nq2C3E5D1K9L5M7N=
Address = 10.0.0.1/32
DNS = 8.8.8.8

[Peer]
PublicKey = WKNwjBYSFXX6NvLhc/OaC1Vxz3DxShF2C1Bz4dE5+0w=
AllowedIPs = 0.0.0.0/0
Endpoint = vpn.example.com:51820
```

**Error Cases**

- **404 Not Found** - Configuration file not found
  ```json
  {
    "error": "file not found"
  }
  ```

#### Example

```bash
curl -X GET http://localhost:8080/peers/123/config
```

Save to file:
```bash
curl -X GET http://localhost:8080/peers/123/config -o peer_123.conf
```

---

## Data Structures

### Request Types

#### AddPeer Request
```go
type Request struct {
	DNS             string `json:"dns,omitempty"`
	VirtualEndpoint string `json:"virtual_endpoint"`
	ID              int64  `json:"id"`
}
```

#### DeletePeer Request
```go
type DelRequest struct {
	PublicKey string `json:"public_key"`
}
```

### Response Types

#### Standard Response Wrapper
```go
type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
```

#### Peer Response (AddPeer)
```go
{
	"public_key": string,
	"preshared_key": string
}
```

---

## Error Handling

All errors are returned with appropriate HTTP status codes:

| Status | Meaning | When Used |
|--------|---------|-----------|
| 200 OK | Success | DeletePeer succeeds |
| 201 Created | Resource created | AddPeer succeeds |
| 400 Bad Request | Invalid request | Missing required fields, invalid JSON |
| 404 Not Found | Resource not found | Config file doesn't exist |
| 500 Internal Server Error | Server error | AWG service failure, file system error |

---

## Config File Location

Peer configuration files are stored at:
```
/etc/amnezia/amneziawg/configs/{id}.conf
```

Where `{id}` is the peer ID from the AddPeer request converted to string format.

---

## Notes

- JSON decode errors return **400 Bad Request** instead of silently ignoring invalid input
- All peer operations are thread-safe (uses Repository lock)
- Config files must exist before they can be retrieved
- Virtual endpoint must be a valid IP address format
