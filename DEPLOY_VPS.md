# Полная инструкция по деплою Go приложения на Jino VPS

## Шаг 1: Подготовка сервера

```bash
# Обновляем пакеты
sudo apt update && sudo apt -y upgrade

# Устанавливаем Go 1.23.3 (ВАЖНО: версия должна быть совместима с go.mod)
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Проверяем версию Go
go version
# Должно показать: go version go1.23.3 linux/amd64

# Устанавливаем Git (если ещё не установлен)
sudo apt -y install git

# Устанавливаем gcc и libc6-dev для CGO (ОБЯЗАТЕЛЬНО для SQLite!)
sudo apt -y install gcc libc6-dev
```

**⚠️ ВАЖНО:** 
- Используйте Go 1.23.3 (не 1.25, так как `golang.org/x/crypto v0.26.0` не поддерживает Go 1.24+)
- Обязательно установите `gcc` и `libc6-dev` для работы с SQLite через CGO

## Шаг 2: Клонирование и подготовка проекта

```bash
# Клонируем репозиторий
git clone https://github.com/Alena152197/newServer.git /opt/server_new
cd /opt/server_new

# Обновляем зависимости
go mod tidy
```

**⚠️ ВАЖНО:** 
- Убедитесь, что `go.mod` содержит `go 1.23` (не 1.25)
- Убедитесь, что `golang.org/x/crypto` версии `v0.26.0` (совместима с Go 1.23)

## Шаг 3: Сборка приложения

```bash
# ОБЯЗАТЕЛЬНО: Собираем с CGO_ENABLED=1 (для SQLite)
CGO_ENABLED=1 go build -buildvcs=false -o server_new .

# Проверяем размер бинарника (должен быть ~13MB, а не ~9MB)
ls -lh server_new
```

**⚠️ ВАЖНО:** 
- ВСЕГДА используйте `CGO_ENABLED=1` при сборке (SQLite требует CGO)
- Без CGO бинарник будет показывать ошибку: "Binary was compiled with 'CGO_ENABLED=0'"

## Шаг 4: Создание папок и файла конфигурации

```bash
# Создаём необходимые папки
sudo mkdir -p /opt/server_new/.tmp
sudo mkdir -p /opt/server_new/uploads

# Создаём файл .env
sudo nano /opt/server_new/.env
```

**Содержимое файла `.env`:**

```env
PORT=4000
JWT_SECRET=your-super-secret-key-change-me-in-production-please-use-strong-random-key
DB_PATH=.tmp/base.sqlite
ALLOWED_ORIGINS=http://c86728771394.vps.myjino.ru,https://c86728771394.vps.myjino.ru
LOG_LEVEL=info
```

**⚠️ ВАЖНО:**
- В `JWT_SECRET` укажите случайный длинный ключ (можно сгенерировать: `openssl rand -hex 32`)
- В `ALLOWED_ORIGINS` укажите ваш домен или IP адрес
- В `DB_PATH` используйте относительный путь `.tmp/base.sqlite` (будет создан в рабочей директории)

**Сохраните:** `Ctrl+O`, Enter, `Ctrl+X`

## Шаг 5: Установка прав доступа

```bash
# Даём права пользователю www-data на всю директорию
sudo chown -R www-data:www-data /opt/server_new

# Проверяем права на бинарник
sudo chmod +x /opt/server_new/server_new

# Проверяем права на .env файл
sudo chmod 644 /opt/server_new/.env

# Проверяем, что www-data может читать .env
sudo -u www-data cat /opt/server_new/.env
```

**⚠️ ВАЖНО:** 
- Все файлы должны принадлежать `www-data:www-data`
- Бинарник должен быть исполняемым (`chmod +x`)
- `.env` должен быть читаемым для `www-data`

## Шаг 6: Создание systemd сервиса

```bash
# Создаём сервис
sudo nano /etc/systemd/system/server_new.service
```

**Содержимое файла `/etc/systemd/system/server_new.service`:**

```ini
[Unit]
Description=Server New API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/server_new
ExecStart=/opt/server_new/server_new
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**⚠️ ВАЖНО:**
- `WorkingDirectory=/opt/server_new` — рабочая директория для относительных путей (`.tmp`, `.env`)
- `User=www-data` — пользователь, от имени которого запускается сервис
- `Restart=always` — автоматический перезапуск при падении

**Сохраните:** `Ctrl+O`, Enter, `Ctrl+X`

```bash
# Перезагружаем systemd
sudo systemctl daemon-reload

# Включаем автозапуск
sudo systemctl enable server_new

