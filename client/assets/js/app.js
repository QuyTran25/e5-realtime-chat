document.addEventListener('DOMContentLoaded', function() {
    // Check authentication
    checkAuthentication();
    
    // Initialize app components
    initializeNavigation();
    initializeUserInfo();
    initializeConnectionStatus();
    
    // Initialize demo data
    initializeDemoData();
});

function checkAuthentication() {
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    
    if (!userData) {
        window.location.href = 'login.html';
        return;
    }
    
    try {
        const user = JSON.parse(userData);
        if (!user.token) {
            throw new Error('Invalid session');
        }
        
        // Set user info in UI
        document.getElementById('userName').textContent = user.name || 'User';
        document.getElementById('userStatus').textContent = 'Đang hoạt động';
        
    } catch (error) {
        localStorage.removeItem('user');
        sessionStorage.removeItem('user');
        window.location.href = 'login.html';
    }
}

function initializeNavigation() {
    const navItems = document.querySelectorAll('.nav-item[data-section]');
    const sections = document.querySelectorAll('.section');
    const logoutBtn = document.getElementById('logoutBtn');
    
    // Navigation click handlers
    navItems.forEach(item => {
        item.addEventListener('click', function() {
            const sectionId = this.dataset.section;
            
            // Remove active class from all nav items and sections
            navItems.forEach(nav => nav.classList.remove('active'));
            sections.forEach(section => section.classList.remove('active'));
            
            // Add active class to clicked nav item and corresponding section
            this.classList.add('active');
            const targetSection = document.getElementById(sectionId + 'Section');
            if (targetSection) {
                targetSection.classList.add('active');
            }
        });
    });
    
    // Logout handler
    logoutBtn.addEventListener('click', function() {
        showLogoutModal();
    });
}

function initializeUserInfo() {
    // Set default avatar if image fails to load
    const userAvatar = document.getElementById('userAvatar');
    const chatUserAvatar = document.getElementById('chatUserAvatar');
    
    userAvatar.onerror = function() {
        this.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzlDQTNBRiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K';
    };
    
    chatUserAvatar.onerror = function() {
        this.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzlDQTNBRiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K';
    };
}

function initializeConnectionStatus() {
    const connectionStatus = document.getElementById('connectionStatus');
    const statusIndicator = connectionStatus.querySelector('.status-indicator');
    const statusText = connectionStatus.querySelector('.status-text');
    
    // Simulate connection status
    setTimeout(() => {
        connectionStatus.classList.add('connected');
        statusText.textContent = 'Đã kết nối';
        
        // Hide status after 3 seconds
        setTimeout(() => {
            connectionStatus.style.opacity = '0';
            setTimeout(() => {
                connectionStatus.style.display = 'none';
            }, 300);
        }, 3000);
    }, 2000);
}

function initializeDemoData() {
    // Demo conversations
    /*const demoConversations = [
        {
            id: 1,
            name: 'Nguyễn Văn A',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzM0QThGNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            lastMessage: 'Chào bạn! Bạn có khỏe không?',
            time: '10:30',
            unread: 2,
            online: true
        },
        {
            id: 2,
            name: 'Trần Thị B',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iI0VGNDQ0NCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            lastMessage: 'Cảm ơn bạn nhiều!',
            time: '09:15',
            unread: 0,
            online: false
        },
        {
            id: 3,
            name: 'Lê Văn C',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzEwQjk4MSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            lastMessage: 'Hẹn gặp lại!',
            time: 'Hôm qua',
            unread: 0,
            online: true
        }
    ];*/
    
    // Demo friend requests
    const demoFriendRequests = [
        {
            id: 1,
            name: 'Phạm Văn D',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iI0Y1OUUwQiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            mutualFriends: 5
        },
        {
            id: 2,
            name: 'Hoàng Thị E',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzgzMzNGRiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            mutualFriends: 12
        }
    ];
    
    // Demo friends list
    const demoFriends = [
        {
            id: 1,
            name: 'Nguyễn Văn A',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzM0QThGNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            status: 'Đang hoạt động',
            online: true
        },
        {
            id: 2,
            name: 'Trần Thị B',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iI0VGNDQ0NCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            status: 'Offline 2 giờ trước',
            online: false
        }
    ];
    
    // Render conversations
    renderConversations(demoConversations);
    
    // Render friend requests
    renderFriendRequests(demoFriendRequests);
    
    // Render friends list
    renderFriends(demoFriends);
    
    // Setup search functionality
    setupSearch();
}

