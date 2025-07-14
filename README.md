# Printly API

[![Go Report Card](https://goreportcard.com/badge/github.com/kimbasn/printly)](https://goreportcard.com/report/github.com/kimbasn/printly)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Printly** is a privacy-first, document printing and collection platform. It enables users to securely upload and print documents at nearby certified printing centers without compromising their privacy.

The backend is built with Go (Gin) and follows clean architecture principles for maintainability and scalability.

---

## ✨ Key Features

- **Secure Authentication**: Robust authentication and authorization using Firebase JWTs.
- **Role-Based Access Control (RBAC)**: Granular permissions for `user`, `manager`, and `admin` roles.
- **Print Center Management**: Endpoints for registering, approving, and managing print centers.
- **Order Management**: Complete workflow for creating, tracking, and processing print orders.
- **User Self-Service**: Secure endpoints for users to manage their own profiles.
- **API Documentation**: Comprehensive API documentation powered by Swagger.

---

## 🛠️ Tech Stack

| Layer             | Tool / Service                               |
| ----------------- | -------------------------------------------- |
| **Backend**       | Go (Gin Framework)                           |
| **Database**      | PostgreSQL                                   |
| **Authentication**| Firebase Authentication                      |
| **Testing**       | Go Test, Testify, GoMock                     |
| **Validation**    | Go Playground Validator                      |
| **Deployment**    | Cloud Run                                    |

---

## 🚀 Getting Started

Follow these instructions to get the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go (version 1.21 or later)
- A Firebase project for authentication.

### 1. Clone the Repository

```bash
git clone https://github.com/kimbasn/printly.git
cd printly
```

### 2. Configuration

The application is configured using environment variables. Create a `.env` file in the root of the project by copying the example file:

```bash
cp .env.example .env
```

Now, edit the `.env` file with your local configuration:

- **Database**: Set your PostgreSQL connection details (`DB_USER`, `DB_PASSWORD`, `DB_NAME`).
- **Firebase**:
    1. Go to your Firebase project settings -> Service Accounts.
    2. Generate a new private key and download the JSON file.
    3. Save this file in the project root (e.g., as `serviceAccountKey.json`).
    4. Set `FIREBASE_CREDENTIALS_FILE=./serviceAccountKey.json` in your `.env` file.

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Application

#### Using Go

If you have a local PostgreSQL instance running, you can start the server directly:

```bash
make run
```

The API server will be available at `http://localhost:8080`.

---

## 📚 API Documentation

Once the server is running, you can access the interactive Swagger API documentation at:

**<http://localhost:8080/swagger/index.html>**

This UI allows you to explore and test all the API endpoints. You can authorize your requests by obtaining a Firebase ID token from your client application and using the "Authorize" button.

---

## 🛣️ Roadmap

Our development plan is outlined in the Roadmap document.

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE.md file for details.
