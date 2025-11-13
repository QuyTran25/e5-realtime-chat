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
    // Load conversations from API instead of demo data
    loadConversationsFromAPI();
    
    // Load friend requests from API
    loadFriendRequestsFromAPI();
    
    // Load friends list from API
    loadFriendsFromAPI();
    
    // Setup search functionality
    setupSearch();
    
    // Auto-refresh friend requests every 30 seconds
    setInterval(loadFriendRequestsFromAPI, 30000);
    
    // Auto-refresh friends list every 60 seconds
    setInterval(loadFriendsFromAPI, 60000);
}

// Load conversations from API
let loadingConversations = false;
async function loadConversationsFromAPI() {
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) return;

    // Prevent concurrent requests
    if (loadingConversations) {
        console.log('⏳ Conversations already loading, skipping...');
        return;
    }

    try {
        loadingConversations = true;
        const user = JSON.parse(userData);
        const response = await fetch('http://localhost:8080/api/conversations', {
            headers: {
                'Authorization': `Bearer ${user.token}`
            }
        });

        if (response.status === 429) {
            console.warn('⚠️ Rate limited. Will retry in 3 seconds...');
            setTimeout(() => loadConversationsFromAPI(), 3000);
            return;
        }

        if (!response.ok) {
            throw new Error('Failed to load conversations');
        }

        const data = await response.json();
        if (data.success && data.conversations) {
            renderConversations(data.conversations.map(conv => {
                // Get unread count from WebSocket tracking
                const unreadCount = window.getUnreadCount ? window.getUnreadCount(conv.user_id) : 0;
                
                return {
                    id: conv.user_id,
                    name: conv.username,
                    avatar: conv.avatar_url || 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzZCNzI4MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
                    lastMessage: conv.last_message,
                    time: formatTime(conv.last_message_at),
                    unread: unreadCount,
                    online: conv.is_online
                };
            }));
        }
    } catch (error) {
        console.error('❌ Error loading conversations:', error);
    } finally {
        loadingConversations = false;
    }
}

// Load friend requests from API (with loading flag to prevent concurrent requests)
let loadingFriendRequests = false;
async function loadFriendRequestsFromAPI() {
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) return;

    // Prevent concurrent requests
    if (loadingFriendRequests) {
        console.log('⏳ Friend requests already loading, skipping...');
        return;
    }

    try {
        loadingFriendRequests = true;
        const user = JSON.parse(userData);
        const response = await fetch('http://localhost:8080/api/friends/requests', {
            headers: {
                'Authorization': `Bearer ${user.token}`
            }
        });

        if (response.status === 429) {
            console.warn('⚠️ Rate limited. Will retry in 4 seconds...');
            setTimeout(() => loadFriendRequestsFromAPI(), 4000);
            return;
        }

        if (!response.ok) {
            throw new Error('Failed to load friend requests');
        }

        const requests = await response.json();
        console.log('✅ Loaded friend requests:', requests);
        
        // Handle null or undefined response
        const requestList = Array.isArray(requests) ? requests : [];
        
        // Update badge
        updateFriendRequestBadge(requestList.length);
        
        // Render friend requests
        renderFriendRequests(requestList);
    } catch (error) {
        console.error('❌ Error loading friend requests:', error);
    } finally {
        loadingFriendRequests = false;
    }
}

// Load friends list from API (with loading flag to prevent concurrent requests)
let loadingFriends = false;
async function loadFriendsFromAPI() {
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) return;

    // Prevent concurrent requests
    if (loadingFriends) {
        console.log('⏳ Friends list already loading, skipping...');
        return;
    }

    try {
        loadingFriends = true;
        const user = JSON.parse(userData);
        const response = await fetch('http://localhost:8080/api/friends', {
            headers: {
                'Authorization': `Bearer ${user.token}`
            }
        });

        if (response.status === 429) {
            console.warn('⚠️ Rate limited. Will retry in 2 seconds...');
            setTimeout(() => loadFriendsFromAPI(), 2000);
            return;
        }

        if (!response.ok) {
            throw new Error('Failed to load friends');
        }

        const friends = await response.json();
        console.log('✅ Loaded friends:', friends);
        
        // Render friends list
        renderFriends(friends || []);
    } catch (error) {
        console.error('❌ Error loading friends:', error);
    } finally {
        loadingFriends = false;
    }
}

// Format time for display
function formatTime(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
    
    // Less than 1 minute
    if (diff < 60000) {
        return 'Vừa xong';
    }
    
    // Less than 1 hour
    if (diff < 3600000) {
        return Math.floor(diff / 60000) + ' phút trước';
    }
    
    // Today
    if (date.toDateString() === now.toDateString()) {
        return date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' });
    }
    
    // Yesterday
    const yesterday = new Date(now);
    yesterday.setDate(yesterday.getDate() - 1);
    if (date.toDateString() === yesterday.toDateString()) {
        return 'Hôm qua';
    }
    
    // Older
    return date.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' });
}