function renderConversations(conversations) {
    const conversationsList = document.getElementById('conversationsList');
    
    conversationsList.innerHTML = conversations.map(conv => `
        <div class="conversation-item" data-id="${conv.id}">
            <img src="${conv.avatar}" alt="${conv.name}" class="conversation-avatar">
            <div class="conversation-info">
                <div class="conversation-name">${conv.name}</div>
                <div class="conversation-preview">${conv.lastMessage}</div>
            </div>
            <div class="conversation-meta">
                <div class="conversation-time">${conv.time}</div>
                ${conv.unread > 0 ? `<div class="unread-badge">${conv.unread}</div>` : ''}
            </div>
        </div>
    `).join('');
    
    // Add click handlers
    conversationsList.querySelectorAll('.conversation-item').forEach(item => {
        item.addEventListener('click', function() {
            const convId = this.dataset.id;
            const conversation = conversations.find(c => c.id == convId);
            openChat(conversation);
            
            // Mark as active
            conversationsList.querySelectorAll('.conversation-item').forEach(i => i.classList.remove('active'));
            this.classList.add('active');
        });
    });
}

function renderFriendRequests(requests) {
    const friendRequestsList = document.getElementById('friendRequestsList');
    
    friendRequestsList.innerHTML = requests.map(req => `
        <div class="friend-request-item" data-id="${req.id}">
            <img src="${req.avatar}" alt="${req.name}" class="friend-avatar">
            <div class="friend-info">
                <div class="friend-name">${req.name}</div>
                <div class="friend-status">${req.mutualFriends} bạn chung</div>
            </div>
            <div class="friend-actions">
                <button class="btn-accept" onclick="acceptFriendRequest(${req.id})">Chấp nhận</button>
                <button class="btn-decline" onclick="declineFriendRequest(${req.id})">Từ chối</button>
            </div>
        </div>
    `).join('');
}

function renderFriends(friends) {
    const friendsList = document.getElementById('friendsList');
    
    friendsList.innerHTML = friends.map(friend => `
        <div class="friend-item" data-id="${friend.id}">
            <img src="${friend.avatar}" alt="${friend.name}" class="friend-avatar">
            <div class="friend-info">
                <div class="friend-name">${friend.name}</div>
                <div class="friend-status">${friend.status}</div>
            </div>
            <div class="friend-actions">
                <button class="btn-message" onclick="startChat(${friend.id})">Nhắn tin</button>
            </div>
        </div>
    `).join('');
}

function setupSearch() {
    const searchInput = document.getElementById('searchInput');
    const searchResults = document.getElementById('searchResults');
    
    searchInput.addEventListener('input', function() {
        const query = this.value.trim();
        
        if (query.length < 2) {
            searchResults.innerHTML = '';
            return;
        }
        
        // Demo search results
        const mockResults = [
            { id: 1, name: 'Nguyễn Văn X', status: 'Đang hoạt động' },
            { id: 2, name: 'Trần Thị Y', status: 'Offline 1 giờ trước' },
            { id: 3, name: 'Lê Văn Z', status: 'Đang hoạt động' }
        ].filter(user => user.name.toLowerCase().includes(query.toLowerCase()));
        
        searchResults.innerHTML = mockResults.map(user => `
            <div class="search-result-item" onclick="viewProfile(${user.id})">
                <img src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzZCNzI4MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K" alt="${user.name}" class="friend-avatar">
                <div class="friend-info">
                    <div class="friend-name">${user.name}</div>
                    <div class="friend-status">${user.status}</div>
                </div>
            </div>
        `).join('');
    });
}

function openChat(conversation) {
    // Update chat header
    document.getElementById('chatUserName').textContent = conversation.name;
    document.getElementById('chatUserStatus').textContent = conversation.online ? 'Đang hoạt động' : 'Offline';
    document.getElementById('chatUserAvatar').src = conversation.avatar;
    
    // Show chat input
    document.getElementById('chatInputContainer').style.display = 'block';
    
    // Hide welcome message and show demo messages
    const chatMessages = document.getElementById('chatMessages');
    chatMessages.innerHTML = `
        <div class="message">
            <img src="${conversation.avatar}" alt="${conversation.name}" class="message-avatar">
            <div class="message-content">
                <div class="message-bubble">${conversation.lastMessage}</div>
                <div class="message-time">10:30</div>
            </div>
        </div>
        <div class="message own">
            <div class="message-content">
                <div class="message-bubble">Chào bạn! Mình khỏe, cảm ơn bạn!</div>
                <div class="message-time">10:31</div>
            </div>
        </div>
    `;
}

