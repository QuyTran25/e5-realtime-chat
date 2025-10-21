// ========================
// 🟢 Danh sách tất cả người dùng (bao gồm cả bạn bè và chưa là bạn)
// ========================
const allUsersData = [
  { name: "Nguyễn Văn A", avatar: "", online: true },
  { name: "Trần Thị B", avatar: "", online: false },
  { name: "Lê Minh C", avatar: "", online: true },
  { name: "Phạm Quốc D", avatar: "", online: false },
  { name: "Võ Thị E", avatar: "", online: true },
  { name: "Đặng Thanh F", avatar: "", online: false }
];

// ========================
// 🟢 Biến toàn cục
// ========================
let friendsListData = [];
let activeChatUser = null;

// ========================
// 🟢 Khi trang load xong
// ========================
document.addEventListener("DOMContentLoaded", function () {
  const searchInput = document.querySelector("#searchInput");
  const searchResults = document.querySelector("#searchResults");
  const friendsList = document.querySelector("#friendsList");

  const chatPanel = document.querySelector("#chatPanel");
  const chatUserName = document.querySelector("#chatUserName");
  const chatUserStatus = document.querySelector("#chatUserStatus");
  const chatUserAvatar = document.querySelector("#chatUserAvatar");
  const chatMessages = document.querySelector("#chatMessages");
  const chatInputContainer = document.querySelector("#chatInputContainer");

  // ====== 1️⃣ GỌI API DANH SÁCH BẠN BÈ ======
  fetch("http://localhost:8080/api/friends")
    .then(res => res.json())
    .then(friends => {
      friendsListData = friends;
      renderFriends(friends);

      // ====== 2️⃣ XỬ LÝ TÌM KIẾM NGƯỜI DÙNG ======
      searchInput.addEventListener("input", function () {
        const keyword = searchInput.value.toLowerCase().trim();

        if (keyword === "") {
          searchResults.innerHTML = "";
          return;
        }

        // 🔍 Tìm trong toàn bộ người dùng (bao gồm cả bạn bè và chưa là bạn)
        const results = allUsersData.filter(u =>
          u.name.toLowerCase().includes(keyword)
        );

        renderSearchResults(results);
      });
    })
    .catch(err => {
      console.error("⚠️ Lỗi khi tải danh sách bạn bè:", err);
      friendsList.innerHTML = `<p style="color:red;">Không thể tải danh sách bạn bè.</p>`;
    });

  // ====== 3️⃣ HÀM HIỂN THỊ DANH SÁCH BẠN BÈ ======
  function renderFriends(list) {
    friendsList.innerHTML = "";
    if (!list || list.length === 0) {
      friendsList.innerHTML = "<p>Chưa có bạn bè nào.</p>";
      return;
    }

    list.forEach(friend => {
      const item = document.createElement("div");
      item.classList.add("friend-item");
      item.innerHTML = `
        <div class="friend-avatar">
          <img src="${friend.avatar || '../assets/images/default-avatar.svg'}" alt="Avatar">
          <span class="status ${friend.online ? 'online' : 'offline'}"></span>
        </div>
        <div class="friend-info">
          <h4>${friend.name}</h4>
          <p>${friend.online ? 'Đang hoạt động' : 'Offline'}</p>
        </div>
      `;

      // Khi nhấn vào bạn bè → mở khung chat
      item.addEventListener("click", function () {
        openChat(friend);
      });

      friendsList.appendChild(item);
      item.addEventListener("click", function () {
        // 1️⃣ Chuyển sang tab Messages (Tin nhắn)
        const messagesTab = document.querySelector('[data-section="messages"]');
        if (messagesTab) messagesTab.click();

        // 2️⃣ Hiển thị tên bạn bè đang chat (hoặc mở đoạn chat cũ)
        const chatHeader = document.querySelector(".chat-header h3");
        if (chatHeader) chatHeader.textContent = friend.name;

        // 3️⃣ Có thể reset nội dung khung chat cũ
        const chatMessages = document.querySelector(".chat-messages");
        if (chatMessages) {
            chatMessages.innerHTML = `<div class="message system">💬 Bắt đầu trò chuyện với <b>${friend.name}</b></div>`;
        }
        });
    });
  }

  // ====== 4️⃣ HÀM HIỂN THỊ KẾT QUẢ TÌM KIẾM ======
  function renderSearchResults(results) {
    searchResults.innerHTML = "";
    if (!results || results.length === 0) {
      searchResults.innerHTML = "<p>Không tìm thấy người dùng nào.</p>";
      return;
    }

    results.forEach(user => {
      const div = document.createElement("div");
      div.classList.add("search-result-item");

      const isFriend = friendsListData.some(f => f.name === user.name);

      div.innerHTML = `
        <div class="result-avatar">
          <img src="${user.avatar || '../assets/images/default-avatar.svg'}" alt="Avatar">
        </div>
        <div class="result-info">
          <h4>${user.name}</h4>
          ${
            isFriend
              ? '<span style="color: gray; font-size: 13px;">Đã là bạn bè</span>'
              : '<button class="add-friend-btn">+ Kết bạn</button>'
          }
        </div>
      `;

      // Nếu là bạn rồi → click để mở chat
      if (isFriend) {
        div.addEventListener("click", () => openChat(user));
      }

      searchResults.appendChild(div);
    });
  }

  // ====== 5️⃣ HÀM MỞ KHUNG CHAT ======
  function openChat(user) {
    activeChatUser = user;

    // Cập nhật thông tin người đang chat
    chatUserName.textContent = user.name;
    chatUserStatus.textContent = user.online ? "Đang hoạt động" : "Offline";
    chatUserAvatar.src = user.avatar || "../assets/images/default-avatar.svg";

    // Ẩn welcome message, hiển thị ô nhập chat
    chatMessages.innerHTML = "";
    chatInputContainer.style.display = "block";

    const welcomeMsg = document.createElement("div");
    welcomeMsg.classList.add("system-message");
    welcomeMsg.textContent = `💬 Đang trò chuyện với ${user.name}`;
    chatMessages.appendChild(welcomeMsg);
  }
});
