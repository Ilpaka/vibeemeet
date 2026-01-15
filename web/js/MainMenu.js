// Главный экран - редирект на Auth или Dashboard
const API_BASE = window.location.origin + "/api/v1";
// Используем относительный путь через nginx proxy
// Nginx проксирует /auth-api/ на auth-сервис localhost:8080/api/v1/auth/
const AUTH_API_BASE = window.location.origin;

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

// Проверка авторизации при загрузке страницы
document.addEventListener("DOMContentLoaded", () => {
  // Защита от множественных редиректов
  if (isRedirecting) {
    return;
  }

  // Если пользователь уже авторизован - редирект на Dashboard
  if (isValidAuth()) {
    isRedirecting = true;
    window.location.href = "Dashboard.html";
    return;
  }

  // Элементы DOM
  const btnLogin = document.getElementById("btn-login");
  const btnRegister = document.getElementById("btn-register");

  // Кнопка "Войти" - редирект на страницу авторизации
  if (btnLogin) {
    btnLogin.addEventListener("click", () => {
      if (!isRedirecting) {
        isRedirecting = true;
        window.location.href = "Auth.html";
      }
    });
  }

  // Кнопка "Регистрация" - редирект на страницу авторизации (вкладка регистрации)
  if (btnRegister) {
    btnRegister.addEventListener("click", () => {
      if (!isRedirecting) {
        isRedirecting = true;
        window.location.href = "Auth.html#register";
      }
    });
  }
});

// FAQ аккордеон
document.addEventListener("DOMContentLoaded", () => {
  const faqQuestions = document.querySelectorAll(".faq-question");
  
  faqQuestions.forEach(question => {
    question.addEventListener("click", () => {
      const faqItem = question.closest(".faq-item");
      const isActive = faqItem.classList.contains("active");
      
      // Закрываем все остальные FAQ
      document.querySelectorAll(".faq-item").forEach(item => {
        if (item !== faqItem) {
          item.classList.remove("active");
        }
      });
      
      // Переключаем текущий FAQ
      faqItem.classList.toggle("active", !isActive);
    });
  });
  
  // Прокрутка к FAQ при клике на стрелку под hero-panel
  const heroScrollIndicator = document.querySelector(".hero-scroll-indicator");
  if (heroScrollIndicator) {
    heroScrollIndicator.addEventListener("click", () => {
      const faqSection = document.querySelector(".faq-section");
      if (faqSection) {
        faqSection.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    });
  }
});