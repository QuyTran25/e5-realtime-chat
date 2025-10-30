-- ============================================
-- E5 Realtime Chat Database Schema
-- ============================================

-- Enable UUID extension (optional, for better IDs)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Table: users
-- Lưu thông tin người dùng
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    is_online BOOLEAN DEFAULT false,
    last_seen_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index cho tìm kiếm user nhanh
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_online ON users(is_online);

-- ============================================
-- Table: rooms
-- Lưu thông tin phòng chat (cho tính năng room/channel)
-- ============================================
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    room_name VARCHAR(100) UNIQUE NOT NULL,
    room_type VARCHAR(20) DEFAULT 'public', -- public, private, group
    description TEXT,
    created_by INT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rooms_type ON rooms(room_type);

-- ============================================
-- Table: messages
-- Lưu tất cả tin nhắn
-- ============================================
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    message_type VARCHAR(20) DEFAULT 'message', -- message, join, leave, data
    from_user_id INT REFERENCES users(id) ON DELETE CASCADE,
    to_user_id INT REFERENCES users(id) ON DELETE CASCADE, -- NULL = public message
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE, -- NULL = direct message
    message_text TEXT NOT NULL,
    message_value DECIMAL(10, 2), -- Cho type "data" (dashboard)
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes cho query nhanh
CREATE INDEX idx_messages_from_user ON messages(from_user_id, created_at DESC);
CREATE INDEX idx_messages_to_user ON messages(to_user_id, created_at DESC);
CREATE INDEX idx_messages_room ON messages(room_id, created_at DESC);
CREATE INDEX idx_messages_type ON messages(message_type);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
-- Index cho direct messages (tin nhắn riêng tư giữa 2 người)
CREATE INDEX idx_messages_direct_conversation ON messages(from_user_id, to_user_id, created_at DESC) WHERE to_user_id IS NOT NULL;
CREATE INDEX idx_messages_direct_reverse ON messages(to_user_id, from_user_id, created_at DESC) WHERE to_user_id IS NOT NULL;

-- ============================================
-- Table: friendships
-- Quản lý quan hệ bạn bè
-- ============================================
CREATE TABLE IF NOT EXISTS friendships (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, blocked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, friend_id),
    CHECK (user_id != friend_id)
);

CREATE INDEX idx_friendships_user ON friendships(user_id, status);
CREATE INDEX idx_friendships_friend ON friendships(friend_id, status);
CREATE INDEX idx_friendships_status_created ON friendships(status, created_at DESC);
CREATE INDEX idx_friendships_both_users ON friendships(user_id, friend_id, status);

-- ============================================
-- Table: user_sessions
-- Quản lý sessions và JWT tokens
-- ============================================
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    refresh_token VARCHAR(500),
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(token);
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);

-- ============================================
-- Table: room_members
-- Quản lý thành viên trong room
-- ============================================
CREATE TABLE IF NOT EXISTS room_members (
    id SERIAL PRIMARY KEY,
    room_id INT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- admin, moderator, member
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(room_id, user_id)
);

CREATE INDEX idx_room_members_room ON room_members(room_id);
CREATE INDEX idx_room_members_user ON room_members(user_id);

-- ============================================
-- Functions and Triggers
-- ============================================

-- Function: Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Auto update updated_at cho users
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger: Auto update updated_at cho friendships
CREATE TRIGGER update_friendships_updated_at
    BEFORE UPDATE ON friendships
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Functions: Friendship & Message Validation
-- ============================================

