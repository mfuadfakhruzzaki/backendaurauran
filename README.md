# Backend Aurauran

Backend Aurauran adalah aplikasi backend yang dibangun dengan Go dan framework Gin. Aplikasi ini menyediakan API untuk autentikasi pengguna, manajemen proyek, dan sumber daya terkait proyek.

## Ringkasan Endpoint

### Autentikasi

- **POST /auth/register** - Mendaftarkan pengguna baru.
- **POST /auth/login** - Login pengguna.
- **POST /auth/logout** - Logout pengguna.
- **GET /auth/verify-email** - Verifikasi email pengguna.
- **POST /auth/request-password-reset** - Meminta reset kata sandi.
- **POST /auth/reset-password** - Reset kata sandi pengguna.
- **GET /auth/reset-password** - Menampilkan formulir reset kata sandi.
- **POST /auth/reset-password-api** - Reset kata sandi melalui API.

### Pengguna

- **GET /users/profile** - Mendapatkan profil pengguna.
- **PUT /users/profile** - Memperbarui profil pengguna.
- **DELETE /users/profile** - Menghapus profil pengguna.

### Proyek

- **POST /projects/** - Membuat proyek baru.
- **GET /projects/** - Mendapatkan daftar proyek.
- **GET /projects/:project_id** - Mendapatkan detail proyek.
- **PUT /projects/:project_id** - Memperbarui proyek.
- **DELETE /projects/:project_id** - Menghapus proyek.

#### Kolaborator

- **POST /projects/:project_id/collaborators/** - Menambahkan kolaborator.
- **GET /projects/:project_id/collaborators/** - Mendapatkan daftar kolaborator.
- **PUT /projects/:project_id/collaborators/:collaborator_id** - Memperbarui peran kolaborator.
- **DELETE /projects/:project_id/collaborators/:collaborator_id** - Menghapus kolaborator.

#### Aktivitas

- **POST /projects/:project_id/activities/** - Membuat aktivitas.
- **GET /projects/:project_id/activities/** - Mendapatkan daftar aktivitas.
- **GET /projects/:project_id/activities/:id** - Mendapatkan detail aktivitas.
- **PUT /projects/:project_id/activities/:id** - Memperbarui aktivitas.
- **DELETE /projects/:project_id/activities/:id** - Menghapus aktivitas.

#### Tugas

- **POST /projects/:project_id/tasks/** - Membuat tugas.
- **GET /projects/:project_id/tasks/** - Mendapatkan daftar tugas.
- **GET /projects/:project_id/tasks/:id** - Mendapatkan detail tugas.
- **PUT /projects/:project_id/tasks/:id** - Memperbarui tugas.
- **DELETE /projects/:project_id/tasks/:id** - Menghapus tugas.

#### Catatan

- **POST /projects/:project_id/notes/** - Membuat catatan.
- **GET /projects/:project_id/notes/** - Mendapatkan daftar catatan.
- **GET /projects/:project_id/notes/:id** - Mendapatkan detail catatan.
- **PUT /projects/:project_id/notes/:id** - Memperbarui catatan.
- **DELETE /projects/:project_id/notes/:id** - Menghapus catatan.

#### Berkas

- **POST /projects/:project_id/files/** - Mengunggah berkas.
- **GET /projects/:project_id/files/** - Mendapatkan daftar berkas.
- **GET /projects/:project_id/files/:id** - Mengunduh berkas.
- **DELETE /projects/:project_id/files/:id** - Menghapus berkas.

#### Notifikasi

- **POST /projects/:project_id/notifications/** - Membuat notifikasi.
- **GET /projects/:project_id/notifications/** - Mendapatkan daftar notifikasi.
- **GET /projects/:project_id/notifications/:id** - Mendapatkan detail notifikasi.
- **PUT /projects/:project_id/notifications/:id** - Memperbarui notifikasi.
- **DELETE /projects/:project_id/notifications/:id** - Menghapus notifikasi.

---

**Catatan:** Semua endpoint di bawah path `/projects/` dilindungi dan memerlukan autentikasi. Anda harus menyertakan token autentikasi dalam permintaan Anda.

---

