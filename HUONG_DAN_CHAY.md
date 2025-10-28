# 🚀 HƯỚNG DẪN CHẠY PROJECT REALTIME CHAT

## 📦 Yêu cầu hệ thống
- Docker Desktop đã cài đặt và đang chạy
- Trình duyệt web (Chrome, Firefox, Edge...)

---

## 🔧 1. KHỞI ĐỘNG HỆ THỐNG

### Bước 1: Mở terminal và di chuyển vào thư mục infra
```powershell
cd d:\LTM\e5-realtime-chat\infra
```

### Bước 2: Khởi động tất cả services bằng Docker Compose
```powershell
docker-compose up -d
```

**Giải thích lệnh:**
- `docker-compose up`: Khởi động các services được định nghĩa trong `docker-compose.yml`
- `-d`: Chạy ở chế độ background (detached mode)

### Bước 3: Kiểm tra trạng thái các container
```powershell
docker-compose ps
```

Bạn sẽ thấy các container sau đang chạy:
- ✅ **e5-postgres** - PostgreSQL database (port 5432)
- ✅ **e5-server** - Go WebSocket server (port 8080)
- ✅ **e5-client** - Nginx web server (port 3000)
- ✅ **e5-redis** - Redis cache
- ✅ **e5-k6** - Load testing tool

---

## 🌐 2. TRUY CẬP ỨNG DỤNG

### Frontend (Giao diện web)
Mở trình duyệt và truy cập:
```
http://localhost:3000
```

**Các trang có sẵn:**
- 🏠 Trang chủ: `http://localhost:3000/index.html`
- 🔐 Đăng nhập: `http://localhost:3000/login.html`
- 📝 Đăng ký: `http://localhost:3000/register.html`

### Backend API
- 🔌 WebSocket endpoint: `ws://localhost:8080/ws`
- 💚 Health check: `http://localhost:8080/healthz`
- 👥 Friends API: `http://localhost:8080/api/friends`

### Database
**Thông tin kết nối PostgreSQL:**
- Host: `localhost`
- Port: `5432`
- Database: `chatdb`
- Username: `chatuser`
- Password: `chatpass`

---

## 🗄️ 3. QUẢN LÝ DATABASE

### Kết nối vào PostgreSQL container
```powershell
docker exec -it e5-postgres psql -U chatuser -d chatdb
```

### Các lệnh SQL hữu ích
```sql
-- Xem tất cả bảng
\dt

-- Xem cấu trúc bảng users
\d users

-- Lấy danh sách tất cả users
SELECT * FROM users;

-- Lấy danh sách tin nhắn gần nhất
SELECT * FROM messages ORDER BY created_at DESC LIMIT 10;

-- Lấy danh sách phòng chat
SELECT * FROM rooms;

-- Thoát khỏi psql
\q
```

### Xem cấu trúc database
Database của bạn có các bảng sau (đã được tạo tự động từ file `migrations/001_init.sql`):

1. **users** - Lưu thông tin người dùng
   - id, username, email, password_hash, avatar_url
   - is_online, last_seen_at, created_at

2. **rooms** - Lưu phòng chat
   - id, room_name, room_type, description
   - created_by, created_at

3. **messages** - Lưu tin nhắn
   - id, message_type, from_user_id, to_user_id
   - room_id, text, value, is_read, created_at

4. **friendships** - Quan hệ bạn bè
   - id, user_id, friend_id, status
   - created_at, accepted_at

5. **room_members** - Thành viên trong phòng
   - id, room_id, user_id, role
   - joined_at, last_read_message_id

6. **user_sessions** - Quản lý phiên đăng nhập
   - id, user_id, session_token, ip_address
   - user_agent, created_at, expires_at

---

## 🔍 4. KIỂM TRA HỆ THỐNG

