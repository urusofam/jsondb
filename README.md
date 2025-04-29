# JSONDB - Легковесная СУБД для JSON документов на Go

JSONDB - простая и эффективная система управления базами данных, предназначенная для хранения, управления и запроса JSON-документов на языке Go.

## Особенности

- **Гибкие варианты хранения**: хранение в памяти или в файловой системе
- **Индексация на основе B-дерева**: для быстрого доступа к данным
- **SQL-подобный язык запросов**: для извлечения данных
- **Функции для работы с данными**: строками, числами и датами
- **Простой API**: для легкой интеграции с вашими приложениями
- **Конкурентная безопасность**: блокировки чтения/записи для безопасной многопоточной работы

## Установка

```bash
go get github.com/yourusername/jsondb
```

## Структура проекта

```
/jsondb
    /api       - Основной API для работы с БД
    /config    - Конфигурация БД
    /index     - Реализация индексов (B-дерево)
    /query     - Парсер и исполнитель запросов
    /storage   - Механизмы хранения данных
```

## Использование

### Создание базы данных

```go
import (
    "github.com/yourusername/jsondb/api"
    "github.com/yourusername/jsondb/config"
    "github.com/yourusername/jsondb/storage"
)

// Создать конфигурацию
cfg := config.NewFileStorageConfig("./data", true)

// Инициализировать БД
db := api.NewDB()

// Создать хранилище для коллекции
userStorage, err := storage.NewFileStorage("./data/users", cfg.UseCache)
if err != nil {
    log.Fatal(err)
}

// Создать коллекцию
err = db.CreateCollection("users", userStorage)
if err != nil {
    log.Fatal(err)
}
```

### Вставка документов

```go
// Получить коллекцию
usersCollection, err := db.GetCollection("users")
if err != nil {
    log.Fatal(err)
}

// Вставить документ
doc := storage.Document{
    ID: "user1",
    Content: map[string]interface{}{
        "name":  "Иван",
        "age":   30,
        "email": "ivan@example.com",
    },
}

err = usersCollection.InsertDocument(doc)
if err != nil {
    log.Fatal(err)
}
```

### Создание индекса

```go
// Создать индекс по полю "age"
err = usersCollection.CreateIndex("age", "btree", 5)
if err != nil {
    log.Fatal(err)
}
```

### Запросы

```go
// Выполнить SQL-подобный запрос
query := "SELECT name, email FROM users WHERE age > 25"
results, err := db.Query(query)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Имя: %s, Email: %s\n", result["name"], result["email"])
}
```

### Поиск по индексам

```go
// Найти документы по индексу
users, err := usersCollection.FindByIndex("age", 30)
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    fmt.Printf("Найден пользователь: %s\n", user.ID)
}
```

### Обновление и удаление

```go
// Получить документ
doc, err := usersCollection.GetDocument("user1")
if err != nil {
    log.Fatal(err)
}

// Обновить документ
doc.Content["age"] = 31
err = usersCollection.UpdateDocument(doc)
if err != nil {
    log.Fatal(err)
}

// Удалить документ
err = usersCollection.DeleteDocument("user1")
if err != nil {
    log.Fatal(err)
}
```

## Поддерживаемые операции в запросах

### Операторы выбора
- `SELECT` - выбор полей
- `FROM` - указание коллекции
- `WHERE` - фильтрация результатов
- `LIMIT` - ограничение количества результатов
- `OFFSET` - смещение результатов

### Операторы сравнения
- `=` - равно
- `!=` - не равно
- `>` - больше
- `<` - меньше
- `>=` - больше или равно
- `<=` - меньше или равно

### Логические операторы
- `AND` - логическое И
- `OR` - логическое ИЛИ

### Примеры запросов

```sql
-- Выбрать все поля для всех пользователей
SELECT * FROM users

-- Выбрать только имя и email активных пользователей
SELECT name, email FROM users WHERE active = true

-- Выбрать пользователей старше 30 лет
SELECT * FROM users WHERE age > 30

-- Сложные условия
SELECT name, age FROM users WHERE age > 25 AND active = true

-- С ограничением количества результатов
SELECT * FROM users LIMIT 10

-- С ограничением и смещением
SELECT * FROM users LIMIT 10 OFFSET 20
```

## Функции для работы с данными

JSONDB предоставляет набор функций для работы с различными типами данных:

### Строковые функции
- `Length(string)` - длина строки
- `ToUpper(string)` - преобразование в верхний регистр
- `ToLower(string)` - преобразование в нижний регистр
- `Substring(string, start, length)` - извлечение подстроки
- `Replace(string, old, new)` - замена подстрок
- `Match(string, pattern)` - проверка соответствия регулярному выражению

### Числовые функции
- `Abs(number)` - абсолютное значение
- `Round(number)` - округление до ближайшего целого
- `Ceil(number)` - округление вверх
- `Floor(number)` - округление вниз
- `Pow(x, y)` - возведение в степень
- `Sqrt(number)` - квадратный корень
- `Min(a, b)` - минимальное значение
- `Max(a, b)` - максимальное значение

### Функции для работы с датами
- `Parse(layout, value)` - разбор строки даты
- `Format(time, layout)` - форматирование даты
- `Now()` - текущее время
- `AddDays(time, days)` - добавление дней
- `AddMonths(time, months)` - добавление месяцев
- `AddYears(time, years)` - добавление лет
- `DaysBetween(time1, time2)` - количество дней между датами
- `MonthsBetween(time1, time2)` - количество месяцев между датами
- `YearsBetween(time1, time2)` - количество лет между датами

## Расширение функциональности

### Добавление нового типа хранения

Для добавления нового типа хранения необходимо реализовать интерфейс `storage.Storage`:

```go
type Storage interface {
    Save(doc Document) error
    Get(id string) (Document, error)
    Delete(id string) error
    List() ([]Document, error)
}
```

### Добавление нового типа индекса

Для добавления нового типа индекса необходимо реализовать интерфейс `index.Index`:

```go
type Index interface {
    Add(doc storage.Document) error
    Remove(id string) error
    Search(field string, value interface{}) ([]string, error)
}
```

### Расширение языка запросов

Для расширения языка запросов можно модифицировать `query.QueryParser` и `query.QueryExecutor`.

## Ограничения

- Отсутствие поддержки транзакций
- Ограниченная поддержка вложенных запросов
- Ограниченная поддержка функций в запросах
- Только базовая реализация B-дерева для индексов

## Дальнейшее развитие

- Добавление поддержки транзакций
- Улучшение производительности индексов
- Расширение языка запросов
- Добавление поддержки других механизмов хранения (например, ключ-значение, сетевой)
- Добавление поддержки репликации и шардинга

## Лицензия

MIT