// Update conversation list (called after sending/receiving message)
window.updateConversationList = loadConversationsFromAPI;

// Expose friend request functions globally
window.loadFriendRequestsFromAPI = loadFriendRequestsFromAPI;
window.loadFriendsFromAPI = loadFriendsFromAPI;

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
    
    if (!requests || requests.length === 0) {
        friendRequestsList.innerHTML = `
            <div style="padding: 40px; text-align: center; color: #9CA3AF;">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="currentColor" style="margin-bottom: 16px;">
                    <path d="M15,14C12.33,14 7,15.33 7,18V20H23V18C23,15.33 17.67,14 15,14M6,10V7H4V10H1V12H4V15H6V12H9V10M15,12A4,4 0 0,0 19,8A4,4 0 0,0 15,4A4,4 0 0,0 11,8A4,4 0 0,0 15,12Z"/>
                </svg>
                <p>Chưa có lời mời kết bạn nào</p>
            </div>
        `;
        return;
    }
    
    friendRequestsList.innerHTML = requests.map(req => `
        <div class="friend-request-item" data-id="${req.id}">
            <img src="${req.avatar || req.avatar_url || 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzZCNzI4MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K'}" alt="${req.name || req.username}" class="friend-avatar">
            <div class="friend-info">
                <div class="friend-name">${req.name || req.username}</div>
                <div class="friend-status">Muốn kết bạn với bạn</div>
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
    
    if (!friends || friends.length === 0) {
        friendsList.innerHTML = `
            <div style="padding: 40px; text-align: center; color: #9CA3AF;">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="currentColor" style="margin-bottom: 16px;">
                    <path d="M16,4C18.21,4 20,5.79 20,8C20,10.21 18.21,12 16,12C13.79,12 12,10.21 12,8C12,5.79 13.79,4 16,4M16,13C18.67,13 22,14.33 22,17V20H10V17C10,14.33 13.33,13 16,13M8,4C10.21,4 12,5.79 12,8C12,10.21 10.21,12 8,12C5.79,12 4,10.21 4,8C4,5.79 5.79,4 8,4M8,13C10.67,13 14,14.33 14,17V20H2V17C2,14.33 5.33,13 8,13Z"/>
                </svg>
                <p>Chưa có bạn bè</p>
            </div>
        `;
        return;
    }
    
    friendsList.innerHTML = friends.map(friend => {
        const isOnline = friend.online || friend.is_online;
        const statusText = isOnline ? 'Đang hoạt động' : 'Offline';
        
        return `
            <div class="friend-item" data-id="${friend.id}">
                <div style="position: relative;">
                    <img src="${friend.avatar || friend.avatar_url || 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzZCNzI4MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K'}" alt="${friend.name || friend.username}" class="friend-avatar">
                    ${isOnline ? '<span class="online-indicator" style="position: absolute; bottom: 2px; right: 2px; width: 12px; height: 12px; border-radius: 50%;"></span>' : ''}
                </div>
                <div class="friend-info">
                    <div class="friend-name">${friend.name || friend.username}</div>
                    <div class="friend-status">${statusText}</div>
                </div>
                <div class="friend-actions">
                    <button class="btn-message" onclick="startChat(${friend.id}, '${friend.name || friend.username}', '${friend.avatar || friend.avatar_url}', ${isOnline})">Nhắn tin</button>
                </div>
            </div>
        `;
    }).join('');
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
    // Set active conversation for WebSocket
    if (window.setActiveConversation) {
        window.setActiveConversation({
            id: conversation.id,
            name: conversation.name,
            avatar: conversation.avatar
        });
    }
    
    // Update chat header
    document.getElementById('chatUserName').textContent = conversation.name;
    document.getElementById('chatUserStatus').textContent = conversation.online ? 'Đang hoạt động' : 'Offline';
    document.getElementById('chatUserAvatar').src = conversation.avatar;
    
    // Show chat input
    document.getElementById('chatInputContainer').style.display = 'block';
    
    // Chat history will be loaded by setActiveConversation via websocket.js
}

// Global functions for button handlers
async function acceptFriendRequest(id) {
    if (!confirm('Chấp nhận lời mời kết bạn?')) return;
    
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) return;

    try {
        // Optimistic UI update - Remove the request item immediately
        const requestItem = document.querySelector(`.friend-request-item[data-id="${id}"]`);
        if (requestItem) {
            requestItem.style.opacity = '0.5';
            requestItem.style.pointerEvents = 'none';
        }
        
        const user = JSON.parse(userData);
        const response = await fetch('http://localhost:8080/api/friends/accept', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${user.token}`
            },
            body: JSON.stringify({ friend_id: id })
        });

        if (!response.ok) {
            // Revert UI if failed
            if (requestItem) {
                requestItem.style.opacity = '1';
                requestItem.style.pointerEvents = 'auto';
            }
            throw new Error('Failed to accept friend request');
        }

        const result = await response.json();
        console.log('✅ Friend request accepted:', result);
        
        showNotification('✅ Đã chấp nhận lời mời kết bạn!');
        
        // Remove the item from DOM
        if (requestItem) {
            requestItem.remove();
        }
        
        // Update badge count
        const remainingRequests = document.querySelectorAll('.friend-request-item').length;
        updateFriendRequestBadge(remainingRequests);
        
        // Wait longer for backend to process and avoid rate limiting
        await new Promise(resolve => setTimeout(resolve, 1500));
        await loadFriendsFromAPI();
        
        // If no more requests, show empty state
        if (remainingRequests === 0) {
            const friendRequestsList = document.getElementById('friendRequestsList');
            friendRequestsList.innerHTML = `
                <div style="padding: 40px; text-align: center; color: #9CA3AF;">
                    <svg width="64" height="64" viewBox="0 0 24 24" fill="currentColor" style="margin-bottom: 16px;">
                        <path d="M15,14C12.33,14 7,15.33 7,18V20H23V18C23,15.33 17.67,14 15,14M6,10V7H4V10H1V12H4V15H6V12H9V10M15,12A4,4 0 0,0 19,8A4,4 0 0,0 15,4A4,4 0 0,0 11,8A4,4 0 0,0 15,12Z"/>
                    </svg>
                    <p>Chưa có lời mời kết bạn nào</p>
                </div>
            `;
        }
        
    } catch (error) {
        console.error('❌ Error accepting friend request:', error);
        showNotification('❌ Không thể chấp nhận lời mời kết bạn');
    }
}

