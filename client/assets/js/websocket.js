// ========================
// 🔐 WebSocket Client với Authentication
// ========================

let ws = null;
let reconnectInterval = null;
let heartbeatInterval = null;
let currentUser = null;
let activeConversation = null; // Store current conversation user
let unreadMessages = {}; // Track unread messages per user: {userId: count}

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
        console.log('👤 Current user:', currentUser);
        
        // Normalize user ID field (backend may return user_id, userId, or id)
        if (!currentUser.id && currentUser.user_id) {
            currentUser.id = currentUser.user_id;
        }
        if (!currentUser.id && currentUser.userId) {
            currentUser.id = currentUser.userId;
        }
        
        console.log('👤 Normalized user ID:', currentUser.id);
        
        // ✅ AUTO FIX: If no ID found, clear storage and force re-login
        if (!currentUser.id) {
            console.error('❌ User data missing ID. Clearing old data...');
            localStorage.removeItem('user');
            sessionStorage.removeItem('user');
            alert('Dữ liệu đăng nhập cũ không hợp lệ. Vui lòng đăng nhập lại!');
            window.location.href = 'login.html';
            return;
        }
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

        // Start heartbeat - send every 20 seconds
        startHeartbeat();
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
        
        // Stop heartbeat
        stopHeartbeat();
        
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
    console.log('📥 Incoming message:', message);
    
    const chatMessages = document.querySelector('.chat-messages');

    // Handle user status changes
    if (message.type === 'user_status') {
        console.log('👤 User status changed:', message.user_id, message.username, message.is_online ? 'ONLINE' : 'OFFLINE');
        updateUserOnlineStatus(message.user_id, message.is_online);
        return;
    }

    // Handle heartbeat acknowledgment
    if (message.type === 'heartbeat_ack') {
        console.log('💚 Heartbeat acknowledged');
        return;
    }

    // Handle incoming message
    if (message.type === 'message') {
        const isFromMe = message.from_user_id === currentUser.id;
        const isToMe = message.to_user_id === currentUser.id;

        console.log('🔍 Message received:', {
            isFromMe,
            isToMe,
            myId: currentUser.id,
            fromUserId: message.from_user_id,
            toUserId: message.to_user_id,
            activeConversationId: activeConversation?.id
        });

        // Check if we should display this message
        let shouldDisplay = false;
        
        if (activeConversation) {
            const isChatWithActiveUser = 
                (isFromMe && message.to_user_id === activeConversation.id) || // I sent to active user
                (isToMe && message.from_user_id === activeConversation.id);   // Active user sent to me
            
            console.log('🔍 isChatWithActiveUser:', isChatWithActiveUser);
            shouldDisplay = isChatWithActiveUser;
            
            if (!isChatWithActiveUser && isToMe) {
                // Message for different conversation - update unread count
                const senderId = message.from_user_id;
                unreadMessages[senderId] = (unreadMessages[senderId] || 0) + 1;
                console.log('🔔 Unread count updated:', unreadMessages);
                updateNotificationBadges();
            }
        } else {
            // No active conversation - only display if it's my message
            shouldDisplay = isFromMe;
            
            if (!isFromMe && isToMe) {
                // Incoming message but no active conversation - update unread
                const senderId = message.from_user_id;
                unreadMessages[senderId] = (unreadMessages[senderId] || 0) + 1;
                console.log('🔔 Unread count updated:', unreadMessages);
                updateNotificationBadges();
            }
        }

        // Update conversation list
        if (window.updateConversationList) {
            window.updateConversationList();
        }

        // If we shouldn't display, stop here
        if (!shouldDisplay) {
            console.log('⏭️ Message not displayed (not for current conversation)');
            return;
        }

        // Display the message
        if (!chatMessages) {
            console.warn('⚠️ chatMessages element not found');
            return;
        }

        const messageDiv = document.createElement('div');
        const isMyMessage = message.from_user_id === currentUser.id;
        messageDiv.className = isMyMessage ? 'message own' : 'message';
        console.log('💬 Displaying message. isMyMessage:', isMyMessage);
        
        if (isMyMessage) {
            messageDiv.innerHTML = `
                <div class="message-content">
                    <div class="message-bubble">${escapeHtml(message.text)}</div>
                    <div class="message-time">${new Date().toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })}</div>
                </div>
            `;
        } else {
            messageDiv.innerHTML = `
                <img src="${activeConversation?.avatar || '../assets/images/default-avatar.svg'}" alt="${message.from}" class="message-avatar">
                <div class="message-content">
                    <div class="message-bubble">${escapeHtml(message.text)}</div>
                    <div class="message-time">${new Date().toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })}</div>
                </div>
            `;
        }
        
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
        console.log('✅ Message displayed successfully');
        return;
    }

    // Handle other message types (join, leave, etc.)
    if (!chatMessages) return;

    const messageDiv = document.createElement('div');
    
    switch (message.type) {
        case 'join':
            messageDiv.className = 'message system';
            // messageDiv.textContent = `✅ ${message.user} đã tham gia`;
            break;

        case 'leave':
            messageDiv.className = 'message system';
            // messageDiv.textContent = `👋 ${message.user} đã rời đi`;
            break;

        case 'data':
            messageDiv.className = 'message system';
            // messageDiv.textContent = `📊 Data: ${message.value}`;
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
            
            // 1. First, call API logout and wait for response
            fetch('http://localhost:8080/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${user.token}`
                }
            })
            .then(response => response.json())
            .then(data => {
                console.log('✅ Logout API successful:', data);
            })
            .catch(error => {
                console.error('❌ Logout API error:', error);
            })
            .finally(() => {
                // 2. Stop heartbeat
                stopHeartbeat();
                
                // 3. Send leave message
                if (ws && ws.readyState === WebSocket.OPEN) {
                    sendMessage({ type: 'leave', user: user.name || user.email });
                }
                
                // 4. Close WebSocket (this will trigger SetUserOffline in backend)
                if (ws) {
                    ws.close();
                }

                // 5. Clear session storage
                localStorage.removeItem('user');
                sessionStorage.removeItem('user');

                // 6. Redirect to login after a short delay to ensure WebSocket closes
                setTimeout(() => {
                    window.location.href = 'login.html';
                }, 300);
            });
        } catch (e) {
            console.error('❌ Logout error:', e);
            // Vẫn xóa session và redirect
            stopHeartbeat();
            if (ws) ws.close();
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
    const chatInput = document.querySelector('#messageInput');
    const sendButton = document.querySelector('#sendBtn');

    if (sendButton && chatInput) {
        sendButton.addEventListener('click', function() {
            const text = chatInput.value.trim();
            console.log('🔍 Send button clicked. Text:', text, 'activeConversation:', activeConversation);
            if (text && activeConversation) {
                sendMessage({
                    type: 'message',
                    from: currentUser.name || currentUser.email,
                    from_user_id: currentUser.id,
                    to_user_id: activeConversation.id,
                    text: text
                });
                chatInput.value = '';
            } else if (!activeConversation) {
                console.error('❌ No active conversation. Please select a user to chat with.');
                alert('Vui lòng chọn người để nhắn tin!');
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

// Set active conversation
function setActiveConversation(user) {
    console.log('✅ Setting active conversation:', user);
    activeConversation = user;
    
    // 🔔 Clear unread count for this conversation
    if (unreadMessages[user.id]) {
        delete unreadMessages[user.id];
        console.log('✅ Cleared unread count for user:', user.id);
        updateNotificationBadges();
    }
    
    loadChatHistory(user.id);
}

// Load chat history from API
async function loadChatHistory(userId) {
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) return;

    try {
        const user = JSON.parse(userData);
        const response = await fetch(`http://localhost:8080/api/messages/history?user_id=${userId}&limit=50`, {
            headers: {
                'Authorization': `Bearer ${user.token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to load chat history');
        }

        const data = await response.json();
        if (data.success && data.messages) {
            displayChatHistory(data.messages);
        }
    } catch (error) {
        console.error('❌ Error loading chat history:', error);
    }
}

// Display chat history
function displayChatHistory(messages) {
    const chatMessages = document.querySelector('.chat-messages');
    if (!chatMessages) return;

    chatMessages.innerHTML = ''; // Clear existing messages

    messages.forEach(msg => {
        const messageDiv = document.createElement('div');
        const isMyMessage = msg.from_user_id === currentUser.id;
        messageDiv.className = isMyMessage ? 'message own' : 'message';

        const messageTime = new Date(msg.created_at).toLocaleTimeString('vi-VN', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });

        if (isMyMessage) {
            messageDiv.innerHTML = `
                <div class="message-content">
                    <div class="message-bubble">${escapeHtml(msg.text)}</div>
                    <div class="message-time">${messageTime}</div>
                </div>
            `;
        } else {
            messageDiv.innerHTML = `
                <img src="${activeConversation?.avatar || '../assets/images/default-avatar.svg'}" alt="${msg.from_username}" class="message-avatar">
                <div class="message-content">
                    <div class="message-bubble">${escapeHtml(msg.text)}</div>
                    <div class="message-time">${messageTime}</div>
                </div>
            `;
        }

        chatMessages.appendChild(messageDiv);
    });

    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Export functions for use in other files
window.setActiveConversation = setActiveConversation;
window.loadChatHistory = loadChatHistory;
window.getUnreadCount = function(userId) {
    return unreadMessages[userId] || 0;
};
window.getTotalUnreadCount = function() {
    return Object.values(unreadMessages).reduce((sum, count) => sum + count, 0);
};

// ========================
// 🔔 Notification Badge Management
// ========================

function updateNotificationBadges() {
    const totalUnread = Object.values(unreadMessages).reduce((sum, count) => sum + count, 0);
    console.log('🔔 Updating notification badges. Total unread:', totalUnread);
    
    // Update menu notification dot
    const messagesNavItem = document.querySelector('.nav-item[href="index.html"]');
    if (messagesNavItem) {
        let notificationDot = messagesNavItem.querySelector('.notification-dot');
        
        if (totalUnread > 0) {
            // Show notification dot
            if (!notificationDot) {
                notificationDot = document.createElement('span');
                notificationDot.className = 'notification-dot';
                messagesNavItem.appendChild(notificationDot);
            }
        } else {
            // Remove notification dot
            if (notificationDot) {
                notificationDot.remove();
            }
        }
    }
    
    // Update conversation list badges
    if (window.updateConversationList) {
        window.updateConversationList();
    }
}

// ========================
// 💓 Heartbeat Management
// ========================

function startHeartbeat() {
    // Clear any existing heartbeat
    stopHeartbeat();
    
    // Send heartbeat every 20 seconds
    heartbeatInterval = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            sendMessage({ type: 'heartbeat' });
            console.log('💓 Heartbeat sent');
        }
    }, 20000);
    
    console.log('✅ Heartbeat started (every 20s)');
}

function stopHeartbeat() {
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
        console.log('❌ Heartbeat stopped');
    }
}

// ========================
// 👤 User Status Management
// ========================

function updateUserOnlineStatus(userId, isOnline) {
    console.log(`🔄 Updating UI for user ${userId} - ${isOnline ? 'ONLINE' : 'OFFLINE'}`);
    
    // Update in friends list
    const friendItems = document.querySelectorAll('.friend-item');
    console.log(`🔍 Found ${friendItems.length} friend items`);
    friendItems.forEach(item => {
        const friendId = item.getAttribute('data-user-id');
        if (friendId && parseInt(friendId) === userId) {
            console.log(`✅ Found matching friend item for user ${userId}`);
            const statusIndicator = item.querySelector('.status');
            const statusText = item.querySelector('.friend-info p');
            
            if (statusIndicator) {
                console.log(`🎨 Updating status indicator class: ${isOnline ? 'online' : 'offline'}`);
                statusIndicator.className = 'status ' + (isOnline ? 'online' : 'offline');
            } else {
                console.warn(`⚠️ Status indicator not found in friend item`);
            }
            
            if (statusText) {
                console.log(`📝 Updating status text: ${isOnline ? 'Đang hoạt động' : 'Offline'}`);
                statusText.textContent = isOnline ? 'Đang hoạt động' : 'Offline';
            } else {
                console.warn(`⚠️ Status text element not found`);
            }
        }
    });
    
    // Update in conversation list
    const conversationItems = document.querySelectorAll('.conversation-item');
    conversationItems.forEach(item => {
        const convUserId = item.getAttribute('data-user-id');
        if (convUserId && parseInt(convUserId) === userId) {
            const statusIndicator = item.querySelector('.status');
            if (statusIndicator) {
                statusIndicator.className = 'status ' + (isOnline ? 'online' : 'offline');
            }
        }
    });
    
    // Update in search results
    const searchResultItems = document.querySelectorAll('.search-result-item');
    searchResultItems.forEach(item => {
        const searchUserId = item.getAttribute('data-user-id');
        if (searchUserId && parseInt(searchUserId) === userId) {
            const statusIndicator = item.querySelector('.status');
            if (statusIndicator) {
                statusIndicator.className = 'status ' + (isOnline ? 'online' : 'offline');
            }
        }
    });
    
    // Update active chat header if this is the user we're chatting with
    if (activeConversation && activeConversation.id === userId) {
        console.log(`🎯 Updating active chat header for user ${userId}`);
        
        // Update status text
        const chatUserStatus = document.querySelector('#chatUserStatus');
        if (chatUserStatus) {
            console.log(`📝 Updating chat header status text: ${isOnline ? 'Đang hoạt động' : 'Offline'}`);
            chatUserStatus.textContent = isOnline ? 'Đang hoạt động' : 'Offline';
        } else {
            console.warn(`⚠️ #chatUserStatus element not found`);
        }
        
        // Update status dot in chat header - try multiple selectors
        let chatHeaderStatusDot = document.querySelector('.chat-header .online-status');
        if (!chatHeaderStatusDot) {
            chatHeaderStatusDot = document.querySelector('.chat-header .status');
        }
        
        if (chatHeaderStatusDot) {
            console.log(`🎨 Updating chat header status dot. Current classes: ${chatHeaderStatusDot.className}`);
            // Remove all status-related classes first
            chatHeaderStatusDot.classList.remove('online', 'offline', 'status');
            // Add base class if needed
            if (!chatHeaderStatusDot.classList.contains('online-status') && !chatHeaderStatusDot.classList.contains('status')) {
                chatHeaderStatusDot.classList.add('status');
            }
            // Add online/offline class
            chatHeaderStatusDot.classList.add(isOnline ? 'online' : 'offline');
            console.log(`🎨 New classes: ${chatHeaderStatusDot.className}`);
        } else {
            console.warn(`⚠️ Status dot in chat header not found`);
        }
    } else {
        console.log(`⏭️ User ${userId} is not active conversation (active: ${activeConversation?.id})`);
    }
    
    // Reload friends list to re-sort by online status
    if (window.loadFriends && typeof window.loadFriends === 'function') {
        window.loadFriends();
    }
}