# Запускаем сервис
sudo systemctl start server_new

# Проверяем статус
sudo systemctl status server_new

# Если всё хорошо, должно показать: Active: active (running)
```

## Шаг 7: Проверка работы сервера

```bash
# Проверяем логи
sudo journalctl -u server_new -n 20 --no-pager

# Проверяем, что сервер слушает порт 4000
sudo ss -tlnp | grep :4000

# Проверяем работу API
curl http://localhost:4000/info
# Должен вернуть: {"message":"Сервер работает на порту 4000"}
```

## Шаг 8: Настройка Nginx как reverse proxy

```bash
# Устанавливаем Nginx
sudo apt -y install nginx

# Создаём конфиг
sudo nano /etc/nginx/sites-available/server_new
```

**Содержимое файла `/etc/nginx/sites-available/server_new`:**

```nginx
server {
    listen 80;
    server_name c86728771394.vps.myjino.ru; # Замените на ваш домен или IP

    location / {
        proxy_pass http://127.0.0.1:4000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**⚠️ ВАЖНО:** 
- Замените `server_name` на ваш домен или IP адрес

**Сохраните:** `Ctrl+O`, Enter, `Ctrl+X`

```bash
# Активируем конфиг
sudo ln -s /etc/nginx/sites-available/server_new /etc/nginx/sites-enabled/

# Проверяем конфигурацию Nginx
sudo nginx -t

# Перезагружаем Nginx
sudo systemctl reload nginx
```

## Шаг 9: SSL сертификат (опционально, если есть домен)

```bash
# Устанавливаем certbot
sudo apt -y install certbot python3-certbot-nginx

# Получаем сертификат (замените на ваш домен)
sudo certbot --nginx -d c86728771394.vps.myjino.ru

# Certbot автоматически настроит Nginx для HTTPS
```

## Шаг 10: Проверка финальной работы

```bash
# Проверяем доступность через Nginx
curl http://c86728771394.vps.myjino.ru/info
# или
curl https://c86728771394.vps.myjino.ru/info

# Проверяем статус сервиса
sudo systemctl status server_new

# Проверяем логи в реальном времени
sudo journalctl -u server_new -f
```

## Чек-лист успешного деплоя

- [ ] Go 1.23.3 установлен и доступен (`go version`)
- [ ] `gcc` и `libc6-dev` установлены
- [ ] Репозиторий клонирован в `/opt/server_new`
- [ ] Бинарник собран с `CGO_ENABLED=1` (размер ~13MB)
- [ ] Папки `.tmp` и `uploads` созданы
- [ ] Файл `.env` создан и настроен
- [ ] Права установлены для `www-data:www-data`
- [ ] Systemd сервис создан и запущен (`Active: active (running)`)
- [ ] Сервер отвечает на `curl http://localhost:4000/info`
- [ ] Nginx настроен и работает
- [ ] SSL сертификат получен (если есть домен)

## Решение проблем

### Ошибка: "Binary was compiled with 'CGO_ENABLED=0'"
**Решение:** Пересоберите бинарник с `CGO_ENABLED=1`:
```bash
CGO_ENABLED=1 go build -buildvcs=false -o server_new .
```

### Ошибка: "permission denied" при создании папки `.tmp`
**Решение:** Установите права для `www-data`:
```bash
sudo chown -R www-data:www-data /opt/server_new
sudo chmod 755 /opt/server_new/.tmp
```

### Ошибка: "Файл .env не найден"
**Решение:** Проверьте, что файл существует и `www-data` может его читать:
```bash
sudo -u www-data cat /opt/server_new/.env
```

### Ошибка: "go.mod requires go >= 1.24.0"
**Решение:** Измените `go.mod` на `go 1.23` и обновите зависимости:
```bash
sed -i 's/go 1.25/go 1.23/' go.mod
go mod tidy
```

## Команды для обновления приложения

```bash
# Останавливаем сервис
sudo systemctl stop server_new

# Переходим в директорию проекта
cd /opt/server_new

# Получаем последние изменения
git pull

# Обновляем зависимости
go mod tidy

# Пересобираем приложение
CGO_ENABLED=1 go build -buildvcs=false -o server_new .

# Устанавливаем права
sudo chown www-data:www-data /opt/server_new/server_new
sudo chmod +x /opt/server_new/server_new

# Запускаем сервис
sudo systemctl start server_new

# Проверяем статус
sudo systemctl status server_new
```

---

**Готово!** Ваше приложение должно быть доступно по адресу вашего VPS сервера.
