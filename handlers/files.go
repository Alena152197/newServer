package handlers

import (
	"fmt"           // для форматирования строк
	"io"            // для копирования данных
	"net/http"      // для HTTP запросов и ответов
	"os"            // для работы с файловой системой
	"path/filepath" // для работы с путями
	"strings"       // для работы со строками
)

// Максимальный размер загружаемого файла (10 мегабайт)
const maxUploadSize = 10 << 20 // 10 * 1024 * 1024 = 10,485,760 байт

// Папка для хранения загруженных файлов
const uploadDir = "./uploads"

// ensureUploadDir проверяет и создаёт папку для загрузки, если её нет
func ensureUploadDir() error {
	// Проверяем, существует ли папка
	_, err := os.Stat(uploadDir)
	if err != nil {
		// Если папки нет, создаём её
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return fmt.Errorf("не удалось создать папку %s: %v", uploadDir, err)
		}
	}
	return nil
}

// UploadFileHandler обрабатывает загрузку файла на сервер
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод запроса — POST
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён. Используйте POST", http.StatusMethodNotAllowed)
		return // выходим из функции
	}

	// Ограничиваем размер тела запроса
	// Если размер превысит maxUploadSize, чтение автоматически остановится
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Парсим multipart/form-data формат
	// Это нужно, чтобы извлечь файл из запроса
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, "Файл слишком большой или неверный формат", http.StatusBadRequest)
		return
	}

	// Получаем файл из запроса
	// "file" — это имя поля в форме
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Не удалось получить файл из запроса", http.StatusBadRequest)
		return
	}
	defer file.Close() // важно закрыть файл после использования

	// Проверяем расширение файла
	// Извлекаем расширение (часть после точки)
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Список разрешённых расширений
	allowedExts := []string{".jpg", ".jpeg", ".png", ".pdf"}

	// Проверяем, есть ли расширение в списке разрешённых
	allowed := false
	for _, e := range allowedExts {
		if ext == e {
			allowed = true
			break // выходим из цикла, так как нашли совпадение
		}
	}

	// Если расширение не разрешено, отправляем ошибку
	if !allowed {
		http.Error(w, "Недопустимый тип файла. Разрешены: jpg, jpeg, png, pdf", http.StatusBadRequest)
		return
	}

	// Убеждаемся, что папка uploads существует
	if err := ensureUploadDir(); err != nil {
		http.Error(w, "Не удалось создать папку для загрузки", http.StatusInternalServerError)
		return
	}

	// Формируем путь для сохранения файла
	// filepath.Join правильно соединяет путь для любой ОС
	filePath := filepath.Join(uploadDir, header.Filename)

	// Создаём файл на диске для сохранения загруженного файла
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Не удалось создать файл на сервере", http.StatusInternalServerError)
		return
	}
	defer dst.Close() // важно закрыть файл после записи

	// Копируем данные из загруженного файла в файл на диске
	// io.Copy читает из file и записывает в dst
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Ошибка при сохранении файла на диск", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Файл успешно загружен", "filename": "%s"}`, header.Filename)
}
