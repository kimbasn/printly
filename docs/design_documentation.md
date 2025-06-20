# Design Document: Printly – Privacy-First Document Printing Platform

**Author**: Kimba SABI N'GOYE  
**Date**: June 19, 2025  
**Reviewers**: TBD  

---

## 1. Purpose

Printly is a privacy-first, document printing and collection platform. It enables users to upload and print documents at nearby certified printing centers. The system supports mobile-friendly interactions, multiple payment options, and ensures secure print access.

---

## 2. Background

In many developing regions, accessing secure and efficient printing services remains a challenge. Users often have to physically visit printing centers, where the process involves:

* Transferring documents via insecure channels like WhatsApp, USB drives, or Bluetooth
* Exposing sensitive content to center staff
* Waiting in long queues due to uncoordinated service flow
* Facing poor guarantees that their documents are deleted after printing

**Printly** addresses these issues by:

* Allowing users to upload documents and pay remotely—without creating an account
* Preserving confidentiality through one-time print access and automatic file deletion
* Reducing wait times by supporting scheduled pickups and real-time center availability
* Supporting popular local payment options such as Mobile Money

---

## 3. Scope

### In Scope

The initial version of Printly will focus on the following capabilities:

* **Anonymous Document Upload**: Users can submit print jobs without creating an account.
* **Secure Payment Processing**: Integration with payment gateways, including Mobile Money.
* **Order Management & Scheduling**: Users can select pickup times and choose between *Pre-Print* or *Print Upon Arrival* modes.
* **Manager Dashboard**: Printing center managers can view, manage, and process orders.
* **Platform Administration Panel**: Admins can supervise platform-wide activity and moderate centers and content.

### Out of Scope

The following features are explicitly excluded from the current release:

* **Live Document Editing**: Users cannot modify document content within the platform.
* **In-Browser Document Preview**: No support for document rendering or previewing in the web interface.
* **Real-Time Messaging**: No chat or messaging system between users and center managers.
* **Printer Driver Integrations**: The platform does not support advanced printer-specific driver communication or automation beyond controlled print triggers.

---

## 4. 🧩 System Overview

**Client App**: Built with Next.js (Chakra UI)  
**Backend API**: Go (Gin Framework)  
**Authentication**: Firebase (Anonymous + Phone Number(or Email)/Password)  
**Database**: Cloud SQL (PostgreSQL)  
**Document Storage**: Google Cloud Storage (secure, encrypted)  
**Async Jobs**: Google Pub/Sub  
**Payment**: Stripe, Paystack, Mobile Money (MoMo, Orange)  
**Deployment**: Cloud Run + Vercel

### Architecture Diagram

## 5. Data Models

### `User`

```go
type User struct {
    UID         string    `json:"uid" firestore:"uid"`                           // Firebase UID
    Role        string    `json:"role" firestore:"role"`                         // "user" | "manager" | "admin"
    Email       string    `json:"email,omitempty" firestore:"email,omitempty"`   // Optional
    PhoneNumber string    `json:"phone_number" firestore:"phone_number"`
    CenterID    string    `json:"center_id,omitempty" firestore:"center_id,omitempty"` // Only for managers
    CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" firestore:"updated_at"`
}
```

---

### `GeoPoint`

```go
type GeoPoint struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}
```

---

### `Address`

```go
type Address struct {
    Label        string   `json:"label"`
    Street       string   `json:"street"`
    City         string   `json:"city"`
    GeoCoordinate GeoPoint `json:"geo_coordinate"`
}
```

---

### `WorkingHour` and `TimeRange`

```go
type TimeRange struct {
    Start string `json:"start"` // Format: "08:00"
    End   string `json:"end"`   // Format: "18:00"
}

