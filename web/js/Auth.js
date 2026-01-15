// Auth.js - Страница входа/регистрации

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

// Проверка валидности авторизации
function isValidAuth() {
  const accessToken = localStorage.getItem("accessToken");
  const user = safeParseJSON("user", {});
  
  // Проверяем наличие токена и валидных данных пользователя
  return accessToken && user && user.id && typeof user.id !== 'undefined';
}

// Инициализация при загрузке DOM
document.addEventListener("DOMContentLoaded", () => {
  // Проверяем авторизацию только после загрузки DOM
  if (isValidAuth() && !isRedirecting) {
    isRedirecting = true;
    window.location.href = "Dashboard.html";
    return;
  }

  // Элементы DOM
  const btnBack = document.getElementById("btn-back");
  const authTabs = document.querySelectorAll(".auth-tab");
  const loginForm = document.getElementById("login-form");
  const registerForm = document.getElementById("register-form");

  // Вход
  const loginEmail = document.getElementById("login-email");
  const loginPassword = document.getElementById("login-password");
  const btnLogin = document.getElementById("btn-login");
  const loginError = document.getElementById("login-error");

  // Регистрация
  const registerName = document.getElementById("register-name");
  const registerEmail = document.getElementById("register-email");
  const registerPassword = document.getElementById("register-password");
  const registerPasswordConfirm = document.getElementById("register-password-confirm");
  const btnRegister = document.getElementById("btn-register");
  const registerError = document.getElementById("register-error");

  // Проверяем наличие всех элементов
  if (!btnBack || !authTabs.length || !loginForm || !registerForm || 
      !loginEmail || !loginPassword || !btnLogin || !loginError ||
      !registerName || !registerEmail || !registerPassword || 
      !registerPasswordConfirm || !btnRegister || !registerError) {
    console.error("Critical DOM elements not found in Auth page!");
    return;
  }

  // Переключение вкладок
  authTabs.forEach(tab => {
    tab.addEventListener("click", () => {
      const targetTab = tab.dataset.tab;
      
      // Обновляем активную вкладку
      authTabs.forEach(t => t.classList.remove("active"));
      tab.classList.add("active");
      
      // Показываем нужную форму
      if (targetTab === "login") {
        loginForm.classList.add("active");
        registerForm.classList.remove("active");
        clearErrors(loginError, registerError);
      } else {
        registerForm.classList.add("active");
        loginForm.classList.remove("active");
        clearErrors(loginError, registerError);
      }
    });
  });

  // Возврат на главную
  btnBack.addEventListener("click", () => {
    if (!isRedirecting) {
      isRedirecting = true;
      window.location.href = "index.html";
    }
  });

  // === ВХОД ===
  btnLogin.addEventListener("click", async () => {
  const email = loginEmail.value.trim();
  const password = loginPassword.value;

  if (!email || !password) {
    showError(loginError, "Заполните все поля");
    return;
  }

  btnLogin.disabled = true;
  btnLogin.textContent = "Вход...";

  try {
    const response = await fetch(`${AUTH_API_BASE}/auth-api/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include", // Включаем cookies для refresh token
      body: JSON.stringify({ email, password }),
    });

    let data;
    try {
      data = await response.json();
    } catch (parseError) {
      console.error("Failed to parse response:", parseError);
      showError(loginError, "Ошибка ответа сервера");
      btnLogin.disabled = false;
      btnLogin.textContent = "Войти";
      return;
    }

    if (!response.ok) {
      showError(loginError, data.error || "Неверный email или пароль");
      btnLogin.disabled = false;
      btnLogin.textContent = "Войти";
      return;
    }

    // Сохраняем токены и данные пользователя
    const saved = saveAuthData(data);
    if (!saved) {
      showError(loginError, "Ошибка сохранения данных пользователя");
      btnLogin.disabled = false;
      btnLogin.textContent = "Войти";
      return;
    }
    
    // Переходим в dashboard
    if (!isRedirecting) {
      isRedirecting = true;
      window.location.href = "Dashboard.html";
    }
  } catch (error) {
    console.error("Login error:", error);
    showError(loginError, "Ошибка подключения к серверу");
    btnLogin.disabled = false;
    btnLogin.textContent = "Войти";
  }
  });

  // === РЕГИСТРАЦИЯ ===
  btnRegister.addEventListener("click", async () => {
  const name = registerName.value.trim();
  const email = registerEmail.value.trim();
  const password = registerPassword.value;
  const passwordConfirm = registerPasswordConfirm.value;

  // Валидация
  if (!name || !email || !password || !passwordConfirm) {
    showError(registerError, "Заполните все поля");
    return;
  }

  if (password.length < 8) {
    showError(registerError, "Пароль должен быть минимум 8 символов");
    return;
  }

  if (password !== passwordConfirm) {
    showError(registerError, "Пароли не совпадают");
    return;
  }

  if (!isValidEmail(email)) {
    showError(registerError, "Неверный формат email");
    return;
  }

  btnRegister.disabled = true;
  btnRegister.textContent = "Регистрация...";

  try {
    // Разделяем имя на first_name и last_name
    const nameParts = name.trim().split(/\s+/);
    const firstName = nameParts[0] || name;
    const lastName = nameParts.slice(1).join(" ") || "";

    const response = await fetch(`${AUTH_API_BASE}/auth-api/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include", // Включаем cookies для refresh token
      body: JSON.stringify({
        email,
        password,
        first_name: firstName,
        last_name: lastName,
      }),
    });

    let data;
    try {
      data = await response.json();
    } catch (parseError) {
      console.error("Failed to parse response:", parseError);
      showError(registerError, "Ошибка ответа сервера");
      btnRegister.disabled = false;
      btnRegister.textContent = "Зарегистрироваться";
      return;
    }

    if (!response.ok) {
      showError(registerError, data.error || "Ошибка регистрации");
      btnRegister.disabled = false;
      btnRegister.textContent = "Зарегистрироваться";
      return;
    }

    // Сохраняем токены и данные пользователя
    const saved = saveAuthData(data);
    if (!saved) {
      showError(registerError, "Ошибка сохранения данных пользователя");
      btnRegister.disabled = false;
      btnRegister.textContent = "Зарегистрироваться";
      return;
    }
    
    // Переходим в dashboard
    if (!isRedirecting) {
      isRedirecting = true;
      window.location.href = "Dashboard.html";
    }
  } catch (error) {
    console.error("Registration error:", error);
    showError(registerError, "Ошибка подключения к серверу");
    btnRegister.disabled = false;
    btnRegister.textContent = "Зарегистрироваться";
  }
  });

  // === УТИЛИТЫ ===

function saveAuthData(data) {
  if (!data || !data.access_token) {
    console.error("Invalid auth data received - missing access_token");
    return false;
  }
  
  localStorage.setItem("accessToken", data.access_token);
  
  // Refresh token приходит в HttpOnly cookie от auth-сервиса
  // Браузер автоматически будет отправлять его с запросами
  // Не нужно сохранять в localStorage - cookie управляется браузером
  
  // Безопасно сохраняем данные пользователя
  if (data.user && typeof data.user === 'object' && data.user.id) {
    // Создаем display_name из first_name и last_name, если есть
    let displayName = "";
    if (data.user.first_name || data.user.last_name) {
      const parts = [];
      if (data.user.first_name) parts.push(data.user.first_name);
      if (data.user.last_name) parts.push(data.user.last_name);
      displayName = parts.join(" ").trim();
    }
    
    // Если display_name не создан, используем email
    if (!displayName && data.user.email) {
      displayName = data.user.email.split("@")[0];
    }
    
    // Сохраняем user объект с добавленным display_name
    const userData = {
      id: data.user.id,
      email: data.user.email,
      email_verified: data.user.email_verified || false,
      first_name: data.user.first_name || null,
      last_name: data.user.last_name || null,
      display_name: displayName
    };
    
    localStorage.setItem("user", JSON.stringify(userData));
    localStorage.setItem("display_name", displayName);
    
    return true;
  } else {
    console.error("Invalid user data in auth response:", data.user);
    return false;
  }
}

function showError(element, message) {
  element.textContent = message;
  element.classList.remove("hidden");
  setTimeout(() => {
    element.classList.add("hidden");
  }, 5000);
}

function clearErrors(loginError, registerError) {
  if (loginError) loginError.classList.add("hidden");
  if (registerError) registerError.classList.add("hidden");
}

function isValidEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(email);
}

  // Enter для отправки форм
  loginPassword.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      btnLogin.click();
    }
  });

  registerPasswordConfirm.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      btnRegister.click();
    }
  });

  // Проверяем хеш URL для открытия нужной вкладки
  if (window.location.hash === "#register") {
    // Открываем вкладку регистрации
    const registerTab = document.querySelector('[data-tab="register"]');
    if (registerTab) {
      registerTab.click();
    }
  }
});
