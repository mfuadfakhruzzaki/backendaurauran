# API Documentation

## Base URL (`https://api.zacht.tech`)

## General Notes

Dibuat dengan Go dan menggunakan framework Gin. Semua request dan response body menggunakan format JSON.

Otentikasi menggunakan JWT. Setiap request yang memerlukan otentikasi harus menyertakan header `Authorization : Bearer <token>`.

Format API menggunakan RESTful API.

Dibuat oleh Kelompok 1 Mata Kuliah Pemrograman Berbasis Objek.

Anggota Kelompok:

- Muhammad Fuad Fakhruzzaki `(21120122130052)`
- Firman Gani Heriansyah `(21120122130043)`
- Raditya Wisnu Cahyo Nugroho `(21120122130039)`
- Rizky Dhafin Almansyah `(21120122120027)`
- Farel Dewangga Rabani `(21120122130037)`

## Auth Routes (`/auth`)

### POST `/auth/register`

- **Method:** `POST`
- **Headers:**
  - `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "username": "string",
    "email": "user@example.com",
    "password": "string",
    "invitation_code": "string" // Optional
  }
  ```

### POST /auth/login

- **Method:** `POST`
- **Headers:**
  - `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "string"
  }
  ```

### POST /auth/logout

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`

### GET /auth/verify-email

- **Method:** `GET`
- **Headers:**
  none
- **Query Parameters:**
  - `token`: string

### POST /auth/request-password-reset

- **Method:** `POST`
- **Headers:**
  - `Content-Type: application/json`
- **Request Body:**

  ```json
  {
    "email": "user@example.com"
  }
  ```

### GET /auth/reset-password

- **Method:** `GET`
- **Headers:**
  none
- **Query Parameters:**
  - `token`: string

### POST /auth/reset-password

- **Method:** `POST`
- **Headers:**
  - `Content-Type: multipart/form-data`
- **Request Body:**

  ```json
  {
    "token": "string",
    "new_password": "string",
    "confirm_password": "string"
  }
  ```

### POST /auth/reset-password-api

- **Method:** `POST`
- **Headers:**
  - `Content-Type: application/json`
- **Request Body:**

  ```json
  {
    "token": "string",
    "new_password": "string"
  }
  ```

## User Routes (`/user`)

### GET /users/profile

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`

### PUT /users/profile

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Request Body:**

  ```json
  {
    "username": "new_username", // Optional
    "email": "new.email@example.com", // Optional
    "password": "newPassword123" // Optional
  }
  ```

### DELETE /users/profile

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`

## Project Routes (`/projects`)

### POST /projects

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Request Body:**

  ```json
  {
    "name": "string",
    "description": "string" // Optional
  }
  ```

### GET /projects

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`

### GET /projects/:id

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### PUT /projects/:id

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**

  ```json
  {
    "name": "string", // Optional
    "description": "string" // Optional
  }
  ```

### DELETE /projects/:id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

## Task Routes (`/projects/:project_id/tasks`)

### POST /projects/:project_id/tasks

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**

  ```json
  {
  "title": "string",
  "description": "string",
  "priority": "low" | "medium" | "high",
  "status": "pending" | "in_progress" | "completed" | "cancelled",
  "deadline": "2024-12-31T23:59:59Z", // Optional (ISO 8601 format)
  "assigned_to": 1                    // Optional (User ID)
  }
  ```

### GET /projects/:project_id/tasks

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### GET /projects/:project_id/tasks/:task_id

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `task_id`: integer

### PUT /projects/:project_id/tasks/:task_id

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
  - `task_id`: integer
- **Request Body:**

  ```json
  {
  "title": "string",               // Optional
  "description": "string",         // Optional
  "priority": "low" | "medium" | "high",    // Optional
  "status": "pending" | "in_progress" | "completed" | "cancelled", // Optional
  "deadline": "2024-12-31T23:59:59Z",     // Optional (ISO 8601 format)
  "assigned_to": 2                    // Optional (User ID)
  }
  ```

### DELETE /projects/:project_id/tasks/:task_id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `task_id`: integer

## Collaboration Routes (`/projects/:project_id/collaborators`)

### POST /projects/:project_id/collaborators

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**

  ```json
  {
  "user_id": 1,
  "role": "admin" | "collaborator"
  }
  ```

### GET /projects/:project_id/collaborators

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### PUT /projects/:project_id/collaborators/:collaborator_id

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
  - `collaborator_id`: integer
- **Request Body:**

  ```json
  {
  "role": "admin" | "collaborator"
  }
  ```

### DELETE /projects/:project_id/collaborators/:collaborator_id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `collaborator_id`: integer

## Note Routes (`/projects/:project_id/notes`)

### POST /projects/:project_id/notes

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**

  ```json
  {
    "content": "string"
  }
  ```

### GET /projects/:project_id/notes

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### GET /projects/:project_id/notes/:note_id

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `note_id`: integer

### PUT /projects/:project_id/notes/:note_id

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
  - `note_id`: integer
- **Request Body:**

  ```json
  {
    "content": "string" // Optional
  }
  ```

### DELETE /projects/:project_id/notes/:note_id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `note_id`: integer

## Activity Routes (`/projects/:project_id/activities`)

### POST /projects/:project_id/activities

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**

  ```json
  {
  "description": "string",
  "type": "task" | "event" | "milestone"
  }
  ```

### GET /projects/:project_id/activities

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### GET /projects/:project_id/activities/:activity_id

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `activity_id`: integer

### PUT /projects/:project_id/activities/:activity_id

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
  - `activity_id`: integer
- **Request Body:**

  ```json
  {
  "description": "string", // Optional
  "type": "task" | "event" | "milestone" // Optional
  }
  ```

### DELETE /projects/:project_id/activities/:activity_id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `activity_id`: integer

## File Routes (`/projects/:project_id/files`)

### POST /projects/:project_id/files

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: multipart/form-data`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**
  - from Data Field: `file` (File)
- Description: Mengupload file ke Google Cloud Storage dan menyimpan metadata file ke database. Hanya tipe file yang diizinkan (misalnya gambar) dan ukuran maksimal 5MB yang diterima.

### GET /projects/:project_id/files

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### GET /projects/:project_id/files/:file_id

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `file_id`: integer

### DELETE /projects/:project_id/files/:file_id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `file_id`: integer

## Notification Routes (`/projects/:project_id/notifications`)

### POST /projects/:project_id/notifications

- **Method:** `POST`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
- **Request Body:**

  ```json
  {
  "user_id": 1,
  "content": "Your task has been updated.",
  "type": "info" | "warning" | "error" | "success",
  "project_id": 2, // Optional
  "is_read": false   // Optional, default false
  }
  ```

### GET /projects/:project_id/notifications

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer

### GET /projects/:project_id/notifications/:notification_id

- **Method:** `GET`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `notification_id`: integer

### PUT /projects/:project_id/notifications/:notification_id

- **Method:** `PUT`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters:**
  - `project_id`: integer
  - `notification_id`: integer
- **Request Body:**

  ```json
  {
  "content": "string",          // Optional
  "type": "info" | "warning" | "error" | "success", // Optional
  "is_read": true               // Optional
  }
  ```

### DELETE /projects/:project_id/notifications/:notification_id

- **Method:** `DELETE`
- **Headers:**
  - `Authorization: Bearer <token>`
- **Path Parameters:**
  - `project_id`: integer
  - `notification_id`: integer
