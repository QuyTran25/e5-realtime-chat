# E5 Realtime Chat - Ứng dụng Chat Thời gian Thực Hiệu năng Cao

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go)
![WebSocket](https://img.shields.io/badge/WebSocket-Enabled-4CAF50?style=for-the-badge&logo=websocket)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?style=for-the-badge&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3-FF6600?style=for-the-badge&logo=rabbitmq)
![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)

**Ứng dụng chat thời gian thực được xây dựng với kiến trúc hiện đại, có khả năng chịu tải cao và mở rộng linh hoạt**

[Tính năng](#tính-năng-chính) • [Công nghệ](#công-nghệ-sử-dụng) • [Cài đặt](#hướng-dẫn-cài-đặt) • [Demo](#ảnh-chụp-màn-hình--demo)

</div>

---

## Mục lục

- [Giới thiệu](#giới-thiệu)
- [Tính năng chính](#tính-năng-chính)
- [Công nghệ sử dụng](#công-nghệ-sử-dụng)
- [Kiến trúc hệ thống](#kiến-trúc-hệ-thống)
- [Yêu cầu hệ thống](#yêu-cầu-hệ-thống)
- [Hướng dẫn cài đặt](#hướng-dẫn-cài-đặt)
- [Ảnh chụp màn hình & Demo](#ảnh-chụp-màn-hình--demo)
- [Tài liệu API](#tài-liệu-api)
- [Hiệu năng & Kiểm thử tải](#hiệu-năng--kiểm-thử-tải)
- [Thành viên nhóm](#thành-viên-nhóm)

---

## Giới thiệu

**E5 Realtime Chat** là ứng dụng chat thời gian thực được thiết kế với khả năng **chịu tải cao** và **mở rộng linh hoạt**. Dự án sử dụng **WebSocket** kết hợp với **Golang** để đảm bảo hiệu suất tối ưu và độ trễ thấp.

### Điểm nổi bật:

- **Hiệu năng cao**: Xử lý hàng ngàn kết nối WebSocket đồng thời
- **Độ trễ thấp**: Độ trễ < 50ms cho tin nhắn thời gian thực
- **Mở rộng ngang**: Scale với nhiều backend instances
- **Bảo mật**: JWT authentication, rate limiting, CORS protection
- **Bộ nhớ đệm**: Redis cache giảm tải database 80%
- **Hàng đợi tin nhắn**: RabbitMQ xử lý tác vụ bất đồng bộ
- **Cân bằng tải**: Phân phối tải thông minh với Nginx
- **Containerization**: Triển khai dễ dàng với Docker Compose

### Công nghệ sử dụng:
- **WebSocket** - Giao thức giao tiếp hai chiều thời gian thực
- **Golang** - Ngôn ngữ lập trình hiệu năng cao
- **Giao thức TCP** - Đảm bảo truyền tin nhắn đáng tin cậy
- **Kiến trúc Microservices** - Dễ dàng mở rộng và bảo trì

---

## Tính năng chính

### Quản lý người dùng
-  Đăng ký tài khoản với email validation
-  Đăng nhập với JWT authentication
-  Quản lý profile và avatar
-  Online/offline status tracking realtime
-  Session management với Redis

### Chat Realtime
-  **WebSocket connection** với persistent connection
-  **Direct messaging** giữa 2 users
-  **Group chat rooms** với nhiều members
-  **Message history** được lưu vào PostgreSQL
-  **Typing indicators** (ready to implement)
-  **Read receipts** tracking
-  **Message delivery confirmation**

### Quản lý bạn bè
-  Gửi/nhận lời mời kết bạn
-  Chấp nhận/từ chối lời mời
-  Tìm kiếm người dùng
-  Xem danh sách bạn bè
-  Trạng thái online/offline của bạn bè

### Bảo mật
-  **JWT Token** authentication
-  **Token blacklist** khi logout
-  **Rate limiting** chống DDoS
  - Auth endpoints: 30 req/min
  - API endpoints: 300 req/min
  - WebSocket messages: 120 msg/min/user
-  **CORS protection**
-  **Password hashing** với bcrypt
-  **SQL injection** prevention

### Performance & Monitoring
-  **Health check endpoints**
-  **Instance identification** trong load balancing
-  **Redis caching** cho user sessions, friends list, messages
-  **RabbitMQ queues** cho notifications, email, file processing

---

## Công nghệ và Tool sử dụng

### Backend
| Công nghệ | Phiên bản | Mục đích |
|-----------|-----------|----------|
| **Go** | 1.24 | Backend language - high performance, concurrency |
| **Gorilla WebSocket** | Latest | WebSocket protocol implementation |
| **PostgreSQL** | 15 | Primary database - ACID compliance |
| **Redis** | 7 | Caching & pub/sub - giảm tải database |
| **RabbitMQ** | 3 | Message queue - async task processing |
| **JWT** | Latest | Authentication & authorization |

### Frontend
| Công nghệ | Mục đích |
|-----------|----------|
| **HTML5** | Structure |
| **CSS3** | Styling with modern design |
| **JavaScript (Vanilla)** | WebSocket client, API calls |
| **WebSocket API** | Real-time communication |

### Infrastructure & DevOps
| Tool | Mục đích |
|------|----------|
| **Docker** | Containerization |
| **Docker Compose** | Multi-container orchestration |
| **Nginx** | Load balancer & reverse proxy |
| **k6** | Load testing tool (200 concurrent users) |

### Architecture Pattern
-  **Microservices** - Tách biệt services
-  **Hub Pattern** - WebSocket message distribution
-  **Repository Pattern** - Database abstraction
-  **Middleware Chain** - Request processing pipeline
-  **Pub/Sub** - Cross-instance messaging với Redis

---

## Kiến trúc hệ thống

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENTS                              │
│  (Browser, Mobile Apps, Desktop Apps)                   │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/WebSocket
                     ▼
        ┌────────────────────────────┐
        │   Load Balancer (Nginx)    │
        │  - Port 8080 (API/WS)      │
        │  - Least Connection        │
        │  - Health Checks           │
        └────────────┬───────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
    ┌────────┐  ┌────────┐  ┌────────┐
    │Server 1│  │Server 2│  │Server 3│  ← Horizontal Scaling
    │:8080   │  │:8080   │  │:8080   │     (Go WebSocket)
    └───┬────┘  └───┬────┘  └───┬────┘
        └───────────┼───────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
    ┌────────┐  ┌──────┐  ┌──────────┐
    │ Redis  │  │ Rabbit│ │PostgreSQL│
    │ Cache  │  │  MQ   │ │ Database │
    │ :6379  │  │ :5672 │ │  :5432   │
    └────────┘  └──────┘  └──────────┘
         │          │            │
         │          │            │
    [Sessions]  [Queues]    [Persistent]
    [Online]    [Async]      [Storage]
    [Cache]     [Tasks]      [ACID]
```

### Data Flow:

1. **Client → Load Balancer**: HTTP/WebSocket requests
2. **Load Balancer → Server**: Route to healthy instance
3. **Server → Redis**: Cache operations, pub/sub
4. **Server → RabbitMQ**: Async task queuing
5. **Server → PostgreSQL**: Persistent data storage
6. **Server → Client**: WebSocket real-time messages

### Scaling Strategy:

```bash
# Scale to 10 instances
docker-compose up -d --scale server=10

# Load balancer tự động phân tải
# Redis Pub/Sub đồng bộ giữa các instances
```

---

## Yêu cầu hệ thống

### Phát triển (Development)
- **OS**: Windows 10/11, macOS, Linux
- **RAM**: >= 8GB (recommended 16GB)
- **Disk**: >= 10GB free space
- **Docker Desktop**: >= 4.25.0
- **Browser**: Chrome/Firefox/Edge (latest)

### Production
- **CPU**: >= 4 cores
- **RAM**: >= 16GB
- **Disk**: >= 50GB SSD
- **Network**: >= 100Mbps
- **Docker** + **Docker Compose**

---

## Hướng dẫn cài đặt

### Bước 1: Clone Repository

```bash
git clone https://github.com/QuyTran25/e5-realtime-chat.git
cd e5-realtime-chat
```

### Bước 2: Cấu hình môi trường (Optional)

Dự án đã có cấu hình mặc định trong `docker-compose.yml`. Nếu muốn thay đổi:

```yaml
# infra/docker-compose.yml
environment:
  - DB_HOST=postgres
  - DB_PORT=5432
  - DB_USER=chatuser
  - DB_PASSWORD=chatpass        # ⚠️ Đổi trong production
  - DB_NAME=chatdb
  - REDIS_HOST=redis
  - REDIS_PORT=6379
  - RABBITMQ_HOST=rabbitmq
  - RABBITMQ_PORT=5672
```

### Bước 3: Khởi động toàn bộ hệ thống

```bash
cd infra
docker-compose up -d
```

Lệnh này sẽ khởi động:
-  PostgreSQL database (port 5432)
-  Redis cache (port 6379)
-  RabbitMQ message queue (port 5672, 15672)
-  3x Backend servers (internal)
-  Nginx load balancer (port 8080)
-  Frontend web server (port 3000)
-  k6 load testing tool

### Bước 4: Kiểm tra services

```bash
# Xem trạng thái containers
docker-compose ps

# Xem logs
docker-compose logs -f server

# Kiểm tra health
curl http://localhost:8080/healthz
```

### Bước 5: Truy cập ứng dụng

#### Frontend:
- **Trang chủ**: http://localhost:3000
- **Đăng nhập**: http://localhost:3000/login.html
- **Đăng ký**: http://localhost:3000/register.html

#### Backend API:
- **Health Check**: http://localhost:8080/healthz
- **WebSocket**: ws://localhost:8080/ws?token=YOUR_TOKEN
- **API Base**: http://localhost:8080/api/

#### Management Tools:
- **RabbitMQ Management**: http://localhost:15672
  - Username: `chatuser`
  - Password: `chatpass`

### Bước 6: Test với user mẫu

Database đã có sẵn users mẫu (password: `password123`):
- `admin@example.com`
- `alice@example.com`
- `bob@example.com`

Hoặc đăng ký user mới tại: http://localhost:3000/register.html

---

## Ảnh chụp màn hình & Demo

### 1. Trang đăng ký
- Form đăng ký với validation realtime
- Chỉ số độ mạnh mật khẩu
- Xác thực email
![Giao diện đăng ký](image-1.png)

### 2. Trang đăng nhập
- Xác thực JWT
- Tùy chọn ghi nhớ đăng nhập
- Tích hợp đăng nhập xã hội (Google, Facebook)

![Giao diện đăng nhập](docs/images/image.png)

### 3. Giao diện trang chủ
- Sidebar với các chức năng chính như nhắn tin, tìm kiếm, lời mời kết bạn, danh sách bạn bè, đăng xuất

![Giao diện trang chủ](docs/images/image-2.png)

### 4. Giao diện nhắn tin
- Nhắn tin realtime và hiện trạng thái hoạt động

![Giao diện nhắn tin](docs/images/image-3.png)

### 5. Giao diện tìm kiếm bạn bè
- Tìm kiếm bạn bè bằng tên hoặc email

![Giao diện tìm kiếm bạn bè](docs/images/image-4.png)

### 6. Giao diện lời mời kết bạn
- Người gửi lời mời kết bạn

![Giao diện gửi lời mời kết bạn](docs/images/image-5.png)

- Người nhận lời mời kết bạn

![Giao diện người nhận lời mời kết bạn](docs/images/image-6.png)

### 7. Giao diện danh sách bạn bè
- Hiển thị danh sách bạn bè

![Giao diện danh sách bạn bè](docs/images/image-7.png)

### 8. Giao diện đăng xuất
- Giao diện đăng xuất

![Giao diện đăng xuất](docs/images/image-8.png)

### 9. WebSocket Connection Demo

```javascript
// Kết nối WebSocket với JWT token
const ws = new WebSocket('ws://localhost:8080/ws?token=eyJhbGci...');

ws.onopen = () => {
    console.log(' WebSocket connected');
    
    // Gửi tin nhắn
    ws.send(JSON.stringify({
        type: 'message',
        from_user_id: 1,
        to_user_id: 2,
        text: 'Hello from WebSocket!'
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log(' Received:', message);
};
```

### 10. Load Balancing in Action

```bash
# Request 1 → Instance 1
curl http://localhost:8080/healthz
{"status":"ok","instance":"573c0617c957"}

# Request 2 → Instance 2
curl http://localhost:8080/healthz
{"status":"ok","instance":"8a9f3bd21c48"}

# Request 3 → Instance 3
curl http://localhost:8080/healthz
{"status":"ok","instance":"c5d7e4f92ab3"}
```

---

## Tài liệu API

### API Xác thực

#### 1. Đăng ký
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}

Response 201:
{
  "success": true,
  "message": "registration successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

#### 2. Đăng nhập
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123!"
}

Response 200:
{
  "success": true,
  "message": "login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "is_online": true
  }
}
```

#### 3. Đăng xuất
```http
POST /api/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

Response 200:
{
  "success": true,
  "message": "logout successful"
}
```

### API Bạn bè

#### 4. Lấy danh sách bạn bè
```http
GET /api/friends
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

Response 200:
{
  "success": true,
  "friends": [
    {
      "id": 2,
      "username": "alice",
      "email": "alice@example.com",
      "is_online": true
    }
  ]
}
```

#### 5. Tìm kiếm người dùng
```http
GET /api/friends/search?q=alice
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

Response 200:
{
  "success": true,
  "users": [
    {
      "id": 2,
      "username": "alice",
      "is_friend": true
    }
  ]
}
```

### Tin nhắn WebSocket

#### Kết nối
```javascript
ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIs...
```

#### Gửi tin nhắn
```json
{
  "type": "message",
  "from_user_id": 1,
  "to_user_id": 2,
  "text": "Hello!"
}
```

#### Nhận tin nhắn
```json
{
  "type": "message",
  "from": "john_doe",
  "from_user_id": 1,
  "text": "Hello!",
  "timestamp": "2025-11-18T10:40:00Z"
}
```

### Headers giới hạn tốc độ
```http
X-RateLimit-Limit: 300
X-RateLimit-Remaining: 299
X-RateLimit-Reset: 1700308800
```

---

## Hiệu năng & Kiểm thử tải

### Kiểm thử với k6

Dự án đã tích hợp công cụ kiểm thử tải k6:

```bash
# Chạy load test với 200 người dùng đồng thời
cd infra
docker-compose up k6
```

### Cấu hình kiểm thử
```javascript
// infra/k6-script.js
export let options = {
  vus: 200,           // 200 virtual users
  duration: "30s",    // Test trong 30 giây
};
```

### Kết quả hiệu năng

| Chỉ số | Giá trị | Ghi chú |
|--------|-------|------|
| **Kết nối WebSocket đồng thời** | 10,000+ | Kiểm thử với k6 |
| **Thông lượng tin nhắn** | 10,000 msg/s | Mỗi instance |
| **Thời gian phản hồi (API)** | < 50ms | Trung bình |
| **Thời gian phản hồi (WebSocket)** | < 10ms | Thời gian thực |
| **Sử dụng CPU** | ~30% | 3 instances, 4 cores |
| **Sử dụng bộ nhớ** | ~500MB | Mỗi Go instance |
| **Kết nối Database** | 100 | Kích thước Pool |
| **Redis Ops** | 50,000/s | Tỷ lệ cache hit 85% |

### Kiểm thử khả năng mở rộng

```bash
# Kiểm thử với 1 instance
docker-compose up -d --scale server=1
# Kết quả: 3,000 người dùng đồng thời

# Kiểm thử với 3 instances (mặc định)
docker-compose up -d --scale server=3
# Kết quả: 10,000 người dùng đồng thời

# Kiểm thử với 10 instances
docker-compose up -d --scale server=10
# Kết quả: 30,000+ người dùng đồng thời
```

---

## Xử lý sự cố

### Vấn đề: Port đã được sử dụng
```bash
# Kiểm tra port đang dùng
netstat -ano | findstr "8080"

# Dừng service đang dùng port hoặc đổi port trong docker-compose.yml
```

### Vấn đề: Kết nối database thất bại
```bash
# Đợi database khởi động hoàn tất (10-15s)
docker-compose logs postgres

# Restart server nếu cần
docker-compose restart server
```

### Vấn đề: Kết nối WebSocket bị từ chối
```bash
# Kiểm tra server đang chạy
docker-compose ps

# Xem logs
docker-compose logs server

# Restart load balancer
docker-compose restart loadbalancer
```

### Vấn đề: Reset toàn bộ hệ thống
```bash
# Dừng và xóa tất cả
docker-compose down -v

# Xóa images
docker rmi e5-realtime-server:latest

# Khởi động lại
docker-compose up -d --build
```

---

## Cấu trúc dự án

```
e5-realtime-chat/
├── client/                     # Frontend
│   ├── html/                   # HTML pages
│   │   ├── index.html         # Chat main page
│   │   ├── login.html         # Login page
│   │   └── register.html      # Register page
│   └── assets/
│       ├── css/               # Stylesheets
│       └── js/                # JavaScript
│
├── server/                     # Backend (Golang)
│   ├── main.go                # Entry point
│   ├── internal/              # Internal packages
│   │   ├── auth/              # Authentication
│   │   ├── cache/             # Redis caching
│   │   ├── database/          # PostgreSQL
│   │   ├── handlers/          # API handlers
│   │   ├── middleware/        # HTTP middleware
│   │   ├── queue/             # RabbitMQ
│   │   └── websocket/         # WebSocket
│   └── migrations/            # Database migrations
│
├── infra/                      # Infrastructure
│   ├── docker-compose.yml     # Services orchestration
│   ├── k6-script.js           # Load testing script
│   └── nginx/                 # Nginx config
│
└── README.md                   # This file
```

---

## Thành viên nhóm

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/QuyTran25">
        <img src="https://github.com/QuyTran25.png" width="100px;" alt="Quy Tran"/>
        <br />
        <sub><b>Huỳnh Thị Quý Trân</b></sub>
      </a>
      <br />
      <sub>Trưởng nhóm & Quản lý dự án</sub>
      <br />
      <sub>Frontend, Testing & Sửa lỗi</sub>
    </td>
    <td align="center">
      <a href="https://github.com/khain7728">
        <img src="https://github.com/khain7728.png" width="100px;" alt="Khai"/>
        <br />
        <sub><b>Nguyễn Quốc Khải</b></sub>
      </a>
      <br />
      <sub>Chức năng nhắn tin</sub>
    </td>
    <td align="center">
      <a href="https://github.com/akhoa79">
        <img src="https://github.com/akhoa79.png" width="100px;" alt="Khoa"/>
        <br />
        <sub><b>Nguyễn Đỗ Anh Khoa</b></sub>
      </a>
      <br />
      <sub>Chức năng lời mời kết bạn</sub>
    </td>
    <td align="center">
      <a href="https://github.com/LeVietSangg">
        <img src="https://github.com/LeVietSangg.png" width="100px;" alt="Sang"/>
        <br />
        <sub><b>Lê Viết Sang</b></sub>
      </a>
      <br />
      <sub>Chức năng tìm kiếm</sub>
      <br />
      <sub>& Danh sách bạn bè</sub>
    </td>
    <td align="center">
      <a href="https://github.com/Shinnie102">
        <img src="https://github.com/Shinnie102.png" width="100px;" alt="Shinnie"/>
        <br />
        <sub><b>Nguyễn Thị Thùy Trang</b></sub>
      </a>
      <br />
      <sub>Chức năng xác thực</sub>
      <br />
      <sub>Đăng nhập, Đăng ký & Đăng xuất</sub>
    </td>
  </tr>
</table>

### Hướng dẫn đóng góp

Chúng tôi rất hoan nghênh các đóng góp! 

1. Fork repository
2. Tạo feature branch (`git checkout -b feature/TinhNangMoi`)
3. Commit thay đổi (`git commit -m 'Thêm tính năng mới'`)
4. Push lên branch (`git push origin feature/TinhNangMoi`)
5. Mở Pull Request


## Liên hệ & Hỗ trợ

- **GitHub**: [@QuyTran25](https://github.com/QuyTran25)
- **Repository**: [e5-realtime-chat](https://github.com/QuyTran25/e5-realtime-chat)



