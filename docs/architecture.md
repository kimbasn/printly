# **System Architecture: Printly**

## 1. **Component Overview**

```text
                        ┌────────────────────────┐
                        │   👩‍💻 Client (Browser)   │
                        │  - Next.js Frontend     │
                        │  - Firebase SDK (Auth)  │
                        └────────────┬───────────┘
                                     │
                             🔒 Auth with Firebase
                                     │
                                     ▼
                        ┌────────────────────────┐
                        │  🌐 Gin API Backend     │
                        │  - Go (Gin Framework)   │
                        │  - Firebase Admin SDK   │
                        └───────┬────┬────┬───────┘
                                │    │    │
     ┌──────────────────────────┘    │    └─────────────────────────────┐
     ▼                               ▼                                  ▼
┌────────────┐             ┌─────────────────┐                 ┌──────────────────┐
│ Cloud SQL  │             │ Cloud Storage   │                 │ Pub/Sub & Tasks  │
│ PostgreSQL │             │ (Private Docs)  │                 │ Print, Cleanup   │
└────────────┘             └─────────────────┘                 └──────────────────┘

                          ▼
                 ┌───────────────────┐
                 │ Payments (Stripe, │
                 │ Paystack, Mobile  │
                 │ Money APIs)       │
                 └───────────────────┘
```

---

## 🔐 2. **Authentication Flow**

* User signs in anonymously or via Firebase Email/Password
* Firebase returns an **ID token**
* Frontend sends this token to backend (in `Authorization: Bearer <token>`)
* Backend uses **Firebase Admin SDK** to:

  * Verify token
  * Extract user UID
  * Authorize based on role (`USER`, `MANAGER`, `ADMIN`)

---

## 🚦 3. **Request Flow**

### Example: Uploading a File to Print

1. User uploads file metadata → `POST /upload`
2. API verifies user and creates order in DB
3. API generates a signed URL from **GCS**
4. User uploads file directly to **Cloud Storage**
5. API tracks file in DB with TTL
6. Order awaits payment → `POST /pay`
7. After payment, job is queued via **Pub/Sub**
8. Manager prints or user triggers print with code

---

## 🔧 4. **Background Services**

| Task                     | Technology           | Description                               |
| ------------------------ | -------------------- | ----------------------------------------- |
| Print Job Processing     | Cloud Tasks / PubSub | Secure document retrieval + print trigger |
| Order Cleanup            | Cloud Scheduler      | TTL-based deletion of uncollected docs    |
| Payment Webhook Handling | Gin + Stripe SDK     | Update order status upon confirmation     |

---

## 📦 5. **Deployment**

| Component     | Platform        | Details                     |
| ------------- | --------------- | --------------------------- |
| Frontend      | Vercel (or GCP) | Next.js with Chakra UI      |
| Backend API   | Cloud Run       | Containerized Gin App       |
| Database      | Cloud SQL       | PostgreSQL                  |
| File Storage  | GCS             | Signed URLs, access-limited |
| Auth Provider | Firebase Auth   | Session & user management   |
| CI/CD         | GitHub Actions  | Auto-deploy via Cloud Build |

---

## 🧠 6. **Monitoring & Logging**

| Service          | Purpose                          |
| ---------------- | -------------------------------- |
| Cloud Logging    | API logs, errors                 |
| Cloud Monitoring | Metrics, latency, failure alerts |
| Pub/Sub DLQ      | Catch failed background jobs     |

---

## 📂 7. **Modular Codebase Structure (Go + Gin)**

```
/cmd/server          → Main entrypoint
/internal/
  ├── auth           → Firebase token validation
  ├── center         → Business logic for printing centers
  ├── order          → Order lifecycle and handlers
  ├── payment        → Payment integration
  ├── storage        → GCS upload URL generation
  ├── admin          → Platform-level tools
  └── middleware     → Auth middleware for roles
/pkg/
  ├── models         → DB schemas
  ├── config         → Env + Firebase/DB configs
  └── utils          → Shared helpers (dates, UUIDs, etc.)
```

