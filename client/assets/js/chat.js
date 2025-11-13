// ========================
// 🟢 Biến toàn cục
// ========================
let friendsListData = [];
let activeChatUser = null;
let searchTimeout = null;

// ========================
// 🟢 Helper: Get auth token
// ========================
function getAuthToken() {
  const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
  if (!userData) return null;
  try {
    const user = JSON.parse(userData);
    return user.token;
  } catch (e) {
    return null;
  }
}

// ========================
// 🟢 Helper: API call với authentication
// ========================
async function apiCall(url, options = {}) {
  const token = getAuthToken();
  if (!token) {
    throw new Error('No authentication token');
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
    ...options.headers
  };

  const response = await fetch(url, { ...options, headers });
  if (!response.ok) {
    throw new Error(`API call failed: ${response.statusText}`);
  }
  return response.json();
}

// ========================
// 🟢 Khi trang load xong
// ========================
document.addEventListener("DOMContentLoaded", function () {
  const searchInput = document.querySelector("#searchInput");
  const searchResults = document.querySelector("#searchResults");
  const friendsList = document.querySelector("#friendsList");
  const friendRequestsList = document.querySelector("#friendRequestsList");

  const chatPanel = document.querySelector("#chatPanel");
  const chatUserName = document.querySelector("#chatUserName");
  const chatUserStatus = document.querySelector("#chatUserStatus");
  const chatUserAvatar = document.querySelector("#chatUserAvatar");
  const chatMessages = document.querySelector("#chatMessages");
  const chatInputContainer = document.querySelector("#chatInputContainer");

  // Kiểm tra xem có ở trang index.html không (có friendsList element)
  if (!friendsList) {
    console.log("Chat.js: Not on main chat page, skipping initialization");
    return;
  }

  // ====== 1️⃣ LOAD DANH SÁCH BẠN BÈ ======
  loadFriends();

  // ====== 2️⃣ LOAD LỜI MỜI KẾT BẠN ======
  loadFriendRequests();

  // ====== 3️⃣ XỬ LÝ TÌM KIẾM NGƯỜI DÙNG (debounced) ======
  if (searchInput) {
    searchInput.addEventListener("input", function () {
      const keyword = searchInput.value.trim();

      // Clear previous timeout
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }

      if (keyword === "") {
        searchResults.innerHTML = "";
        return;
      }

      // Debounce: Chỉ search sau 500ms user ngừng gõ
      searchTimeout = setTimeout(() => {
        searchUsers(keyword);
      }, 500);
    });
  }

  // ====== FUNCTIONS ======

  // Load danh sách bạn bè
  async function loadFriends() {
    try {
      const friends = await apiCall("http://localhost:8080/api/friends");
      friendsListData = friends || [];
      renderFriends(friends || []);
    } catch (err) {
      console.error("⚠️ Lỗi khi tải danh sách bạn bè:", err);
      if (friendsList) {
        friendsList.innerHTML = `<p style="color:red; padding: 20px;">Không thể tải danh sách bạn bè. Vui lòng đăng nhập lại.</p>`;
      }
    }
  }

  // Load lời mời kết bạn
  async function loadFriendRequests() {
    try {
      const requests = await apiCall("http://localhost:8080/api/friends/requests");
      renderFriendRequests(requests || []);
      
      // Update badge
      const badge = document.querySelector("#friendRequestBadge");
      if (badge) {
        if (requests && requests.length > 0) {
          badge.textContent = requests.length;
          badge.style.display = 'inline';
        } else {
          badge.style.display = 'none';
        }
      }
    } catch (err) {
      console.error("⚠️ Lỗi khi tải lời mời kết bạn:", err);
    }
  }

  // Tìm kiếm người dùng
  async function searchUsers(keyword) {
    try {
      if (!searchResults) {
        console.error("searchResults element not found");
        return;
      }
      
      searchResults.innerHTML = '<p style="padding: 10px; color: gray;">Đang tìm kiếm...</p>';
      
      const url = `http://localhost:8080/api/friends/search?q=${encodeURIComponent(keyword)}`;
      const users = await apiCall(url);
      
      renderSearchResults(users || []);
    } catch (err) {
      console.error("Lỗi khi tìm kiếm:", err);
      if (searchResults) {
        searchResults.innerHTML = '<p style="padding: 10px; color: red;">Lỗi khi tìm kiếm</p>';
      }
    }
  }

  // ====== RENDER FUNCTIONS ======

  // Render danh sách bạn bè
  function renderFriends(list) {
    if (!friendsList) return;
    
    friendsList.innerHTML = "";
    if (!list || list.length === 0) {
      friendsList.innerHTML = "<p style='padding: 20px; text-align: center; color: gray;'>Chưa có bạn bè nào</p>";
      return;
    }

    list.forEach(friend => {
      const item = document.createElement("div");
      item.classList.add("friend-item");
      item.style.cursor = "pointer";
      item.innerHTML = `
        <div class="friend-avatar">
          <img src="${friend.avatar || friend.avatar_url || '../assets/images/default-avatar.svg'}" alt="Avatar">
          <span class="status ${friend.online || friend.is_online ? 'online' : 'offline'}"></span>
        </div>
        <div class="friend-info">
          <h4>${friend.name || friend.username}</h4>
          <p>${friend.online || friend.is_online ? 'Đang hoạt động' : 'Offline'}</p>
        </div>
      `;

      item.addEventListener("click", function () {
        openChat(friend);
      });

      friendsList.appendChild(item);
    });
  }

  // Render lời mời kết bạn
  function renderFriendRequests(requests) {
    if (!friendRequestsList) return;

    friendRequestsList.innerHTML = "";
    if (!requests || requests.length === 0) {
      friendRequestsList.innerHTML = "<p style='padding: 20px; text-align: center; color: gray;'>Không có lời mời kết bạn nào</p>";
      return;
    }

    requests.forEach(req => {
      const item = document.createElement("div");
      item.classList.add("friend-request-item");
      item.innerHTML = `
        <img src="${req.avatar || req.avatar_url || '../assets/images/default-avatar.svg'}" alt="${req.name || req.username}" class="friend-avatar">
        <div class="friend-info">
          <div class="friend-name">${req.name || req.username}</div>
          <div class="friend-status">${req.email || ''}</div>
        </div>
        <div class="friend-actions">
          <button class="btn-accept" data-id="${req.id}">Chấp nhận</button>
          <button class="btn-decline" data-id="${req.id}">Từ chối</button>
        </div>
      `;

      // Accept button
      const acceptBtn = item.querySelector(".btn-accept");
      acceptBtn.addEventListener("click", async (e) => {
        e.stopPropagation();
        await acceptFriendRequest(req.id);
      });

      // Decline button
      const declineBtn = item.querySelector(".btn-decline");
      declineBtn.addEventListener("click", async (e) => {
        e.stopPropagation();
        await rejectFriendRequest(req.id);
      });

      friendRequestsList.appendChild(item);
    });
  }

  // Render kết quả tìm kiếm
  function renderSearchResults(results) {
    if (!searchResults) return;

    searchResults.innerHTML = "";
    if (!results || results.length === 0) {
      searchResults.innerHTML = "<p style='padding: 10px; color: gray;'>Không tìm thấy người dùng nào</p>";
      return;
    }

    results.forEach(user => {
      const div = document.createElement("div");
      div.classList.add("search-result-item");
      div.style.cursor = "pointer";
      div.style.padding = "10px";
      div.style.borderBottom = "1px solid #eee";
      div.style.display = "flex";
      div.style.alignItems = "center";
      div.style.gap = "10px";

      const isFriend = user.status === 'friend';
      const isPending = user.status === 'pending';

      div.innerHTML = `
        <div class="result-avatar" style="width: 40px; height: 40px; border-radius: 50%; overflow: hidden;">
          <img src="${user.avatar || user.avatar_url || '../assets/images/default-avatar.svg'}" alt="Avatar" style="width: 100%; height: 100%; object-fit: cover;">
        </div>
        <div class="result-info" style="flex: 1;">
          <h4 style="margin: 0; font-size: 14px;">${user.name || user.username}</h4>
          <p style="margin: 0; font-size: 12px; color: gray;">${user.email || ''}</p>
        </div>
        <div class="result-action">
          ${
            isFriend
              ? '<span style="color: green; font-size: 13px;">✓ Bạn bè</span>'
              : isPending
              ? '<span style="color: orange; font-size: 13px;">⏳ Đang chờ</span>'
              : '<button class="add-friend-btn" data-id="' + user.id + '" style="padding: 5px 10px; background: #3B82F6; color: white; border: none; border-radius: 5px; cursor: pointer;">+ Kết bạn</button>'
          }
        </div>
      `;

      // Click để mở chat nếu đã là bạn
      if (isFriend) {
        div.addEventListener("click", () => openChat(user));
      }

      // Add friend button
      const addBtn = div.querySelector(".add-friend-btn");
      if (addBtn) {
        addBtn.addEventListener("click", async (e) => {
          e.stopPropagation();
          await sendFriendRequest(user.id, addBtn);
        });
      }

      searchResults.appendChild(div);
    });
  }

  // ====== ACTION FUNCTIONS ======

  // Gửi lời mời kết bạn
  async function sendFriendRequest(friendId, button) {
    try {
      button.disabled = true;
      button.textContent = "Đang gửi...";
      
      await apiCall("http://localhost:8080/api/friends/request", {
        method: "POST",
        body: JSON.stringify({ friend_id: friendId })
      });

      button.textContent = "✓ Đã gửi";
      button.style.background = "#10B981";
      
      alert("Đã gửi lời mời kết bạn!");
    } catch (err) {
      console.error("❌ Lỗi gửi lời mời:", err);
      button.disabled = false;
      button.textContent = "+ Kết bạn";
      alert("Không thể gửi lời mời kết bạn");
    }
  }

  // Chấp nhận lời mời kết bạn
  async function acceptFriendRequest(friendId) {
    try {
      await apiCall("http://localhost:8080/api/friends/accept", {
        method: "POST",
        body: JSON.stringify({ friend_id: friendId })
      });

      alert("Đã chấp nhận lời mời kết bạn!");
      
      // Reload danh sách
      loadFriends();
      loadFriendRequests();
    } catch (err) {
      console.error("❌ Lỗi chấp nhận lời mời:", err);
      alert("Không thể chấp nhận lời mời");
    }
  }

  // Từ chối lời mời kết bạn
  async function rejectFriendRequest(friendId) {
    try {
      await apiCall("http://localhost:8080/api/friends/reject", {
        method: "POST",
        body: JSON.stringify({ friend_id: friendId })
      });

      alert("Đã từ chối lời mời kết bạn");
      
      // Reload danh sách
      loadFriendRequests();
    } catch (err) {
      console.error("❌ Lỗi từ chối lời mời:", err);
      alert("Không thể từ chối lời mời");
    }
  }

  // Mở khung chat
  function openChat(user) {
    activeChatUser = user;

    if (!chatUserName || !chatUserStatus || !chatUserAvatar || !chatMessages || !chatInputContainer) {
      return;
    }

    // Chuyển sang tab Messages
    const messagesTab = document.querySelector('[data-section="messages"]');
    if (messagesTab) messagesTab.click();

    // Cập nhật thông tin người đang chat
    chatUserName.textContent = user.name || user.username;
    chatUserStatus.textContent = (user.online || user.is_online) ? "Đang hoạt động" : "Offline";
    chatUserAvatar.src = user.avatar || user.avatar_url || "../assets/images/default-avatar.svg";

    // Hiển thị ô nhập chat
    chatMessages.innerHTML = "";
    chatInputContainer.style.display = "block";

    const welcomeMsg = document.createElement("div");
    // welcomeMsg.classList.add("system-message");
    // welcomeMsg.style.textAlign = "center";
    // welcomeMsg.style.padding = "20px";
    // welcomeMsg.style.color = "gray";
    // welcomeMsg.textContent = `💬 Đang trò chuyện với ${user.name || user.username}`;
    chatMessages.appendChild(welcomeMsg);

    // 🔥 GỌI setActiveConversation từ websocket.js để thiết lập chat riêng tư
    if (typeof window.setActiveConversation === 'function') {
      window.setActiveConversation({
        id: user.id,
        name: user.name || user.username,
        avatar: user.avatar || user.avatar_url
      });
    }
  }
});
