# 📡 Printly REST API Documentation

**Author**: Kimba SABI N'GOYE  
**Date**: June 19, 2025  
**Reviewers**: TBD  

---

This document outlines the full REST API specification for the Printly platform, including endpoints for:

* [Users](#users-api)
* [Print Centers](#print-centers-api)
* [Orders](#orders-api)
* [Webhooks](#webhooks)
* [System Tasks](#system-tasks)

---

## 🔐 Authentication

* **Authentication Provider**: Firebase Authentication
* **Clients**: Use Firebase SDK to log in (anonymous or email/password)
* **Backend**: Verifies bearer token using Firebase Admin SDK
* **Token Handling**: Pass `Authorization: Bearer <token>` in headers

---

## Enpoints overview

| **Entity**     | **Method & Path**                     | **Roles Allowed**    | **Description**                                 |
|----------------|----------------------------------------|-----------------------|--------------------------------------------------|
| **Users**      | `POST /users`                          | All                   | Register or sync user                            |
|                | `GET /users`                           | Admin                 | List all users                                   |
|                | `GET /users/:uid`                      | Admin                 | Get user by UID                                  |
|                | `PUT /users/:uid`                      | Admin                 | Update user profile                              |
|                | `DELETE /users/:uid`                   | Admin                 | Delete a user                                    |
|                | `GET /users/me`                        | Authenticated         | Get current user profile                         |
|                | `PATCH /users/me`                      | Authenticated         | Update own user profile                          |
|                | `DELETE /users/me`                     | Authenticated         | Delete own account                               |
| **Print Centers** | `GET /centers`                     | All                   | List all print centers                           |
|                | `POST /centers`                        | Authenticated         | Register a new center                            |
|                | `GET /centers/:id`                     | All                   | Get center details                               |
|                | `PUT /centers/:id`                     | Manager, Admin        | Update center info                               |
|                | `GET /centers/pending`                 | Admin                 | List centers pending approval                    |
|                | `PATCH /centers/:id/status`            | Admin                 | Approve or suspend a center                      |
|                | `DELETE /centers/:id`                  | Admin                 | Delete a center                                  |
| **Orders**     | `POST /centers/:id/orders`             | Authenticated         | Create new order & get upload URL                |
|                | `POST /orders/:id/pay`                 | Authenticated         | Start payment process                            |
|                | `POST /orders/:id/schedule`            | Authenticated         | Set pickup time and print mode                   |
|                | `GET /orders/status/:code`             | All                   | Get order status by pickup code                  |
|                | `GET /orders/:code/receipt`            | Authenticated         | View order receipt                               |
|                | `GET /centers/:id/orders`              | Manager, Admin        | List orders of a center                          |
|                | `POST /orders/:code/verify`            | Manager               | Verify pickup code before printing               |
|                | `POST /orders/:id/print`               | Manager               | Trigger printing                                 |
|                | `PATCH /orders/:id/status`             | Manager, Admin        | Update order status (e.g., CANCELLED, FAILED)    |
|                | `GET /admin/orders`                    | Admin                 | Get all orders across the platform               |
|                | `GET /admin/orders/:id`                | Admin                 | Get detailed info of an order                    |
|                | `DELETE /admin/orders/:id`             | Admin                 | Force delete order                               |
| **Webhooks**   | `POST /webhooks/payment`               | Internal              | Handle asynchronous payment status updates       |
| **System Tasks** | `POST /tasks/order/cleanup`         | Internal              | Delete expired or completed document files       |
|                | `POST /tasks/order/timeout`            | Internal              | Mark overdue orders as CANCELLED                 |

---

### Users API

#### `POST /users`

**Authentication:** Optional (anonymous or authenticated)

**Description:** Registers a new user or syncs an existing Firebase-authenticated user with the backend database.

**Request:**

```json
{
  "uid": "firebase-uid-123",
  "email": "user@example.com",
  "display_name": "Kimba",
  "role": "USER"  // optional: USER, MANAGER, ADMIN
}
```

**Response:**

```json
{
  "uid": "firebase-uid-123",
  "email": "user@example.com",
  "display_name": "Kimba",
  "role": "USER",
  "created_at": "2025-06-23T12:00:00Z"
}
```

**Notes:**

* If the user already exists, it updates fields like `email` or `display_name`.
* Role defaults to `USER` if not provided.

---

#### `GET /users`

**Authentication:** Admin only

**Description:** Lists all users.

**Response:**

```json
[
  {
    "uid": "uid123",
    "email": "user@example.com",
    "display_name": "Alice",
    "role": "USER"
  },
  {
    "uid": "uid456",
    "email": "admin@example.com",
    "display_name": "Admin",
    "role": "ADMIN"
  }
]
```

---

#### `GET /users/:uid`

**Authentication:** Admin only

**Description:** Retrieves full details for a specific user by UID.

**Response:**

```json
{
  "uid": "uid123",
  "email": "user@example.com",
  "display_name": "Alice",
  "role": "USER"
}
```

---

#### `PUT /users/:uid`

**Authentication:** Admin only

**Description:** Updates a user's profile.

**Request:**

```json
{
  "email": "newmail@example.com",
  "display_name": "New Name",
  "role": "MANAGER"
}
```

**Response:**

```json
{
  "updated": true
}
```

---

#### `DELETE /users/:uid`

**Authentication:** Admin only

**Description:** Permanently deletes a user by UID.

**Response:**

```json
{
  "deleted": true
}
```

---

#### `PATCH /users/me`

**Authentication:** Required

**Description:** Allows the authenticated user to update their own profile.

**Request:**

```json
{
  "display_name": "Kimba Updated",
  "email": "kimba@example.com"
}
```

**Response:**

```json
{
  "updated": true
}
```

---

#### `GET /users/me`

**Authentication:** Required

**Description:** Fetches the currently authenticated user's profile.

**Response:**

```json
{
  "uid": "uid123",
  "email": "kimba@example.com",
  "display_name": "Kimba",
  "role": "USER"
}
```

---

#### `DELETE /users/me`

**Authentication:** Required

**Description:** Deletes the account of the currently authenticated user.

**Response:**

```json
{
  "deleted": true
}
```

---

### Print Centers API

#### `GET /centers`

**Authentication:** Not required
**Description:** List all public (approved) print centers. Supports optional query filters like location, city, service type, etc.

**Response:**

```json
[
  {
    "id": "center123",
    "name": "Alpha Print Center",
    "email": "contact@alpha.com",
    "phone_number": "+22991234567",
    "location": {
      "number": 12,
      "type": "Avenue",
      "street": "Kennedy",
      "city": "Cotonou",
      "geo_point": {
        "lat": 6.45,
        "lng": 2.35
      }
    },
    "services": [
      {
        "name": "color print",
        "paper_size": "A4",
        "price": 100,
        "description": "Full color A4 print"
      }
    ],
    "working_hours": [
      {
        "day": "Monday",
        "start": "08:00",
        "end": "18:00"
      }
    ]
  }
]
```

---

#### `POST /centers`

**Authentication:** Authenticated user (user, manager, admin)
**Description:** Registers a new print center owned by the authenticated user.

**Request:**

```json
{
  "name": "Alpha Print Center",
  "email": "owner@alpha.com",
  "phone_number": "+22991234567",
  "location": {
    "number": 12,
    "type": "Avenue",
    "street": "Kennedy",
    "city": "Cotonou",
    "geo_point": {
      "lat": 6.45,
      "lng": 2.35
    }
  },
  "services": [
    {
      "name": "color print",
      "paper_size": "A4",
      "price": 100,
      "description": "High-quality A4 color prints"
    }
  ],
  "working_hours": [
    {
      "day": "Monday",
      "start": "08:00",
      "end": "18:00"
    }
  ]
}
```

**Response:**

```json
{
  "id": "center123",
  "name": "Alpha Print Center",
  "approved": false,
  "owner_uid": "firebase-uid-123"
}
```

**Notes:**

* The center is created with `approved = false` and requires admin approval.
* Ownership is linked to the authenticated user.

---

#### `GET /centers/:id`

**Authentication:** Not required
**Description:** Returns detailed info for the specified center.

**Response:**

```json
{
  "id": "center123",
  "name": "Alpha Print Center",
  "email": "owner@alpha.com",
  "phone_number": "+22991234567",
  "approved": true,
  "location": { ... },
  "services": [ ... ],
  "working_hours": [ ... ]
}
```

---

#### `PUT /centers/:id`

**Authentication:** Required (Manager or Admin)
**Description:** Updates center info. Must be the owner or an admin.

**Request:**

```json
{
  "name": "Updated Print Center",
  "phone_number": "+22991111222",
  "services": [
    {
      "name": "A3 B/W Print",
      "paper_size": "A3",
      "price": 80
    }
  ]
}
```

**Response:**

```json
{
  "updated": true
}
```

---

#### `GET /admin/centers/pending`

**Authentication:** Admin
**Description:** List all print centers awaiting approval.

**Response:**

```json
[
  {
    "id": "center123",
    "name": "Alpha Print Center",
    "email": "owner@alpha.com",
    "owner_uid": "uid123"
  }
]
```

---

#### `PATCH /admin/centers/:id/status`

**Authentication:** Admin
**Description:** Approve or suspend a print center.

**Request:**

```json
{
  "approved": true
}
```

**Response:**

```json
{
  "updated": true
}
```

---

#### `DELETE /admin/centers/:id`

**Authentication:** Admin
**Description:** Delete a print center and all associated data.

**Response:**

```json
{
  "deleted": true
}
```

---

### Orders API

#### `POST /centers/:id/orders`

**Authentication:** Admin, User, Manager
**Description:** Create a new order and get a signed upload URL.

**Request:**

```json
{
  "file_name": "cv.pdf",
  "mime_type": "application/pdf",
  "print_options": {
    "copies": 2,
    "pages": "1-4,7",
    "color": "color",
    "paper_size": "A4"
  }
}
```

**Response:**

```json
{
  "order_id": "ord_123456",
  "upload_url": "https://storage.googleapis.com/printly/documents/ord_123456?signature=abc...",
  "code": "X9A4C2",
  "expires_at": "2025-06-25T10:00:00Z"
}
```

#### Notes

* Creates an order in `AWAITING_DOCUMENT` status.
* Upload URL is valid for 10 minutes.
* The client upload the document to GCS using upload_url. Then provide feedback to backend.
  
#### `POST /orders/:id/pay`

**Authentication:**: Authenticated user (user, manager, admin)
**Description:**: Initiate payment for an order.

**Request:**

```json
{
  "method": "MOBILE_MONEY",
  "provider": "MTN",
  "phone": "+22991234567"
}
```

**Response:**

```json
{
  "payment_url": "https://paygateway.com/session/abc123",
  "status": "PENDING"
}
```

**Notes:**

* Moves order to `PENDING_PAYMENT` status.
* `payment_url` redirects to external gateway.

#### `POST /orders/:id/schedule`

**Authentication:** Authenticated user (user, manager, admin)
**Description:** Schedule pickup time and specify print mode.

**Request:**

```json
{
  "pickup_time": "2025-06-25T10:30:00Z",
  "print_mode": "PRINT_UPON_ARRIVAL"
}
```

**Response:**

```json
{
  "status": "SCHEDULED"
}
```

**Notes:**

* Allowed `print_mode`: `PRE_PRINT` or `PRINT_UPON_ARRIVAL`.

#### `GET /orders/status/:code`

**Authentication:** Not required
**Description:** Check status of an order using pickup code.

**Response:**

```json
{
  "order_id": "ord_123456",
  "status": "AWAITING_USER",
  "print_mode": "PRINT_UPON_ARRIVAL"
}
```

---

#### `GET /orders/:code/receipt`

**Authentication:** Authenticated user (user, manager, admin)
**Description:** Get receipt details by pickup code.

**Response:**

```json
{
  "order_id": "ord_123456",
  "amount": 300,
  "paid": true,
  "printed": true,
  "pickup_time": "2025-06-25T10:30:00Z"
}
```

#### `GET /centers/:id/orders`

**Authentication:** Manager or Admin
**Description:** List orders associated with a given print center.

**Response:**

```json
[
  {
    "order_id": "ord_123456",
    "status": "READY_TO_PRINT",
    "pickup_time": "2025-06-25T10:30:00Z"
  }
]
```

#### `POST /orders/:code/verify`

**Authentication:** Manager
**Description:** Verify a user's pickup code before printing.

**Request:**

```json
{
  "code": "X9A4C2"
}
```

**Response:**

```json
{
  "authorized": true,
  "message": "Valid code"
}
```

---

#### `POST /orders/:id/print`

**Authentication:** Manager
**Description:** Trigger printing of an order.

**Response:**

```json
{
  "status": "PRINTING"
}
```

#### `PATCH /orders/:id/status`

**Authentication:** Manager or Admin
**Description:** Update order status.

**Request:**

```json
{
  "status": "CANCELLED"
}
```

**Response:**

```json
{
  "updated": true
}
```

#### `GET /admin/orders`

**Authentication:** Admin
**Description:** List all orders across the platform.

**Response:**

```json
[
  {
    "order_id": "ord_123456",
    "user_uid": "uid_abc",
    "center_id": "center123",
    "status": "PRINTED",
    "created_at": "2025-06-23T18:00:00Z"
  }
]
```

#### `GET /admin/orders/:id`

**Authentication:** Admin
**Description:** View full details of an order by ID.

**Response:**

```json
{
  "order_id": "ord_123456",
  "user_uid": "uid_abc",
  "center_id": "center123",
  "status": "PRINTED",
  "print_options": {
    "copies": 2,
    "pages": "1-4,7",
    "color": "color",
    "paper_size": "A4"
  },
  "pickup_time": "2025-06-25T10:30:00Z"
}
```

#### `DELETE /admin/orders/:id`

**Authentication:** Admin
**Description:** Force delete an order (admin only).

**Response:**

```json
{
  "deleted": true
}
```
