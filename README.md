# API Documentation

## Base URL (`https://api.zacht.tech`)

### General Notes

- API dibuat menggunakan Go dan framework Gin.
- Semua request dan response body menggunakan format JSON.
- Otentikasi menggunakan JWT. Tambahkan header berikut untuk request yang membutuhkan otentikasi:
  ```
  Authorization: Bearer <token>
  ```
- Format API menggunakan RESTful.
- Dokumentasi ini disusun oleh Kelompok 1 dari Mata Kuliah Pemrograman Berbasis Objek.

### Anggota Kelompok
- Muhammad Fuad Fakhruzzaki `(21120122130052)`
- Firman Gani Heriansyah `(21120122130043)`
- Raditya Wisnu Cahyo Nugroho `(21120122130039)`
- Rizky Dhafin Almansyah `(21120122120027)`
- Farel Dewangga Rabani `(21120122130037)`

---

## Endpoints

### 1. **Auth Routes**

#### POST `/auth/register`
- **Headers:** `Content-Type: application/json`
- **Body:**
  ```json
  {
    "username": "zacht",
    "email": "user@example.com",
    "password": "string",
    "invitation_code": "optional"
  }
  ```

#### POST `/auth/login`
- **Headers:** `Content-Type: application/json`
- **Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "string"
  }
  ```

#### POST `/auth/logout`
- **Headers:** `Authorization: Bearer <token>`

#### GET `/auth/verify-email`
- **Query Parameter:**
  - `token`: string

#### POST `/auth/request-password-reset`
- **Headers:** `Content-Type: application/json`
- **Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```

#### POST `/auth/reset-password`
- **Headers:** `Content-Type: application/json`
- **Body:**
  ```json
  {
    "token": "reset_token",
    "new_password": "string"
  }
  ```

---

### 2. **User Routes**

#### GET `/users/profile`
- **Headers:** `Authorization: Bearer <token>`

#### PUT `/users/profile`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Body:**
  ```json
  {
    "username": "new_username",
    "email": "new.email@example.com",
    "password": "newPassword123"
  }
  ```

#### DELETE `/users/profile`
- **Headers:** `Authorization: Bearer <token>`

---

### 3. **Project Routes**

#### POST `/projects`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Body:**
  ```json
  {
    "title": "Project Title",
    "description": "Description",
    "priority": "Medium",
    "deadline": "2024-12-31T23:59:59Z",
    "status": "Pending"
  }
  ```

#### GET `/projects`
- **Headers:** `Authorization: Bearer <token>`

#### GET `/projects/:id`
- **Headers:** `Authorization: Bearer <token>`

#### PUT `/projects/:id`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Body:**
  ```json
  {
    "title": "Updated Title",
    "description": "Updated Description"
  }
  ```

#### DELETE `/projects/:id`
- **Headers:** `Authorization: Bearer <token>`

---

### 4. **Task Routes**

#### POST `/projects/:project_id/tasks`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Body:**
  ```json
  {
    "title": "Task Title",
    "description": "Description",
    "priority": "High",
    "status": "In Progress",
    "deadline": "2024-12-31T23:59:59Z",
    "assigned_to_id": 4
  }
  ```

#### GET `/projects/:project_id/tasks`
- **Headers:** `Authorization: Bearer <token>`

#### GET `/projects/:project_id/tasks/:task_id`
- **Headers:** `Authorization: Bearer <token>`

#### PUT `/projects/:project_id/tasks/:task_id`
- **Headers:**
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Body:**
  ```json
  {
    "title": "Updated Task Title",
    "status": "Completed"
  }
  ```

#### DELETE `/projects/:project_id/tasks/:task_id`
- **Headers:** `Authorization: Bearer <token>`