### Xem logs của từng service
```powershell
# Xem log database
docker-compose logs postgres

# Xem log server
docker-compose logs server

# Xem log client
docker-compose logs client

# Xem log real-time (theo dõi liên tục)
docker-compose logs -f server
```

### Kiểm tra server có hoạt động không
```powershell
curl http://localhost:8080/healthz
```
Kết quả mong đợi: `ok`

---

## 🛑 5. DỪNG VÀ KHỞI ĐỘNG LẠI

### Dừng tất cả services
```powershell
docker-compose down
```

### Dừng và xóa cả dữ liệu database
```powershell
docker-compose down -v
```
⚠️ **Cảnh báo:** Lệnh này sẽ xóa tất cả dữ liệu trong database!

### Khởi động lại sau khi dừng
```powershell
docker-compose up -d
```

### Rebuild toàn bộ (khi có thay đổi code)
```powershell
docker-compose up -d --build
```

---

## 🔧 6. TÍCH HỢP DATABASE VÀO FRONTEND (KẾ HOẠCH SAU NÀY)

Hiện tại frontend chưa kết nối database, nhưng backend đã sẵn sàng. 

### Các API endpoint có thể dùng (khi tích hợp):

**Trong file `server/database/db.go` đã có sẵn các phương thức:**

1. **User Management:**
   - `GetUserByID(id)` - Lấy user theo ID
   - `GetUserByUsername(username)` - Lấy user theo username
   - `GetUserByEmail(email)` - Lấy user theo email
   - `CreateUser(user)` - Tạo user mới
   - `UpdateUser(user)` - Cập nhật user
   - `UpdateUserOnlineStatus(id, isOnline)` - Cập nhật trạng thái online
   - `GetOnlineUsers()` - Lấy danh sách user đang online

2. **Message Management:**
   - `CreateMessage(msg)` - Lưu tin nhắn mới
   - `GetMessagesByRoom(roomID, limit)` - Lấy tin nhắn theo phòng
   - `GetPrivateMessages(user1ID, user2ID, limit)` - Lấy tin nhắn riêng tư
   - `MarkMessageAsRead(messageID)` - Đánh dấu đã đọc
   - `GetUnreadMessageCount(userID)` - Đếm tin nhắn chưa đọc

3. **Room Management:**
   - `CreateRoom(room)` - Tạo phòng chat
   - `GetRoomByID(id)` - Lấy thông tin phòng
   - `GetRoomByName(name)` - Tìm phòng theo tên
   - `GetPublicRooms()` - Lấy danh sách phòng công khai
   - `AddUserToRoom(roomID, userID, role)` - Thêm user vào phòng
   - `RemoveUserFromRoom(roomID, userID)` - Xóa user khỏi phòng

4. **Friendship Management:**
   - `CreateFriendship(userID, friendID)` - Tạo lời mời kết bạn
   - `AcceptFriendship(friendshipID)` - Chấp nhận lời mời
   - `GetFriends(userID)` - Lấy danh sách bạn bè
   - `GetPendingFriendRequests(userID)` - Lấy lời mời kết bạn chờ duyệt

### Để tích hợp sau này:

**Bước 1:** Tạo REST API endpoints trong `server/main.go`:
```go
http.HandleFunc("/api/login", loginHandler)
http.HandleFunc("/api/register", registerHandler)
http.HandleFunc("/api/messages", messagesHandler)
http.HandleFunc("/api/users", usersHandler)
```

**Bước 2:** Cập nhật JavaScript trong `client/assets/js/`:
```javascript
// Gọi API từ frontend
fetch('http://localhost:8080/api/login', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({username, password})
})
.then(response => response.json())
.then(data => {
    // Xử lý response
});
```

---

## 📊 7. CÔNG CỤ QUẢN LÝ DATABASE (TÙY CHỌN)

