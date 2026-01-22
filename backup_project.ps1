# Скрипт для резервного копирования проекта на флешку
# Использование: .\backup_project.ps1 -Destination "E:\backup\server_new"

param(
    [Parameter(Mandatory=$true)]
    [string]$Destination
)

$Source = $PSScriptRoot
$ProjectName = Split-Path -Leaf $Source

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Резервное копирование проекта" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Источник: $Source" -ForegroundColor Yellow
Write-Host "Назначение: $Destination\$ProjectName" -ForegroundColor Yellow
Write-Host ""

# Создаём папку назначения
$BackupPath = Join-Path $Destination $ProjectName
if (-not (Test-Path $BackupPath)) {
    New-Item -ItemType Directory -Path $BackupPath -Force | Out-Null
    Write-Host "✓ Создана папка: $BackupPath" -ForegroundColor Green
}

# Список папок и файлов для копирования
$ItemsToCopy = @(
    "config",
    "handlers",
    "middleware",
    "models",
    "services",
    "utils",
    "migrations",
    "docs",
    ".github",
    "main.go",
    "go.mod",
    "go.sum",
    "Dockerfile",
    ".dockerignore",
    ".gitignore"
)

# Копируем файлы и папки
$CopiedCount = 0
$SkippedCount = 0

foreach ($Item in $ItemsToCopy) {
    $SourcePath = Join-Path $Source $Item
    $DestPath = Join-Path $BackupPath $Item
    
    if (Test-Path $SourcePath) {
        try {
            Copy-Item -Path $SourcePath -Destination $DestPath -Recurse -Force
            Write-Host "✓ Скопировано: $Item" -ForegroundColor Green
            $CopiedCount++
        }
        catch {
            Write-Host "✗ Ошибка при копировании $Item : $_" -ForegroundColor Red
        }
    }
    else {
        Write-Host "⊘ Пропущено (не найдено): $Item" -ForegroundColor Gray
        $SkippedCount++
    }
}

# Копируем .env файл отдельно (если есть)
$EnvFile = Join-Path $Source ".env"
if (Test-Path $EnvFile) {
    $EnvBackup = Join-Path $BackupPath ".env.backup"
    Copy-Item -Path $EnvFile -Destination $EnvBackup -Force
    Write-Host "✓ Скопирован .env файл (как .env.backup)" -ForegroundColor Green
    Write-Host "  ВАЖНО: После восстановления переименуйте .env.backup в .env" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Резервное копирование завершено!" -ForegroundColor Green
Write-Host "Скопировано: $CopiedCount элементов" -ForegroundColor Green
Write-Host "Пропущено: $SkippedCount элементов" -ForegroundColor Gray
Write-Host "Путь к резервной копии: $BackupPath" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "ВАЖНО:" -ForegroundColor Yellow
Write-Host "1. Сохраните .env файл отдельно (он содержит ваши настройки)" -ForegroundColor Yellow
Write-Host "2. После восстановления выполните: go mod download" -ForegroundColor Yellow
Write-Host "3. Создайте .env файл на основе .env.example" -ForegroundColor Yellow