type WorkingHour struct {
    Day      string      `json:"day"`      // e.g. "Monday"
    Intervals []TimeRange `json:"intervals"`
}
```

---

### `Service`

```go
type Service struct {
    Type      string  `json:"type"`       // "color", "black_white"
    PaperSize string  `json:"paper_size"` // "A4", "A3"
    Price     float64 `json:"price"`      // In FCFA or selected currency
}
```

---

### `PrintCenter`

```go
type PrintCenter struct {
    ID           string         `json:"id"`
    Name         string         `json:"name"`
    Email        string         `json:"email"`
    Location     GeoPoint       `json:"location"`               // {Lat, Lng}
    WorkingHours []WorkingHour  `json:"working_hours"`
    Addresses    []Address      `json:"addresses"`
    Services     []Service      `json:"services"`
    Approved     bool           `json:"approved"`               // Set by admin
    OwnerUID     string         `json:"owner_uid"`              // Manager's Firebase UID
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
}
```

---

### `Order`

```go
type OrderStatus string

const (
    StatusPending         OrderStatus = "PENDING"
    StatusPaid            OrderStatus = "PAID"
    StatusReadyToPrint    OrderStatus = "READY_TO_PRINT"
    StatusPrinting        OrderStatus = "PRINTING"
    StatusAwaitingUser    OrderStatus = "AWAITING_USER"
    StatusPrinted         OrderStatus = "PRINTED"
    StatusReadyForPickup  OrderStatus = "READY_FOR_PICKUP"
    StatusCompleted       OrderStatus = "COMPLETED"
    StatusCancelled       OrderStatus = "CANCELLED"
    StatusFailed          OrderStatus = "FAILED"
)

type PrintMode string

const (
    PrePrint         PrintMode = "PRE_PRINT"
    PrintUponArrival PrintMode = "PRINT_UPON_ARRIVAL"
)

type PrintOptions struct {
    Copies     int    `json:"copies"`       // e.g., 2
    Pages      string `json:"pages"`        // e.g., "1-3,5"
    Color      string `json:"color"`        // "color" or "black_white"
    PaperSize  string `json:"paper_size"`   // e.g., "A4"
}

type Order struct {
    ID                  string       `json:"id"`
    Code                string       `json:"code"`         // Pickup code
    UserUID             string       `json:"user_uid"`
    CenterID            string       `json:"center_id"`
    Status              OrderStatus  `json:"status"`
    PrintMode           PrintMode    `json:"print_mode"`
    PickupTime          time.Time    `json:"pickup_time"`
    PrintOptions        PrintOptions `json:"print_options"`
    CreatedAt           time.Time    `json:"created_at"`
    UpdatedAt           time.Time    `json:"updated_at"`
}
```

---

### `Document`

```go
type Document struct {
    OrderID    string    `json:"order_id"`
    FileName   string    `json:"file_name"`
    MimeType   string    `json:"mime_type"`
    URL        string    `json:"url"`          // Temporary signed URL
    UploadedAt time.Time `json:"uploaded_at"`
}
```

---

## 6. Workflows

* [Order Placement](#61-document-upload--order-creation)

### 6.1 Document Upload & Order Creation

This workflow covers how a user uploads a document, configures print options, and creates an order. It supports optional payment and pickup scheduling steps.

#### Actors

| Actor                          | Role                                                            |
| ------------------------------ | --------------------------------------------------------------- |
| **User**                       | Initiates the print job (may be anonymous)                      |
| **Client (Web/App)**           | Authenticates via Firebase and interacts with backend + GCS     |
| **Backend API**                | Validates tokens, manages orders, coordinates payment and logic |
| **Firebase Auth**              | Manages anonymous and authenticated sessions                    |
| **Google Cloud Storage (GCS)** | Receives uploaded files via signed URLs                         |
| **Payment Gateway**            | Processes mobile money or card payments                         |
| **Database**                   | Stores user, order, and print center metadata                   |

#### Status Lifecycle

```text
CREATED
  └── (upload_url issued to client)
AWAITING_DOCUMENT
  └── (GCS confirms file upload via webhook)
PENDING_PAYMENT
  └── (backend receives payment gateway confirmation)
PAID
  └── (client schedules pickup)
