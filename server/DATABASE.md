# 🗄️ Database Setup Guide

## Quick Start với Docker

### 1. Chạy Database với Docker Compose

```bash
cd infra
docker-compose up -d postgres
```

Database sẽ chạy trên:
- **Host:** localhost
- **Port:** 5432
- **Database:** chatdb
- **Username:** chatuser
- **Password:** chatpass

### 2. Kiểm tra Database đã chạy

```bash
docker ps | grep postgres
```

Hoặc kiểm tra logs:

```bash
docker logs e5-postgres
```

### 3. Kết nối vào Database (Optional)

```bash
# Sử dụng psql
docker exec -it e5-postgres psql -U chatuser -d chatdb

# Hoặc sử dụng pgAdmin, DBeaver, etc với thông tin ở trên
```

---

## Database Schema

### Tables:

1. **users** - Thông tin người dùng
   - id, username, email, password_hash, avatar_url
   - is_online, last_seen_at, created_at

2. **messages** - Tin nhắn chat
   - id, message_type, from_user_id, to_user_id, room_id
   - message_text, message_value, created_at

3. **rooms** - Phòng chat
   - id, room_name, room_type, description, created_by

4. **friendships** - Quan hệ bạn bè
   - id, user_id, friend_id, status

5. **user_sessions** - Sessions & JWT tokens
   - id, user_id, token, expires_at

6. **room_members** - Thành viên trong room
   - id, room_id, user_id, role

---

## Sample Data

Database đã được populate với dữ liệu mẫu:

### Users (password: `password123`):
- `admin` - Administrator
- `alice` - Regular user
- `bob` - Regular user

### Rooms:
- `general` - Public chat room

### Friendships:
- alice ↔ bob (accepted)

---

## Queries Hữu Ích

### Xem tất cả users:
```sql
SELECT id, username, email, is_online FROM users;
```

### Xem tin nhắn gần đây:
```sql
SELECT * FROM v_recent_messages LIMIT 10;
```

### Xem danh sách bạn bè của user:
```sql
SELECT * FROM v_user_friends WHERE user_id = 1;
```

### Xóa tất cả messages (test):
```sql
DELETE FROM messages;
```

---

## Migrations

Migration files ở folder: `server/migrations/`

- `001_init.sql` - Initial schema

Khi start container, PostgreSQL tự động chạy tất cả `.sql` files trong folder này.

---

## Troubleshooting

### Database không start:

```bash
# Kiểm tra logs
docker logs e5-postgres

# Xóa và tạo lại
docker-compose down -v
docker-compose up -d postgres
```

### Connection refused:

Đợi 5-10 giây sau khi start container để database khởi động hoàn toàn.

### Reset database:

```bash
docker-compose down -v postgres
docker volume rm infra_postgres_data
docker-compose up -d postgres
```

---

## Production Notes

⚠️ **Trước khi deploy production:**

1. ✅ Đổi `POSTGRES_PASSWORD` trong `docker-compose.yml`
2. ✅ Xóa sample users hoặc đổi password
3. ✅ Enable SSL: `sslmode=require`
4. ✅ Backup database định kỳ
5. ✅ Monitor database performance
6. ✅ Setup connection pooling
7. ✅ Add database replicas (read replicas)

---

## Backup & Restore

### Backup:
```bash
docker exec e5-postgres pg_dump -U chatuser chatdb > backup.sql
```

### Restore:
```bash
docker exec -i e5-postgres psql -U chatuser chatdb < backup.sql
```

---

## Connect từ Go Code

```go
import "e5realtimechat/internal/database"

db, err := database.NewDB(
    "localhost", // host
    "5432",      // port
    "chatuser",  // user
    "chatpass",  // password
    "chatdb",    // dbname
)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```
