# 🛣️ Printly Development Roadmap

## 🧱 Phase 1: Core MVP – Anonymous Document Upload & Print Scheduling

**Duration**: 1–2 weeks
**Goal**: Allow anonymous users to upload documents, choose print center, schedule, and pay.

* [ ] Firebase Authentication (Anonymous + Manager roles)
* [ ] Upload endpoint (`/upload`) and document storage (GCS signed URL)
* [ ] Center discovery (`/centers`, `/centers/:id`)
* [ ] Payment integration: mock or real (e.g. Mobile Money)
* [ ] Order creation, pickup code generation
* [ ] Order scheduling (`/schedule`)
* [ ] Status retrieval (`/status/:code`)
* [ ] Minimal UI for placing orders

---

## 🧾 Phase 2: Print Center Management Dashboard

**Duration**: 1–1.5 weeks
**Goal**: Enable managers to register centers, view incoming jobs, and verify print requests.

* [ ] Manager login (Firebase + role-based auth)
* [ ] Register center (`/register/center`)
* [ ] Dashboard view (`/dashboard/orders`)
* [ ] Code verification (`/order/:code/verify`)
* [ ] Manual print trigger (`/order/:id/print`)
* [ ] Real-time updates (polling or WebSocket)
* [ ] UI dashboard for managers

---

## 🔐 Phase 3: Admin Panel & Platform Supervision

**Duration**: 1 week
**Goal**: Platform oversight for verifying centers, seeing usage stats, banning bad actors.

* [ ] Admin auth and role control
* [ ] List pending centers (`/admin/centers/pending`)
* [ ] Approve or delete centers
* [ ] View user list and center details
* [ ] View platform stats (`/dashboard/stats` or `/admin/stats`)
* [ ] Admin UI panel (lightweight)

---

## ⚙️ Phase 4: Reliability, Background Tasks & Cleanup

**Duration**: 1–1.5 weeks
**Goal**: Add reliability, print lifecycle management, and background jobs.

* [ ] Webhook for payment completion (`/webhooks/payment`)
* [ ] Auto-expire unpaid orders (`/tasks/order/timeout`)
* [ ] Auto-delete printed documents (`/tasks/order/cleanup`)
* [ ] Retry logic and failure tracking
* [ ] Email or SMS notifications (optional)
* [ ] Full audit logs (optional)
* [ ] Production readiness checklist (monitoring, secrets, logs)

---

## 🔧 Tech Stack / Tools Used

| Layer             | Tool / Service                     |
| ----------------- | ---------------------------------- |
| Backend           | Go (Gin)                           |
| Auth              | Firebase Auth                      |
| Storage           | Firebase Storage or GCS            |
| DB                | Firestore or PostgreSQL (optional) |
| Payment           | Mobile Money / Mock Gateway        |
| Real-time Updates | WebSocket / SSE / Polling          |
| Deployment        | Cloud Run / App Engine / Render    |