async function declineFriendRequest(id) {
    if (!confirm('Từ chối lời mời kết bạn?')) return;
    
    const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (!userData) return;

    try {
        // Optimistic UI update - Remove the request item immediately
        const requestItem = document.querySelector(`.friend-request-item[data-id="${id}"]`);
        if (requestItem) {
            requestItem.style.opacity = '0.5';
            requestItem.style.pointerEvents = 'none';
        }
        
        const user = JSON.parse(userData);
        const response = await fetch('http://localhost:8080/api/friends/reject', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${user.token}`
            },
            body: JSON.stringify({ friend_id: id })
        });

        if (!response.ok) {
            // Revert UI if failed
            if (requestItem) {
                requestItem.style.opacity = '1';
                requestItem.style.pointerEvents = 'auto';
            }
            throw new Error('Failed to reject friend request');
        }

        const result = await response.json();
        console.log('✅ Friend request rejected:', result);
        
        showNotification('Đã từ chối lời mời kết bạn');
        
        // Remove the item from DOM
        if (requestItem) {
            requestItem.remove();
        }
        
        // Update badge count
        const remainingRequests = document.querySelectorAll('.friend-request-item').length;
        updateFriendRequestBadge(remainingRequests);
        
        // If no more requests, show empty state
        if (remainingRequests === 0) {
            const friendRequestsList = document.getElementById('friendRequestsList');
            friendRequestsList.innerHTML = `
                <div style="padding: 40px; text-align: center; color: #9CA3AF;">
                    <svg width="64" height="64" viewBox="0 0 24 24" fill="currentColor" style="margin-bottom: 16px;">
                        <path d="M15,14C12.33,14 7,15.33 7,18V20H23V18C23,15.33 17.67,14 15,14M6,10V7H4V10H1V12H4V15H6V12H9V10M15,12A4,4 0 0,0 19,8A4,4 0 0,0 15,4A4,4 0 0,0 11,8A4,4 0 0,0 15,12Z"/>
                    </svg>
                    <p>Chưa có lời mời kết bạn nào</p>
                </div>
            `;
        }
        
    } catch (error) {
        console.error('❌ Error rejecting friend request:', error);
        showNotification('❌ Không thể từ chối lời mời kết bạn');
    }
}

function startChat(friendId, friendName, friendAvatar, isOnline) {
    // Chuyển sang tab "Messages"
    document.querySelector('[data-section="messages"]').click();

    // Mở đoạn chat
    openChat({
        id: friendId,
        name: friendName,
        avatar: friendAvatar || 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0iIzZCNzI4MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPgo8L3N2Zz4K',
        lastMessage: 'Bắt đầu cuộc trò chuyện',
        online: isOnline
    });

    // Hiển thị thông báo nhỏ
    // showNotification(`💬 Đang trò chuyện với ${friendName}`);
}

function viewProfile(userId) {
    showNotification('Xem thông tin người dùng (chức năng đang phát triển)');
}

function updateFriendRequestBadge(count) {
    const badge = document.getElementById('friendRequestBadge');
    
    if (!badge) return;
    
    if (count && count > 0) {
        badge.textContent = count;
        badge.style.display = 'inline-block';
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
    try {
        if (typeof logout === 'function') {
            // Call the logout function implemented in websocket.js
            logout();
            return;
        }
    } catch (e) {
        console.warn('logout() function not available, falling back to local clear');
    }

    // Fallback behaviour (best-effort): clear local session and redirect
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

    }, 1500);
}