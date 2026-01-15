// Dashboard.js - Главная панель для авторизованных пользователей

// Используем относительный путь через nginx proxy
// Nginx проксирует /auth-api/ на auth-сервис localhost:8080/api/v1/auth/
const AUTH_API_BASE = window.location.origin;
const VIBEEMEET_API_BASE = window.location.origin + "/api/v1";

// Флаг для предотвращения множественных редиректов
let isRedirecting = false;

// Безопасный парсинг JSON из localStorage
function safeParseJSON(key, defaultValue = {}) {
  try {
    const value = localStorage.getItem(key);
    if (!value || value === "undefined" || value === "null") {
      return defaultValue;
    }
    return JSON.parse(value);
  } catch (error) {
    console.error(`Error parsing ${key} from localStorage:`, error);
    return defaultValue;
  }
}

// Функция для обновления access token через refresh token
// Refresh token приходит в HttpOnly cookie, браузер отправляет его автоматически
async function refreshAccessToken() {
  try {
    const response = await fetch(`${AUTH_API_BASE}/auth-api/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include", // Включаем cookies для refresh token
    });
    
    if (!response.ok) {
      return false;
    }
    
    const data = await response.json();
    localStorage.setItem("accessToken", data.access_token);
    // Refresh token остается в cookie, не нужно сохранять в localStorage
    
    return true;
  } catch (error) {
    console.error("Refresh token error:", error);
    return false;
  }
}

// Проверка авторизации при загрузке
document.addEventListener("DOMContentLoaded", async () => {
  // Защита от множественных редиректов
  if (isRedirecting) {
    return;
  }

  const accessToken = localStorage.getItem("accessToken");
  const user = safeParseJSON("user", {});

  // Если нет токена или пользователя - редирект
  if (!accessToken) {
    isRedirecting = true;
    window.location.href = "index.html";
    return;
  }

  // Если есть токен, но нет user.id - попробуем обновить токен
  if (!user || !user.id) {
    console.log("User data incomplete, attempting to refresh token...");
    const refreshed = await refreshAccessToken();
    
    if (!refreshed) {
      // Если не удалось обновить - очищаем и редиректим
      localStorage.removeItem("accessToken");
      localStorage.removeItem("user");
      localStorage.removeItem("display_name");
      isRedirecting = true;
      window.location.href = "index.html";
      return;
    }
    
    // После обновления токена нужно получить данные пользователя
    // Но так как мы не знаем user_id, просто редиректим на index
    // Пользователь должен войти заново
    isRedirecting = true;
    window.location.href = "index.html";
    return;
  }

  // Элементы DOM
  const btnLogout = document.getElementById("btn-logout");
  const btnCreateRoom = document.getElementById("btn-create-room");
  const btnJoinRoom = document.getElementById("btn-join-room");
  const btnJoinSubmit = document.getElementById("btn-join-submit");
  const roomIdInput = document.getElementById("room-id-input");
  const userName = document.getElementById("user-name");
  const greeting = document.getElementById("greeting");

  // Проверяем, что все необходимые элементы существуют
  if (!btnLogout || !btnCreateRoom || !btnJoinRoom || !btnJoinSubmit || !roomIdInput) {
    console.error("Critical DOM elements not found!");
    return;
  }

  // Показываем имя пользователя
  if (user.display_name) {
    if (userName) userName.textContent = user.display_name;
    if (greeting) greeting.textContent = `Привет, ${user.display_name}!`;
  }

  // Выход
  btnLogout.addEventListener("click", async () => {
    try {
      // Вызываем logout на Auth-сервисе
      await fetch(`${AUTH_API_BASE}/auth-api/logout`, {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${accessToken}`,
        },
        credentials: "include", // Включаем cookies для refresh token
      });
    } catch (error) {
      console.error("Logout error:", error);
    }
    
    // Очищаем localStorage
    localStorage.removeItem("accessToken");
    localStorage.removeItem("user");
    localStorage.removeItem("display_name");
    // Refresh token в cookie будет очищен сервером при logout
    
    // Редирект на главную
    if (!isRedirecting) {
      isRedirecting = true;
      window.location.href = "index.html";
    }
  });

  // Создание комнаты
  btnCreateRoom.addEventListener("click", async () => {
    btnCreateRoom.disabled = true;
    btnCreateRoom.textContent = "Создание...";

    try {
      // Создаем комнату через VibeMeet API с JWT токеном
      const response = await fetch(`${VIBEEMEET_API_BASE}/rooms`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          title: `Комната ${user.display_name || "Пользователь"}`,
          description: "Новая видеоконференция",
          max_participants: 10,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: "Ошибка создания комнаты" }));
        throw new Error(errorData.error || "Ошибка создания комнаты");
      }

      const data = await response.json();
      
      // Сохраняем ID комнаты (API возвращает "id", не "room_id")
      localStorage.setItem("current_room_id", data.id);
      
      // Переходим в комнату
      window.location.href = `room.html?room=${data.id}`;
    } catch (error) {
      console.error("Create room error:", error);
      alert("Ошибка создания комнаты: " + error.message);
      btnCreateRoom.disabled = false;
      btnCreateRoom.textContent = "Создать комнату";
    }
  });

  // Показать поле для ввода ID комнаты
  btnJoinRoom.addEventListener("click", () => {
    const joinSection = document.querySelector(".join-room-section");
    
    if (joinSection && joinSection.classList.contains("hidden")) {
      joinSection.classList.remove("hidden");
      setTimeout(() => {
        if (roomIdInput) roomIdInput.focus();
      }, 300);
    }
  });

  // Присоединиться к комнате по ID
  btnJoinSubmit.addEventListener("click", () => {
    const roomId = roomIdInput.value.trim();
    
    if (!roomId) {
      roomIdInput.style.borderColor = "#ff6b6b";
      setTimeout(() => {
        roomIdInput.style.borderColor = "";
      }, 2000);
      return;
    }
    
    // Переходим в комнату
    window.location.href = `room.html?room=${encodeURIComponent(roomId)}`;
  });

  // Enter для присоединения
  roomIdInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      btnJoinSubmit.click();
    }
  });
});