### Cách 1: pgAdmin (GUI tool)
1. Tải và cài đặt [pgAdmin](https://www.pgadmin.org/)
2. Tạo kết nối mới:
   - Host: `localhost`
   - Port: `5432`
   - Database: `chatdb`
   - Username: `chatuser`
   - Password: `chatpass`

### Cách 2: DBeaver (Universal Database Tool)
1. Tải [DBeaver Community Edition](https://dbeaver.io/download/)
2. Tạo kết nối PostgreSQL với thông tin trên

### Cách 3: VS Code Extension
1. Cài extension "PostgreSQL" của Chris Kolkman
2. Kết nối với thông tin database ở trên

---

## 🐛 8. TROUBLESHOOTING

### Lỗi: Port đã được sử dụng
```
Error: Bind for 0.0.0.0:3000 failed: port is already allocated
```
**Giải pháp:** Dừng ứng dụng đang chạy trên port đó hoặc đổi port trong `docker-compose.yml`

### Lỗi: Database connection failed
**Giải pháp:** Đợi vài giây để PostgreSQL khởi động hoàn tất, sau đó restart server:
```powershell
docker-compose restart server
```

### Lỗi: Cannot connect to Docker daemon
**Giải pháp:** Đảm bảo Docker Desktop đang chạy

### Reset toàn bộ hệ thống
```powershell
# Dừng và xóa tất cả
docker-compose down -v

# Xóa các image cũ
docker rmi e5-realtime-server:latest

# Khởi động lại từ đầu
docker-compose up -d --build
```

---

## 📝 9. CẤU TRÚC PROJECT

```
e5-realtime-chat/
├── client/                 # Frontend
│   ├── html/              # Các file HTML
│   │   ├── index.html     # Trang chat chính
│   │   ├── login.html     # Trang đăng nhập
│   │   └── register.html  # Trang đăng ký
│   └── assets/            # CSS, JS, images
│       ├── css/           # Stylesheets
│       └── js/            # JavaScript files
│
├── server/                # Backend (Go)
│   ├── main.go            # Entry point
│   ├── client.go          # WebSocket client logic
│   ├── hub.go             # WebSocket hub (message broker)
│   ├── friends.go         # Friends API handler
│   ├── database/          # Database layer
│   │   └── db.go          # Database operations
│   └── migrations/        # SQL migration files
│       └── 001_init.sql   # Initial schema
│
└── infra/                 # Infrastructure
    ├── docker-compose.yml # Docker orchestration
    └── nginx/             # Nginx config
        └── default.conf
```

---

## 🎯 10. NEXT STEPS (KẾ HOẠCH PHÁT TRIỂN)

### Phase 1: Hiện tại ✅
- [x] Frontend static HTML/CSS/JS
- [x] Backend WebSocket server
- [x] PostgreSQL database với schema hoàn chỉnh
- [x] Docker containerization

### Phase 2: Kết nối Frontend - Backend (Kế hoạch)
- [ ] Tạo REST API cho login/register
- [ ] Xác thực JWT tokens
- [ ] Kết nối WebSocket từ frontend
- [ ] Lưu messages vào database
- [ ] Hiển thị history từ database

### Phase 3: Tính năng nâng cao (Tương lai)
- [ ] Upload avatar/files
- [ ] Typing indicators
- [ ] Online/offline status real-time
- [ ] Push notifications
- [ ] Group chat rooms
- [ ] Search messages
- [ ] Emoji reactions

---

## 📞 LIÊN HỆ & HỖ TRỢ

Nếu gặp vấn đề, kiểm tra:
1. Docker Desktop có đang chạy không?
2. Các port 3000, 8080, 5432 có bị chiếm không?
3. Logs của các container có lỗi gì không?

**Xem logs chi tiết:**
```powershell
docker-compose logs -f
```

---

## 📚 TÀI LIỆU THAM KHẢO

- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Go WebSocket Tutorial](https://github.com/gorilla/websocket)
- [Nginx Documentation](https://nginx.org/en/docs/)

---

**🎉 Chúc bạn phát triển project thành công!**
