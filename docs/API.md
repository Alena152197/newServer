# API Документация

## Базовый URL
`http://localhost:4000`

## Авторизация
Большинство эндпоинтов требуют токен в заголовке:
```
Authorization: Bearer <токен>
```

Токен получается при успешной авторизации через `/auth/login`.

---

## Эндпоинты

### GET /info
Получить информацию о сервере.

**Заголовки:** Не требуются

**Ответ:** `200 OK`
```json
{
  "message": "Сервер работает на порту 4000"
}
```

---

### POST /auth/register
Регистрация нового пользователя.

**Заголовки:** Не требуются

**Тело запроса:**
```json
{
  "username": "user",
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:** `201 Created`
```json
{
  "id": 1,
  "username": "user",
  "email": "user@example.com"
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат данных или валидация не пройдена
- `409 Conflict` - Пользователь с таким email или username уже существует

---

### POST /auth/login
Авторизация пользователя.

**Заголовки:** Не требуются

**Тело запроса:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:** `200 OK`
```json
{
  "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "user",
    "email": "user@example.com"
  }
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат данных
- `401 Unauthorized` - Неверный email или пароль

---

### POST /auth/reset-simple
Простое восстановление пароля (для обучения).

**Заголовки:** Не требуются

**Тело запроса:**
```json
{
  "email": "user@example.com",
  "newPassword": "newpassword123"
}
```

**Ответ:** `200 OK`
```json
{
  "success": true
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат данных или пароль слишком короткий
- `404 Not Found` - Пользователь не найден

---

### GET /me
Получить информацию о текущем пользователе.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Ответ:** `200 OK`
```json
{
  "id": 1,
  "username": "user",
  "email": "user@example.com"
}
```

**Ошибки:**
- `401 Unauthorized` - Токен недействителен или отсутствует

---

### PUT /me
Обновить профиль текущего пользователя.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Тело запроса:**
```json
{
  "email": "newemail@example.com",
  "currentPassword": "oldpassword",
  "newPassword": "newpassword123"
}
```

**Примечание:** Все поля опциональны. Для смены пароля обязательно указать `currentPassword` и `newPassword`.

**Ответ:** `200 OK`
```json
{
  "success": true
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат данных или валидация не пройдена
- `401 Unauthorized` - Токен недействителен или текущий пароль неверен
- `409 Conflict` - Email уже занят

---

### DELETE /me
Удалить аккаунт текущего пользователя.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Ответ:** `204 No Content` (без тела)

**Ошибки:**
- `401 Unauthorized` - Токен недействителен
- `404 Not Found` - Пользователь не найден

---

### GET /tasks
Получить список задач текущего пользователя.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Параметры запроса:**
- `page` (опционально) - номер страницы (по умолчанию 1)
- `limit` (опционально) - количество на странице (по умолчанию 10, максимум 100)
- `status` (опционально) - фильтр по статусу (`pending`, `in_progress`, `completed`)

**Примеры:**
```
GET /tasks
GET /tasks?page=1&limit=10
GET /tasks?status=pending&page=2&limit=20
```

**Заголовки ответа:**
```
X-Total-Count: 25
```

**Ответ:** `200 OK`
```json
[
  {
    "id": 1,
    "title": "Задача 1",
    "description": "Описание задачи",
    "status": "pending",
    "userid": 1
  },
  {
    "id": 2,
    "title": "Задача 2",
    "description": "Описание задачи 2",
    "status": "in_progress",
    "userid": 1
  }
]
```

**Ошибки:**
- `401 Unauthorized` - Токен недействителен

---

### POST /tasks
Создать новую задачу.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Тело запроса:**
```json
{
  "title": "Новая задача",
  "description": "Описание задачи",
  "status": "pending"
}
```

**Примечание:** Поле `status` опционально, по умолчанию `pending`.

**Ответ:** `201 Created`
```json
{
  "id": 1,
  "title": "Новая задача",
  "description": "Описание задачи",
  "status": "pending",
  "userid": 1
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат данных или валидация не пройдена
- `401 Unauthorized` - Токен недействителен

---

### GET /tasks/:id
Получить задачу по ID.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Ответ:** `200 OK`
```json
{
  "id": 1,
  "title": "Задача",
  "description": "Описание",
  "status": "pending",
  "userid": 1
}
```

**Ошибки:**
- `401 Unauthorized` - Токен недействителен
- `404 Not Found` - Задача не найдена или не принадлежит пользователю

---

### PUT /tasks/:id
Обновить задачу.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Тело запроса:**
```json
{
  "title": "Обновлённая задача",
  "description": "Новое описание",
  "status": "completed"
}
```

**Примечание:** Все поля опциональны. Обновляются только переданные поля.

**Ответ:** `200 OK`
```json
{
  "id": 1,
  "title": "Обновлённая задача",
  "description": "Новое описание",
  "status": "completed",
  "userid": 1
}
```

**Ошибки:**
- `400 Bad Request` - Неверный формат данных
- `401 Unauthorized` - Токен недействителен
- `404 Not Found` - Задача не найдена или не принадлежит пользователю

---

### DELETE /tasks/:id
Удалить задачу.

**Заголовки:**
```
Authorization: Bearer <токен>
```

**Ответ:** `204 No Content` (без тела)

**Ошибки:**
- `401 Unauthorized` - Токен недействителен
- `404 Not Found` - Задача не найдена или не принадлежит пользователю

---

### POST /upload
Загрузить файл на сервер.

**Заголовки:** Не требуются (но можно добавить авторизацию)

**Тело запроса:** `multipart/form-data`
- `file` - файл для загрузки

**Ограничения:**
- Максимальный размер: 10 МБ
- Разрешённые типы: `.jpg`, `.jpeg`, `.png`, `.pdf`

**Ответ:** `201 Created`
```json
{
  "message": "Файл успешно загружен",
  "filename": "photo.png"
}
```

**Ошибки:**
- `400 Bad Request` - Файл слишком большой, неверный формат или недопустимый тип
- `405 Method Not Allowed` - Использован неверный HTTP метод
- `500 Internal Server Error` - Ошибка при сохранении файла

---

## Коды статусов

- `200 OK` - Успешный запрос
- `201 Created` - Ресурс успешно создан
- `204 No Content` - Успешный запрос без тела ответа
- `400 Bad Request` - Неверный формат запроса
- `401 Unauthorized` - Требуется авторизация или токен недействителен
- `404 Not Found` - Ресурс не найден
- `405 Method Not Allowed` - Метод не разрешён для данного эндпоинта
- `409 Conflict` - Конфликт данных (например, email уже занят)
- `429 Too Many Requests` - Слишком много запросов (rate limiting)
- `500 Internal Server Error` - Внутренняя ошибка сервера

---

## Примеры использования

### Получение токена и работа с задачами

```bash
# 1. Регистрация
curl -X POST http://localhost:4000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user","email":"user@example.com","password":"password123"}'

# 2. Авторизация
curl -X POST http://localhost:4000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# 3. Получение задач (с токеном)
curl -X GET http://localhost:4000/tasks?page=1&limit=10 \
  -H "Authorization: Bearer <ваш_токен>"

# 4. Создание задачи
curl -X POST http://localhost:4000/tasks \
  -H "Authorization: Bearer <ваш_токен>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Новая задача","description":"Описание","status":"pending"}'
```

---

