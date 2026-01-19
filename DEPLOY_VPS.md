# Деплой на VPS: что и в каком порядке делать 🚀

## 1) Подготовка сервера

```bash
# Обновляем пакеты
sudo apt update && sudo apt -y upgrade

# Устанавливаем Go (версия 1.25 для вашего проекта)
wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Устанавливаем Git
sudo apt -y install git

# Клонируем репозиторий (замените <repo_url> на ваш URL)
git clone <repo_url> /opt/server_new
cd /opt/server_new
```

## 2) Сборка и запуск

```bash
# Собираем приложение
go build -o server_new .

# Создаём systemd сервис
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

```bash
# Запускаем сервис
sudo systemctl daemon-reload
sudo systemctl enable server_new
sudo systemctl start server_new

# Проверяем статус
sudo systemctl status server_new
```

## 3) Nginx как reverse proxy

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
    server_name api.example.com;

    location / {
        proxy_pass http://127.0.0.1:4000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Активируем конфиг
sudo ln -s /etc/nginx/sites-available/server_new /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 4) SSL сертификат

```bash
# Устанавливаем certbot
sudo apt -y install certbot python3-certbot-nginx

# Получаем сертификат (замените api.example.com на ваш домен)
sudo certbot --nginx -d api.example.com
```

## 5) Чек-лист

- [ ] Go установлен и приложение собрано
- [ ] systemd сервис создан и запущен
- [ ] Nginx настроен как reverse proxy
- [ ] SSL сертификат получен
- [ ] Приложение доступно по HTTPS
- [ ] Логи проверены (`sudo journalctl -u server_new -f`)

---

## Что было изучено

- Как подготовить VPS и установить Go
- Как настроить systemd для автозапуска приложения
- Как настроить Nginx как reverse proxy и получить SSL сертификат