READY_TO_PRINT or AWAITING_USER
```

#### Sequence of Operations

1. **User selects:**

   * A print center
   * A document file
   * Print options (`PrintOptions`):

     * `copies`: number of copies
     * `pages`: e.g., `"1-3,5"`
     * `color`: `"color"` or `"black_white"`
     * `paper_size`: `"A4"`, `"A3"`, etc.

2. **Client obtains Firebase ID token**
   (anonymous or authenticated session).

3. **Client → Backend `POST /orders`**

   **Request:**

   ```json
   {
     "center_id": "center123",
     "file_name": "cv.pdf",
     "print_options": {
       "copies": 2,
       "pages": "1-3",
       "color": "black_white",
       "paper_size": "A4"
     }
   }
   ```

4. **Backend:**

   * Verifies Firebase token
   * Creates an `Order` with:

     ```json
     {
       "status": "AWAITING_DOCUMENT",
       "paid": false,
       "printed": false,
       "pickup_time": null,
       "file_url": null
     }
     ```

   * Computes price based on `print_options`
   * Generates a signed `upload_url` (GCS)

5. **Backend → Client**

   **Response:**

    ```json
   {
     "upload_url": "https://storage.googleapis.com/upload_abc",
     "order_id": "order456",
     "receipt": {
       "base_price": 150,
       "copies": 2,
       "total": 300,
       "currency": "XOF"
     }
   }
   ```

6. **Client → PUT document** to `upload_url`

7. **GCS Webhook → Backend (Cloud Function)**

   * Receives upload notification
   * Validates bucket + order ID
   * Updates `Order`:

     ```json
     {
       "status": "PENDING_PAYMENT",
       "file_url": "https://signed_url_to_file"
     }
     ```

8. **Payment Flow** (Optional)

    * **Client → POST `/orders/:id/pay`**

        **Request:**

        ```json
            {
                "method": "MOBILE_MONEY",
                "provider": "MTN",
                "phone": "+22991234567"
            }
        ```

    * **Backend**
      * Initializes payment session
      * Returns:

        ```json
        {
            "payment_url": "https://gateway/pay/abc",
            "status": "PENDING"
        }
        ```

    * **Payment gateway → backend (webhook)**

        * On successful payment:

            ```json
            {
                "order_id": "order456",
                "status": "PAID"
            }
            ```

        * Backend updates `Order.status = PAID`

9. Pickup Scheduling (Optional)

    * **Client → POST `/orders/:id/schedule`**

        **Request:**

        ```json
        {
            "pickup_time": "2025-06-25T10:30:00Z",
            "print_mode": "PRINT_UPON_ARRIVAL"      
        }
        ```

    * **Backend**

        * Updates order with `pickup_time` & `print_mode`
        * Computes final status:

           * `READY_TO_PRINT` if `PRE_PRINT`
           * `AWAITING_USER` if `PRINT_UPON_ARRIVAL`

        * Generates secure pickup `code`

            **Response:**

            ```json
            {
                "code": "X9A4C2"
            }
            ```

#### Summary of Status Transitions For this workflow

| Event                       | New Status                          |
| --------------------------- | ----------------------------------- |
| Order created               | `AWAITING_DOCUMENT`                 |
| File uploaded (GCS webhook) | `PENDING_PAYMENT`                   |
| Payment success             | `PAID`                              |
| Pickup scheduled            | `READY_TO_PRINT` or `AWAITING_USER` |


---

## 7. 📡 API Design

See [API Documentation](./api_documentation.md) for the full REST API, including:

* Auth flow
* Order creation
* Payment session
* Code verification
* Admin endpoints

---

## 8. 🔐 Authentication & Authorization

* Users authenticate via Firebase anonymously
* Managers login with email/password via Firebase
* Backend uses Firebase Admin SDK to validate ID tokens
* Gin middleware ensures token validity and enforces roles

---

## 9. 🔒 Security & Privacy

* Firebase ID tokens securely validate users without manual JWT handling
* Files are stored encrypted in GCS with limited TTL
* Documents are only accessible during their lifecycle (upload → print)
* No preview or download allowed for managers
* HTTPS enforced via Google Cloud

---

## 9. 🔁 Order Status Lifecycle

| Status              | Description                                                            | Set When / Triggered By                                 | Next Possible Status(es)                       |
| ------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------- | ---------------------------------------------- |
| `CREATED`           | Order initialized; client will soon receive upload URL                 | `POST /orders` request                                  | `AWAITING_DOCUMENT`                            |
| `AWAITING_DOCUMENT` | Awaiting file upload by user                                           | After order creation (`POST /orders`)                   | `PENDING_PAYMENT`                              |
| `PENDING_PAYMENT`   | File uploaded, awaiting payment (if required)                          | GCS webhook confirms file upload                        | `PAID`, `CANCELLED`                            |
| `PAID`              | Payment successful                                                     | Payment gateway confirms (webhook)                      | `READY_TO_PRINT`, `AWAITING_USER`              |
| `AWAITING_USER`     | For *Print Upon Arrival* orders: waiting for user to show up with code | After scheduling with `print_mode = PRINT_UPON_ARRIVAL` | `PRINTING`                                     |
| `READY_TO_PRINT`    | For *Pre-Print* orders: manager can print in advance                   | After scheduling with `print_mode = PRE_PRINT`          | `PRINTING`                                     |
| `PRINTING`          | Print job has started (manually or system-triggered)                   | Manager triggers via dashboard                          | `PRINTED`, `FAILED`                            |
| `PRINTED`           | Document successfully printed and deleted                              | System confirms via printer controller or webhook       | `READY_FOR_PICKUP` (if Pre-Print), `COMPLETED` |
| `READY_FOR_PICKUP`  | Waiting for user to come collect printed document                      | After printing in `Pre-Print` mode                      | `COMPLETED`                                    |
| `COMPLETED`         | User collected document (confirmation via UI or manager)               | Pickup code verified and validated                      | —                                              |
| `CANCELLED`         | Order was cancelled by user, timeout, or failure to pay                | Manual cancel or background timeout task                | —                                              |
| `FAILED`            | Technical issue during printing (retry-eligible or permanent)          | Printer backend or error handling                       | `PRINTING` (retry), `CANCELLED`                |

---

### 🔁 Typical Status Transition Paths

#### A. Pre-Print Mode

```text
CREATED → AWAITING_DOCUMENT → PENDING_PAYMENT → PAID → READY_TO_PRINT → PRINTING → PRINTED → READY_FOR_PICKUP → COMPLETED
```

#### B. Print Upon Arrival Mode

```text
CREATED → AWAITING_DOCUMENT → PENDING_PAYMENT → PAID → AWAITING_USER → PRINTING → PRINTED → COMPLETED
```

#### C. If Payment Skipped (Free Services)

```text
CREATED → AWAITING_DOCUMENT → PAID → AWAITING_USER or READY_TO_PRINT → ...
```

#### D. Order Cancelled / Failed

```text
CREATED → AWAITING_DOCUMENT → CANCELLED
PENDING_PAYMENT → CANCELLED
PRINTING → FAILED → PRINTING (retry) or CANCELLED
```

---

## 10. 🧪 Testing Strategy

* Unit tests for each Gin handler and service
* Integration tests with PostgreSQL test DB
* E2E tests with Cypress for user flows
* Mock payment APIs for non-production environments

---

## 11. 🚀 Deployment Plan

* Containerize backend and frontend
* Use Cloud Run for stateless backend
* Vercel or Cloud Run for Next.js frontend
* GitHub Actions CI/CD with secrets managed via Secret Manager
* Infrastructure can be provisioned with Terraform

---

## 12. 📊 Monitoring & Logging

* GCP Cloud Logging for access and application logs
* Pub/Sub Dead Letter Topics for task failures
* Alerts via GCP Monitoring for API errors, queue buildup
* Per-order tracing using request IDs

---

## 13. 📆 Roadmap Milestones

| Milestone               | Target Date   |
|-------------------------|---------------|
| MVP with Upload + Payment | July 1, 2025 |
| Dashboard + Pre-Print     | July 10, 2025 |
| Admin Panel Live          | July 20, 2025 |
| Production Ready          | July 31, 2025 |

---

## 14. 📌 Future Work

* QR code pickup validation
* Order notification via WhatsApp/SMS
* Printer status health monitoring
* Offline document queuing
* Multi-language support

---