// Global functions for button handlers
function acceptFriendRequest(id) {
    if (confirm('Chấp nhận lời mời kết bạn?')) {
        document.querySelector(`[data-id="${id}"]`).remove();
        updateFriendRequestBadge();
        showNotification('Đã chấp nhận lời mời kết bạn!');
    }
}

function declineFriendRequest(id) {
    if (confirm('Từ chối lời mời kết bạn?')) {
        document.querySelector(`[data-id="${id}"]`).remove();
        updateFriendRequestBadge();
        showNotification('Đã từ chối lời mời kết bạn!');
    }
}

function startChat(friendId) {
    // Tìm bạn trong danh sách friends demo
    const allFriends = [
        {
            id: 1,
            name: 'Nguyễn Văn A',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzM0QThGNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            status: 'Đang hoạt động',
            online: true
        },
        {
            id: 2,
            name: 'Trần Thị B',
            avatar: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iI0VGNDQ0NCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
            status: 'Offline 2 giờ trước',
            online: false
        }
    ];

    const friend = allFriends.find(f => f.id === friendId);
    if (!friend) return;

    // Chuyển sang tab "Messages"
    document.querySelector('[data-section="messages"]').click();

    // Giả lập mở đoạn chat
    openChat({
        id: friend.id,
        name: friend.name,
        avatar: friend.avatar,
        lastMessage: 'Bắt đầu cuộc trò chuyện mới với ' + friend.name,
        online: friend.online
    });

    // Hiển thị thông báo nhỏ
    showNotification(`Đang trò chuyện với ${friend.name}`);
}

function viewProfile(userId) {
    showNotification('Xem thông tin người dùng (chức năng đang phát triển)');
}

function updateFriendRequestBadge() {
    const badge = document.getElementById('friendRequestBadge');
    const remainingRequests = document.querySelectorAll('.friend-request-item').length;
    
    if (remainingRequests > 0) {
        badge.textContent = remainingRequests;
    } else {
        badge.style.display = 'none';
    }
}

function showNotification(message) {
    // Create simple notification
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 70px;
        right: 20px;
        background: #1F2937;
        color: white;
        padding: 12px 16px;
        border-radius: 8px;
        font-size: 14px;
        z-index: 1001;
        animation: slideIn 0.3s ease;
    `;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Add CSS for animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
`;
document.head.appendChild(style);

// Initialize logout modal when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    initializeLogoutModal();
});

// Logout Modal Functions
function showLogoutModal() {
    const modal = document.getElementById('logoutModal');
    modal.classList.add('show');
    document.body.style.overflow = 'hidden';
}

function hideLogoutModal() {
    const modal = document.getElementById('logoutModal');
    modal.classList.remove('show');
    document.body.style.overflow = 'auto';
}

function initializeLogoutModal() {
    const modal = document.getElementById('logoutModal');
    const cancelBtn = document.getElementById('logoutCancel');
    const confirmBtn = document.getElementById('logoutConfirm');
    
    if (!modal || !cancelBtn || !confirmBtn) return;
    
    // Cancel logout
    cancelBtn.addEventListener('click', hideLogoutModal);
    
    // Confirm logout
    confirmBtn.addEventListener('click', function() {
        performLogout();
    });
    
    // Close modal when clicking overlay
    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            hideLogoutModal();
        }
    });
    
    // Close modal with Escape key
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && modal.classList.contains('show')) {
            hideLogoutModal();
        }
    });
}

function performLogout() {
    const confirmBtn = document.getElementById('logoutConfirm');
    
    // Show loading state
    confirmBtn.classList.add('loading');
    confirmBtn.disabled = true;
    
    // Simulate logout process
    setTimeout(() => {
        // Clear user data
        localStorage.removeItem('user');
        sessionStorage.removeItem('user');
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
        
        // Show success message briefly
        showNotification('Đã đăng xuất thành công!');
        
        // Redirect to login
        setTimeout(() => {
            window.location.href = 'login.html';
        }, 1000);
        
    }, 1500); // Simulate API call delay
}