# 📡 Printly REST API Documentation

**Author**: Kimba SABI N'GOYE  
**Date**: June 19, 2025  
**Reviewers**: TBD  

---

This document outlines the full REST API specification for the Printly platform, including endpoints for:

* [USER-FACING ENDPOINTS](#user-facing-endpoints) (anonymous or authenticated)
* [Print Center Managers](#manager-facing-endpoints)
* [Platform Administrators](#admin-facing-endpoints)
* [System Tasks and Webhooks](#internal-tasks--webhooks)

---

## 🔐 Authentication

* **Authentication Provider**: Firebase Authentication
* **Clients**: Use Firebase SDK to log in (anonymous or email/password)
* **Backend**: Verifies bearer token using Firebase Admin SDK
* **Token Handling**: Pass `Authorization: Bearer <token>` in headers

---

## USER-FACING ENDPOINTS

### `GET /centers`

List available print centers near a location.

Response

```json
[
  {
    "id": "center123",
    "name": "Alpha Print Center",
    "email": "contact@alpha.com",
    "services": [
      {
        "type": "color",
        "paper_size": "A4",
        "price": 100
      }
    ],
    "working_hours": [
      {
        "day": "Monday",
        "intervals": [
          { "start": "08:00", "end": "18:00" }
        ]
      }
    ],
    "addresses": [
      {
        "label": "Main",
        "street": "12 Ave Kennedy",
        "city": "Cotonou",
        "lat": 6.45,
        "lng": 2.35
      }
    ]
  }
]
```

---

### `GET /centers/:id`

Get detailed info for a center.

Response

```json
{
  "id": "center123",
  "name": "Alpha Print Center",
  "email": "contact@alpha.com",
  "services": [
    {
      "type": "color",
      "paper_size": "A4",
      "price": 100
    }
  ],
  "working_hours": [
    {
      "day": "Monday",
      "intervals": [
        { "start": "08:00", "end": "18:00" }
      ]
    }
  ],
  "addresses": [
    {
      "label": "Main",
      "street": "12 Ave Kennedy",
      "city": "Cotonou",
      "lat": 6.45,
      "lng": 2.35
    }
  ]
}
```

---

### `POST /upload`

Upload a document for printing.

This endpoint returns a time-limited `upload_url` to enable **secure, efficient, and scalable direct file uploads** to cloud storage (e.g., Google Cloud Storage or AWS S3) without routing the file through the backend server.

#### 📦 Upload Flow

1. **Client calls** `POST /upload` with:

   ```json
   {
     "center_id": "center123",
     "file_name": "cv.pdf",
     "mime_type": "application/pdf"
   }
   ```

2. **Backend actions**:

   * Creates a new `Order` with a unique pickup `code`
   * Generates a **signed upload URL** using GCS SDK
   * Returns the upload URL and order metadata to the client

  ```json
  {
    "upload_url": "...",
    "expires_at": "..."
  }
  ```

1. **Client uploads** the document directly to the returned `upload_url` via HTTP PUT or POST

2. **Cloud triggers** notify backend when upload is complete

#### 🛡️ Security Considerations

* `upload_url` is **time-limited** (e.g., 10 minutes)
* It only allows upload to **one file location**
* It **does not expose storage credentials**
* The file is only accessible to the print center after upload and payment

Request

```json
{
  "center_id": "center123",
  "file_name": "cv.pdf",
  "mime_type": "application/pdf"
}
```

Response

```json
{
  "upload_url": "https://storage.googleapis.com/printly/documents/abc123?signature=xyz...",
  "code": "X9A4C2"
}
```

---

### `POST /pay`

Start payment process.

Request

```json
{
  "order_id": "order456",
  "method": "MOBILE_MONEY",
  "provider": "MTN",
  "phone": "+22991234567"
}
```

Response

```json
{
  "payment_url": "https://paygateway.com/session/xyz",
  "status": "PENDING"
}
```

---

### `POST /schedule`

Schedule pickup time and select print mode.

Request

```json
{
  "order_id": "order456",
  "pickup_time": "2025-06-25T10:30:00Z",
  "print_mode": "PRINT_UPON_ARRIVAL"
}
```

Response

```json
{
  "status": "SCHEDULED"
}
```

---

### `GET /status/:code`

Get order status by pickup code.

Response

```json
{
  "order_id": "order456",
  "status": "AWAITING_USER",
  "print_mode": "PRINT_UPON_ARRIVAL"
}
```

---

### `GET /order/:code/receipt`

View receipt of the order.

Response

```json
{
  "order_id": "order456",
  "amount": 300,
  "paid": true,
  "printed": true,
  "pickup_time": "2025-06-25T10:30:00Z"
}
```

---

## MANAGER-FACING ENDPOINTS

### `POST /register/center`

Register a new print center.

Request

```json
{
  "name": "Alpha Print Center",
  "location": { "lat": 6.45, "lng": 2.35 },
  "contact_email": "owner@alpha.com",
  "services": ["color", "A4"],
  "working_hours": [
    {
      "day": "Monday",
      "intervals": [{ "start": "08:00", "end": "18:00" }]
    }
  ]
}
```

---

### `GET /dashboard/orders`

List orders for the manager's center.

Response

```json
[
  {
    "order_id": "order456",
    "status": "READY_TO_PRINT",
    "pickup_time": "2025-06-25T10:30:00Z"
  }
]
```

---

### `POST /order/:code/verify`

Verify user code to trigger printing.

Response

```json
{
  "authorized": true,
  "message": "Valid code"
}
```

---

### `POST /order/:id/print`

Print a confirmed order.

Response

```json
{
  "status": "PRINTING"
}
```

---

### `PATCH /order/:id/status`

Manually update order status.

Request

```json
{
  "status": "CANCELLED"
}
```

Response

```json
{ "updated": true }
```

---

### `POST /order/:id/cancel`

Cancel an order (manager initiated).

---

### `GET /dashboard/stats`

View performance metrics.

Response

```json
{
  "total_orders": 100,
  "printed": 90,
  "failed": 3,
  "revenue": 35000
}
```

---

## ADMIN-FACING ENDPOINTS

### `GET /admin/centers/pending`

List centers pending approval.

Response

```json
[
  {
    "id": "center123",
    "name": "New Print Center"
  }
]
```

---

### `POST /admin/centers/:id/approve`

Approve a print center.

Response

```json
{ "approved": true }
```

---

### `DELETE /admin/centers/:id`

Delete a print center.

---

### `GET /admin/users`

List all users.

---

## INTERNAL TASKS & WEBHOOKS

### `POST /webhooks/payment`

Handle payment status from gateway.

---

### `POST /tasks/order/cleanup`

Delete printed/expired documents.

---

### `POST /tasks/order/timeout`

Mark overdue orders as cancelled.