-- Function: Kiểm tra 2 user có phải bạn bè không
CREATE OR REPLACE FUNCTION are_friends(user1_id INT, user2_id INT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM friendships
        WHERE status = 'accepted'
        AND (
            (user_id = user1_id AND friend_id = user2_id)
            OR (user_id = user2_id AND friend_id = user1_id)
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Function: Kiểm tra có thể gửi tin nhắn trực tiếp không
CREATE OR REPLACE FUNCTION can_send_direct_message(sender_id INT, receiver_id INT)
RETURNS BOOLEAN AS $$
BEGIN
    -- Không thể gửi cho chính mình
    IF sender_id = receiver_id THEN
        RETURN FALSE;
    END IF;
    
    -- Admin có thể gửi cho bất kỳ ai (optional)
    -- IF EXISTS (SELECT 1 FROM users WHERE id = sender_id AND username = 'admin') THEN
    --     RETURN TRUE;
    -- END IF;
    
    -- Phải là bạn bè mới được gửi tin nhắn trực tiếp
    RETURN are_friends(sender_id, receiver_id);
END;
$$ LANGUAGE plpgsql;

-- Trigger Function: Validate direct message
CREATE OR REPLACE FUNCTION validate_direct_message()
RETURNS TRIGGER AS $$
BEGIN
    -- Nếu là tin nhắn trực tiếp (có to_user_id và không có room_id)
    IF NEW.to_user_id IS NOT NULL AND NEW.room_id IS NULL THEN
        -- Kiểm tra có phải bạn bè không
        IF NOT can_send_direct_message(NEW.from_user_id, NEW.to_user_id) THEN
            RAISE EXCEPTION 'Cannot send direct message: users must be friends first. User % tried to send to user %', NEW.from_user_id, NEW.to_user_id;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Check friendship before inserting direct message
CREATE TRIGGER check_direct_message_friendship
    BEFORE INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION validate_direct_message();

-- Function: Get friendship status between 2 users
CREATE OR REPLACE FUNCTION get_friendship_status(user1_id INT, user2_id INT)
RETURNS VARCHAR(20) AS $$
DECLARE
    friendship_status VARCHAR(20);
BEGIN
    SELECT status INTO friendship_status
    FROM friendships
    WHERE (user_id = user1_id AND friend_id = user2_id)
       OR (user_id = user2_id AND friend_id = user1_id)
    LIMIT 1;
    
    IF friendship_status IS NULL THEN
        RETURN 'none';
    END IF;
    
    RETURN friendship_status;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Insert Default Data
-- ============================================

-- Default room: General
INSERT INTO rooms (room_name, room_type, description) 
VALUES ('general', 'public', 'General chat room for everyone')
ON CONFLICT (room_name) DO NOTHING;

-- Sample users (for testing only)
INSERT INTO users (username, email, password_hash, avatar_url) VALUES
    ('admin', 'admin@chat.com', '$2a$10$gRvWmIQ7jfGOsb9NiNHA7uOA3tKCJZc.LSTqMHjGSv86JW0YwHjYe', 'https://i.pravatar.cc/150?img=1'),
    ('alice', 'alice@chat.com', '$2a$10$gRvWmIQ7jfGOsb9NiNHA7uOA3tKCJZc.LSTqMHjGSv86JW0YwHjYe', 'https://i.pravatar.cc/150?img=2'),
    ('bob', 'bob@chat.com', '$2a$10$gRvWmIQ7jfGOsb9NiNHA7uOA3tKCJZc.LSTqMHjGSv86JW0YwHjYe', 'https://i.pravatar.cc/150?img=3')
ON CONFLICT (username) DO NOTHING;

-- Sample friendships
INSERT INTO friendships (user_id, friend_id, status) VALUES
    (2, 3, 'accepted')
ON CONFLICT (user_id, friend_id) DO NOTHING;

-- Add users to general room
INSERT INTO room_members (room_id, user_id) 
SELECT 1, id FROM users WHERE username IN ('admin', 'alice', 'bob')
ON CONFLICT DO NOTHING;

-- ============================================
-- Views (Optional - for convenience)
-- ============================================

-- View: Recent messages with user info
CREATE OR REPLACE VIEW v_recent_messages AS
SELECT 
    m.id,
    m.message_type,
    m.message_text,
    m.message_value,
    m.created_at,
    fu.id AS from_user_id,
    fu.username AS from_username,
    fu.avatar_url AS from_avatar,
    tu.id AS to_user_id,
    tu.username AS to_username,
    r.id AS room_id,
    r.room_name
FROM messages m
LEFT JOIN users fu ON m.from_user_id = fu.id
LEFT JOIN users tu ON m.to_user_id = tu.id
LEFT JOIN rooms r ON m.room_id = r.id
ORDER BY m.created_at DESC;

-- View: User friends list
CREATE OR REPLACE VIEW v_user_friends AS
SELECT 
    f.user_id,
    u.id AS friend_id,
    u.username AS friend_username,
    u.avatar_url AS friend_avatar,
    u.is_online AS friend_online,
    u.last_seen_at AS friend_last_seen,
    f.status,
    f.created_at AS friends_since
FROM friendships f
JOIN users u ON f.friend_id = u.id
WHERE f.status = 'accepted';

-- ============================================
-- Success Message
-- ============================================
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE '✅ Database schema created successfully!';
    RAISE NOTICE '============================================';
    RAISE NOTICE '📊 Tables: users, rooms, messages, friendships, user_sessions, room_members';
    RAISE NOTICE '� Auth: JWT token support with user_sessions';
    RAISE NOTICE '�👥 Friendship: Must be friends to send direct messages';
    RAISE NOTICE '💬 Messages: Support both room messages and direct messages';
    RAISE NOTICE '';
    RAISE NOTICE '📝 Sample Data:';
    RAISE NOTICE '   - Users: admin, alice, bob';
    RAISE NOTICE '   - Password: password123 (bcrypt hash)';
    RAISE NOTICE '   - Room: general (public)';
    RAISE NOTICE '   - Friends: alice ↔ bob (accepted)';
    RAISE NOTICE '';
    RAISE NOTICE '🔍 Key Functions:';
    RAISE NOTICE '   - are_friends(user1_id, user2_id) → Check if users are friends';
    RAISE NOTICE '   - can_send_direct_message(sender_id, receiver_id) → Validate message permission';
    RAISE NOTICE '   - get_friendship_status(user1_id, user2_id) → Get friendship status';
    RAISE NOTICE '';
    RAISE NOTICE '⚡ Triggers:';
    RAISE NOTICE '   - Messages: Auto-validate friendship before sending direct messages';
    RAISE NOTICE '   - Users/Friendships: Auto-update updated_at timestamp';
    RAISE NOTICE '============================================';
END $$;

-- Xóa dữ liệu cũ (nếu có)
DELETE FROM messages;
DELETE FROM friendships;
DELETE FROM users WHERE username IN ('nguyenvana', 'tranthib');

-- Thêm 2 người dùng (không chỉ định id, để SERIAL tự tăng)
INSERT INTO users (username, email, password_hash, avatar_url, created_at, updated_at, last_seen_at, is_online) VALUES
('nguyenvana', 'vana@email.com', '$2a$10$gRvWmIQ7jfGOsb9NiNHA7uOA3tKCJZc.LSTqMHjGSv86JW0YwHjYe', 'https://i.pravatar.cc/150?img=1', '2024-01-15 10:00:00', '2024-01-15 10:00:00', '2024-10-30 14:30:00', true),
('tranthib', 'thib@email.com', '$2a$10$gRvWmIQ7jfGOsb9NiNHA7uOA3tKCJZc.LSTqMHjGSv86JW0YwHjYe', 'https://i.pravatar.cc/150?img=2', '2024-02-20 11:30:00', '2024-02-20 11:30:00', '2024-10-30 15:00:00', true);

-- Thêm quan hệ bạn bè (2 người đã kết bạn)
-- Lấy id của 2 user vừa tạo
INSERT INTO friendships (user_id, friend_id, status, created_at, updated_at) 
SELECT u1.id, u2.id, 'accepted', '2024-03-10 09:00:00', '2024-03-10 09:30:00'
FROM users u1, users u2
WHERE u1.username = 'nguyenvana' AND u2.username = 'tranthib';

-- Không thêm tin nhắn nào giữa 2 người
-- Bảng messages để trống cho 2 người này

-- Kiểm tra dữ liệu
SELECT 'Users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'Friendships', COUNT(*) FROM friendships
UNION ALL
SELECT 'Messages', COUNT(*) FROM messages;

-- Xem chi tiết người dùng
SELECT id, username, email, avatar_url, is_online, last_seen_at FROM users 
WHERE username IN ('nguyenvana', 'tranthib');

-- Xem quan hệ bạn bè
SELECT 
    f.id as friendship_id,
    u1.username as user_1,
    u2.username as user_2,
    f.status,
    f.created_at
FROM friendships f
JOIN users u1 ON f.user_id = u1.id
JOIN users u2 ON f.friend_id = u2.id
WHERE u1.username IN ('nguyenvana', 'tranthib') 
   OR u2.username IN ('nguyenvana', 'tranthib')
ORDER BY f.created_at;