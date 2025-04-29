package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/yourusername/jsondb/api"
	"github.com/yourusername/jsondb/config"
	"github.com/yourusername/jsondb/storage"
)

func main() {
	// Создать конфигурацию для файлового хранения
	cfg := config.NewFileStorageConfig("./data", true)
	
	// Инициализировать базу данных
	db := api.NewDB()
	
	// Создать и добавить коллекцию 'users'
	userCollectionPath := filepath.Join(cfg.DataDir, "users")
	userStorage, err := storage.NewFileStorage(userCollectionPath, cfg.UseCache)
	if err != nil {
		log.Fatalf("Ошибка создания хранилища: %v", err)
	}
	
	if err := db.CreateCollection("users", userStorage); err != nil {
		log.Fatalf("Ошибка создания коллекции: %v", err)
	}
	
	usersCollection, err := db.GetCollection("users")
	if err != nil {
		log.Fatalf("Ошибка получения коллекции: %v", err)
	}
	
	// Создать индекс по полю 'age'
	if err := usersCollection.CreateIndex("age", "btree", cfg.DefaultBTreeOrder); err != nil {
		log.Fatalf("Ошибка создания индекса: %v", err)
	}
	
	// Вставить несколько документов
	users := []storage.Document{
		{
			ID: "user1",
			Content: map[string]interface{}{
				"name":    "Иван",
				"age":     30,
				"email":   "ivan@example.com",
				"active":  true,
				"roles":   []string{"user", "admin"},
				"address": map[string]interface{}{
					"city":    "Москва",
					"country": "Россия",
				},
			},
		},
		{
			ID: "user2",
			Content: map[string]interface{}{
				"name":    "Мария",
				"age":     25,
				"email":   "maria@example.com",
				"active":  true,
				"roles":   []string{"user"},
				"address": map[string]interface{}{
					"city":    "Санкт-Петербург",
					"country": "Россия",
				},
			},
		},
		{
			ID: "user3",
			Content: map[string]interface{}{
				"name":    "Алексей",
				"age":     40,
				"email":   "alex@example.com",
				"active":  false,
				"roles":   []string{"user"},
				"address": map[string]interface{}{
					"city":    "Новосибирск",
					"country": "Россия",
				},
			},
		},
	}
	
	for _, user := range users {
		if err := usersCollection.InsertDocument(user); err != nil {
			log.Fatalf("Ошибка вставки документа: %v", err)
		}
	}
	
	// Получить документ по ID
	user, err := usersCollection.GetDocument("user1")
	if err != nil {
		log.Fatalf("Ошибка получения документа: %v", err)
	}
	fmt.Println("Получен пользователь по ID:", user.ID)
	fmt.Printf("Имя: %s, Возраст: %d\n", user.Content["name"], int(user.Content["age"].(float64)))
	
	// Найти документы по индексу
	usersAge30, err := usersCollection.FindByIndex("age", 30)
	if err != nil {
		log.Fatalf("Ошибка поиска по индексу: %v", err)
	}
	fmt.Printf("Найдено %d пользователей с возрастом 30\n", len(usersAge30))
	
	// Выполнить запрос
	fmt.Println("\nВыполнение запросов:")
	
	// Запрос: выбрать имя и email всех активных пользователей
	queryStr := "SELECT name, email FROM users WHERE active = true"
	results, err := db.Query(queryStr)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	
	fmt.Printf("Результаты запроса '%s':\n", queryStr)
	for _, result := range results {
		fmt.Printf("  Имя: %s, Email: %s\n", result["name"], result["email"])
	}
	
	// Запрос: выбрать всех пользователей старше 30 лет
	queryStr = "SELECT * FROM users WHERE age > 30"
	results, err = db.Query(queryStr)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	
	fmt.Printf("\nРезультаты запроса '%s':\n", queryStr)
	for _, result := range results {
		fmt.Printf("  ID: %s, Имя: %s, Возраст: %v\n", result["_id"], result["name"], result["age"])
	}
	
	// Запрос с логическими операторами
	queryStr = "SELECT name, age FROM users WHERE age >= 25 AND active = true"
	results, err = db.Query(queryStr)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	
	fmt.Printf("\nРезультаты запроса '%s':\n", queryStr)
	for _, result := range results {
		fmt.Printf("  Имя: %s, Возраст: %v\n", result["name"], result["age"])
	}
	
	// Использование функций для строк в запросе
	// Примечание: в текущей реализации это потребовало бы расширения синтаксиса запросов
	// Для демонстрации используем доступ к API функций
	fmt.Println("\nИспользование функций:")
	
	name := "Ivan"
	upperName := db.Functions.StringFunctions.ToUpper(name)
	fmt.Printf("ToUpper('%s') = '%s'\n", name, upperName)
	
	value := 25.75
	roundedValue := db.Functions.NumberFunctions.Round(value)
	fmt.Printf("Round(%f) = %f\n", value, roundedValue)
	
	currentTime := db.Functions.DateFunctions.Now()
	formattedTime := db.Functions.DateFunctions.Format(currentTime, "2006-01-02 15:04:05")
	fmt.Printf("Current time: %s\n", formattedTime)
	
	// Обновить документ
	user.Content["age"] = 31
	if err := usersCollection.UpdateDocument(user); err != nil {
		log.Fatalf("Ошибка обновления документа: %v", err)
	}
	fmt.Println("\nВозраст пользователя обновлен до 31")
	
	// Удалить документ
	if err := usersCollection.DeleteDocument("user3"); err != nil {
		log.Fatalf("Ошибка удаления документа: %v", err)
	}
	fmt.Println("Пользователь user3 удален")
	
	// Получить количество документов в коллекции
	count, err := usersCollection.Size()
	if err != nil {
		log.Fatalf("Ошибка получения размера: %v", err)
	}
	fmt.Printf("Количество документов в коллекции: %d\n", count)
}
