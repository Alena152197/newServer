1.

Скачиваем установщик Go с официального сайта https://go.dev/dl/
Устанавливаем Go по инструкции для своей операционной системы
Проверяем, что Go установлен правильно — запускаем команду go version

# Проверяем установку Go
go version

# Создаём папку проекта
mkdir linkhub_server
cd linkhub_server

# Инициализируем модуль
go mod init linkhub_server


# Создание файла main.go и объявление пакета
package main
(файлы в пакете main можно запускать командой go run)

# Подключение необходимых библиотек
"encoding/json" — для работы с JSON (чтение и запись данных в формате JSON)
"fmt" — для форматирования строк (например, вставка переменных в текст)
"log" — для вывода сообщений в консоль
"net/http" — стандартная библиотека для создания HTTP сервера


import (
	"encoding/json" 
	"fmt"           
	"log"           
	"net/http"  
)

#  Объявление константы для порта
Порт — это номер, по которому сервер будет слушать запросы. Мы выносим его в константу, чтобы не повторять число 4000 по всему коду. Если захотим изменить порт, нужно будет поменять только одно место.

const port = 4000 // порт, на котором будет работать сервер

```
package ...
import ...
const / var (глобальные)
type
func
```

# Создание структуры для ответа на /info

type InfoResponse struct {
	Message string `json:"message"` // тег json говорит, как поле будет называться в JSON
}

# Создание структуры для задачи

Задача — это объект с несколькими полями (ID, название, статус). Структура описывает, какие поля есть у задачи и как они будут выглядеть в JSON.

// type Task struct — создаём структуру с именем Task

// Структура для задачи
type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

#  Создание обработчика для маршрута GET /info

// Обработчик для маршрута GET /info
func infoHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод запроса — GET
	if r.Method != http.MethodGet {
		// Если не GET, отправляем ошибку 405 (Method Not Allowed)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return // выходим из функции
	}

	// Создаём ответ
	response := InfoResponse{
		Message: fmt.Sprintf("Сервер работает на порту %d :)", port),
	}

	// Устанавливаем заголовок, что ответ будет в формате JSON
	w.Header().Set("Content-Type", "application/json")

	// Кодируем ответ в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(response)
}

# Создание обработчика для получения списка задач (GET /tasks)

Скачайте установщик:
Сайт: https://jmeubank.github.io/tdm-gcc/
Или прямая ссылка: https://jmeubank.github.io/tdm-gcc/download/
Выберите последнюю версию (64-bit)
Запустите установщик:
Дважды кликните на скачанный .exe
Следуйте инструкциям (можно оставить настройки по умолчанию)
Перезапустите терминал:
Закройте текущий PowerShell
Откройте новый PowerShell
Проверьте установку:
   gcc --version
Должно показать версию GCC
Запустите сервер:
   $env:CGO_ENABLED=1   go run main.go
Вариант 2: MinGW-w64