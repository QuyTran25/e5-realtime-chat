// ========================
// 🔐 WebSocket Client với Authentication
// ========================

let ws = null;
let reconnectInterval = null;
let currentUser = null;

// Khởi tạo kết nối WebSocket
function initWebSocket() {
    // Lấy token từ localStorage hoặc sessionStorage
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    
    if (!userData) {
        console.error('❌ No user session found. Redirecting to login...');
        window.location.href = 'login.html';
        return;
    }

    try {
        currentUser = JSON.parse(userData);
    } catch (e) {
        console.error('❌ Invalid user data. Redirecting to login...');
        window.location.href = 'login.html';
        return;
    }

    if (!currentUser.token) {
        console.error('❌ No token found. Redirecting to login...');
        window.location.href = 'login.html';
        return;
    }

    // Kết nối WebSocket với token
    const wsUrl = `ws://localhost:8080/ws?token=${currentUser.token}`;
    console.log('🔌 Connecting to WebSocket:', wsUrl);

    ws = new WebSocket(wsUrl);

    ws.onopen = function() {
        console.log('✅ WebSocket connected');
        if (reconnectInterval) {
            clearInterval(reconnectInterval);
            reconnectInterval = null;
        }

        // Gửi tin nhắn join
        sendMessage({
            type: 'join',
            user: currentUser.name || currentUser.email
        });
    };

    ws.onmessage = function(event) {
        console.log('📨 Message received:', event.data);
        try {
            const message = JSON.parse(event.data);
            handleIncomingMessage(message);
        } catch (e) {
            console.error('❌ Failed to parse message:', e);
        }
    };

    ws.onerror = function(error) {
        console.error('❌ WebSocket error:', error);
    };

    ws.onclose = function(event) {
        console.log('🔌 WebSocket disconnected. Code:', event.code, 'Reason:', event.reason);
        
        // Nếu đóng vì unauthorized (1008), redirect về login
        if (event.code === 1008 || event.reason.includes('Unauthorized')) {
            console.log('🚪 Session expired. Redirecting to login...');
            localStorage.removeItem('user');
            sessionStorage.removeItem('user');
            window.location.href = 'login.html';
            return;
        }

        // Tự động reconnect sau 5 giây
        if (!reconnectInterval) {
            console.log('🔄 Will reconnect in 5 seconds...');
            reconnectInterval = setTimeout(() => {
                console.log('🔄 Reconnecting...');
                initWebSocket();
            }, 5000);
        }
    };
}

// Gửi tin nhắn qua WebSocket
function sendMessage(messageObj) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        console.error('❌ WebSocket is not connected');
        return;
    }

    try {
        const jsonString = JSON.stringify(messageObj);
        ws.send(jsonString);
        console.log('📤 Message sent:', jsonString);
    } catch (e) {
        console.error('❌ Failed to send message:', e);
    }
}

// Xử lý tin nhắn nhận được
function handleIncomingMessage(message) {
    const chatMessages = document.querySelector('.chat-messages');
    if (!chatMessages) return;

    const messageDiv = document.createElement('div');
    
    switch (message.type) {
        case 'message':
            messageDiv.className = message.from === currentUser.name ? 'message sent' : 'message received';
            messageDiv.innerHTML = `
                <div class="message-content">
                    <div class="message-header">
                        <span class="message-sender">${message.from}</span>
                        <span class="message-time">${new Date().toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })}</span>
                    </div>
                    <div class="message-text">${escapeHtml(message.text)}</div>
                </div>
            `;
            break;

        case 'join':
            messageDiv.className = 'message system';
            messageDiv.textContent = `✅ ${message.user} đã tham gia`;
            break;

        case 'leave':
            messageDiv.className = 'message system';
            messageDiv.textContent = `👋 ${message.user} đã rời đi`;
            break;

        case 'data':
            messageDiv.className = 'message system';
            messageDiv.textContent = `📊 Data: ${message.value}`;
            break;

        default:
            console.log('Unknown message type:', message.type);
            return;
    }

    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Escape HTML để tránh XSS
function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, m => map[m]);
}

// Đăng xuất
function logout() {
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    
    if (userData) {
        try {
            const user = JSON.parse(userData);
            
            // Gọi API logout
            fetch('http://localhost:8080/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${user.token}`
                }
            })
            .then(response => response.json())
            .then(data => {
                console.log('✅ Logout successful:', data);
            })
            .catch(error => {
                console.error('❌ Logout error:', error);
            })
            .finally(() => {
                // Đóng WebSocket
                if (ws) {
                    sendMessage({ type: 'leave', user: user.name || user.email });
                    ws.close();
                }

                // Xóa session
                localStorage.removeItem('user');
                sessionStorage.removeItem('user');

                // Redirect về login
                window.location.href = 'login.html';
            });
        } catch (e) {
            console.error('❌ Logout error:', e);
            // Vẫn xóa session và redirect
            localStorage.removeItem('user');
            sessionStorage.removeItem('user');
            window.location.href = 'login.html';
        }
    }
}

// Khởi tạo khi trang load
document.addEventListener('DOMContentLoaded', function() {
    // Kiểm tra đăng nhập
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) {
        window.location.href = 'login.html';
        return;
    }

    try {
        currentUser = JSON.parse(userData);
        
        // Hiển thị thông tin user
        const userNameElement = document.querySelector('#currentUserName');
        if (userNameElement) {
            userNameElement.textContent = currentUser.name || currentUser.email;
        }

        const userAvatarElement = document.querySelector('#currentUserAvatar');
        if (userAvatarElement && currentUser.avatar) {
            userAvatarElement.src = currentUser.avatar;
        }

    } catch (e) {
        console.error('❌ Invalid user data');
        window.location.href = 'login.html';
        return;
    }

    // Khởi tạo WebSocket
    initWebSocket();

    // Xử lý gửi tin nhắn
    const chatInput = document.querySelector('#chatInput');
    const sendButton = document.querySelector('#sendButton');

    if (sendButton && chatInput) {
        sendButton.addEventListener('click', function() {
            const text = chatInput.value.trim();
            if (text) {
                sendMessage({
                    type: 'message',
                    from: currentUser.name || currentUser.email,
                    text: text
                });
                chatInput.value = '';
            }
        });

        chatInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendButton.click();
            }
        });
    }

    // Xử lý nút logout
    const logoutButton = document.querySelector('#logoutButton');
    if (logoutButton) {
        logoutButton.addEventListener('click', function(e) {
            e.preventDefault();
            
            // Hiển thị modal xác nhận
            const confirmLogout = confirm('Bạn có chắc muốn đăng xuất?');
            if (confirmLogout) {
                logout();
            }
        });
    }
});

// Đóng WebSocket khi đóng trang
window.addEventListener('beforeunload', function() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
    }
});
