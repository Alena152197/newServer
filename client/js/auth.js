// Базовый URL API (можно настроить через переменную окружения)
const API_BASE_URL = window.location.origin;

/**
 * Выполняет запрос к API
 */
async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // Добавляем токен, если он есть
    const token = localStorage.getItem('token');
    if (token) {
        defaultOptions.headers['Authorization'] = `Bearer ${token}`;
    }

    const config = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers,
        },
    };

    try {
        const response = await fetch(url, config);
        const data = await response.json();

        if (!response.ok) {
            return {
                success: false,
                error: data.error || `Ошибка: ${response.status}`,
            };
        }

        return {
            success: true,
            ...data,
        };
    } catch (error) {
        return {
            success: false,
            error: error.message || 'Ошибка сети',
        };
    }
}

/**
 * Регистрация нового пользователя
 */
async function register(username, email, password) {
    return await apiRequest('/auth/register', {
        method: 'POST',
        body: JSON.stringify({
            username,
            email,
            password,
        }),
    });
}

/**
 * Вход в систему
 */
async function login(email, password) {
    return await apiRequest('/auth/login', {
        method: 'POST',
        body: JSON.stringify({
            email,
            password,
        }),
    });
}

/**
 * Получение информации о текущем пользователе
 */
async function getCurrentUser() {
    const token = localStorage.getItem('token');
    if (!token) {
        return {
            success: false,
            error: 'Токен не найден',
        };
    }

    return await apiRequest('/me', {
        method: 'GET',
    });
}

/**
 * Выход из системы
 */
function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = 'index.html';
}

/**
 * Проверка, авторизован ли пользователь
 */
function isAuthenticated() {
    return !!localStorage.getItem('token');
}

/**
 * Получение сохранённого пользователя
 */
function getStoredUser() {
    const userStr = localStorage.getItem('user');
    if (userStr) {
        try {
            return JSON.parse(userStr);
        } catch (e) {
            return null;
        }
    }
    return null;
}
