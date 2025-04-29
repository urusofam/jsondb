package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urusofam/jsondb/api"
	"github.com/urusofam/jsondb/config"
	"github.com/urusofam/jsondb/storage"
)

// CLI представляет интерактивный командный интерфейс
type CLI struct {
	DB         *api.DB
	Config     *config.DBConfig
	Reader     *bufio.Reader
	CurrentDir string
}

// NewCLI создает новый экземпляр CLI
func NewCLI(config *config.DBConfig) *CLI {
	return &CLI{
		DB:         api.NewDB(),
		Config:     config,
		Reader:     bufio.NewReader(os.Stdin),
		CurrentDir: config.DataDir,
	}
}

// Run запускает интерактивный цикл обработки команд
func (cli *CLI) Run() {
	fmt.Println("=== JSONDB CLI ===")
	fmt.Println("Введите 'help' для просмотра доступных команд")
	fmt.Println("Введите 'exit' для выхода")
	fmt.Println()

	// Убедимся, что директория для данных существует
	if err := os.MkdirAll(cli.Config.DataDir, 0755); err != nil {
		fmt.Printf("Ошибка при создании директории данных: %v\n", err)
		return
	}

	for {
		fmt.Print("> ")
		input, err := cli.Reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Ошибка чтения ввода: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" {
			break
		}

		// Разделить ввод на команду и аргументы
		parts := strings.SplitN(input, " ", 2)
		command := strings.ToLower(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		// Обработать команду
		if err := cli.handleCommand(command, args); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
		}
	}

	fmt.Println("Выход из программы")
}

// handleCommand обрабатывает команду с аргументами
func (cli *CLI) handleCommand(command, args string) error {
	switch command {
	case "help":
		cli.printHelp()
	case "ls", "list":
		return cli.listCommand(args)
	case "mkdir":
		return cli.mkdirCommand(args)
	case "create-collection":
		return cli.createCollectionCommand(args)
	case "list-collections":
		return cli.listCollectionsCommand()
	case "drop-collection":
		return cli.dropCollectionCommand(args)
	case "insert":
		return cli.insertDocumentCommand(args)
	case "get":
		return cli.getDocumentCommand(args)
	case "update":
		return cli.updateDocumentCommand(args)
	case "delete":
		return cli.deleteDocumentCommand(args)
	case "list-docs":
		return cli.listDocumentsCommand(args)
	case "create-index":
		return cli.createIndexCommand(args)
	case "drop-index":
		return cli.dropIndexCommand(args)
	case "query":
		return cli.queryCommand(args)
	default:
		return fmt.Errorf("неизвестная команда: %s", command)
	}
	return nil
}

// printHelp выводит справку по командам
func (cli *CLI) printHelp() {
	fmt.Println("Доступные команды:")
	fmt.Println("  help                               - показать эту справку")
	fmt.Println("  exit                               - выйти из программы")
	fmt.Println("  ls [path]                          - показать файлы в текущей директории или по указанному пути")
	fmt.Println("  mkdir <dir>                        - создать директорию")
	fmt.Println("  create-collection <name>           - создать новую коллекцию")
	fmt.Println("  list-collections                   - показать все коллекции")
	fmt.Println("  drop-collection <name>             - удалить коллекцию")
	fmt.Println("  insert <collection> <json>         - вставить документ в коллекцию")
	fmt.Println("  get <collection> <id>              - получить документ по ID")
	fmt.Println("  update <collection> <id> <json>    - обновить документ")
	fmt.Println("  delete <collection> <id>           - удалить документ")
	fmt.Println("  list-docs <collection> [limit]     - показать документы в коллекции")
	fmt.Println("  create-index <collection> <field>  - создать индекс по полю")
	fmt.Println("  drop-index <collection> <field>    - удалить индекс")
	fmt.Println("  query <sql>                        - выполнить SQL-подобный запрос")
	fmt.Println()
	fmt.Println("Примеры:")
	fmt.Println("  create-collection users")
	fmt.Println("  insert users {\"_id\":\"user1\",\"name\":\"Иван\",\"age\":30,\"email\":\"ivan@example.com\"}")
	fmt.Println("  create-index users age")
	fmt.Println("  query SELECT * FROM users WHERE age > 25")
}

// listCommand выводит список файлов
func (cli *CLI) listCommand(path string) error {
	if path == "" {
		path = cli.CurrentDir
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	fmt.Printf("Содержимое директории %s:\n", path)
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("  [DIR] %s\n", entry.Name())
		} else {
			info, _ := entry.Info()
			size := info.Size()
			fmt.Printf("  [FILE] %s (%d байт)\n", entry.Name(), size)
		}
	}

	return nil
}

// mkdirCommand создает директорию
func (cli *CLI) mkdirCommand(dir string) error {
	if dir == "" {
		return fmt.Errorf("требуется указать имя директории")
	}

	return os.MkdirAll(dir, 0755)
}

// createCollectionCommand создает новую коллекцию
func (cli *CLI) createCollectionCommand(name string) error {
	if name == "" {
		return fmt.Errorf("требуется указать имя коллекции")
	}

	// Проверка, существует ли уже коллекция
	_, err := cli.DB.GetCollection(name)
	if err == nil {
		return fmt.Errorf("коллекция %s уже существует", name)
	}

	// Создать директорию для коллекции
	collectionPath := filepath.Join(cli.Config.DataDir, name)
	if err := os.MkdirAll(collectionPath, 0755); err != nil {
		return err
	}

	// Создать хранилище
	var collectionStorage storage.Storage
	if cli.Config.StorageType == config.StorageTypeFile {
		fs, err := storage.NewFileStorage(collectionPath, cli.Config.UseCache)
		if err != nil {
			return err
		}
		collectionStorage = fs
	} else {
		collectionStorage = storage.NewMemoryStorage()
	}

	// Создать коллекцию
	if err := cli.DB.CreateCollection(name, collectionStorage); err != nil {
		return err
	}

	fmt.Printf("Коллекция %s успешно создана\n", name)
	return nil
}

// listCollectionsCommand выводит список коллекций
func (cli *CLI) listCollectionsCommand() error {
	collections := make([]string, 0, len(cli.DB.Collections))
	for name := range cli.DB.Collections {
		collections = append(collections, name)
	}

	if len(collections) == 0 {
		fmt.Println("Коллекции отсутствуют")
		return nil
	}

	fmt.Println("Доступные коллекции:")
	for _, name := range collections {
		collection, _ := cli.DB.GetCollection(name)
		count, _ := collection.Size()
		fmt.Printf("  - %s (%d документов)\n", name, count)
	}

	return nil
}

// dropCollectionCommand удаляет коллекцию
func (cli *CLI) dropCollectionCommand(name string) error {
	if name == "" {
		return fmt.Errorf("требуется указать имя коллекции")
	}

	// Проверка, существует ли коллекция
	_, err := cli.DB.GetCollection(name)
	if err != nil {
		return err
	}

	// Удалить коллекцию из БД
	if err := cli.DB.DropCollection(name); err != nil {
		return err
	}

	// Если используется файловое хранилище, удалить файлы
	if cli.Config.StorageType == config.StorageTypeFile {
		collectionPath := filepath.Join(cli.Config.DataDir, name)
		if err := os.RemoveAll(collectionPath); err != nil {
			fmt.Printf("Предупреждение: не удалось удалить файлы коллекции: %v\n", err)
		}
	}

	fmt.Printf("Коллекция %s успешно удалена\n", name)
	return nil
}

// insertDocumentCommand вставляет документ в коллекцию
func (cli *CLI) insertDocumentCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return fmt.Errorf("требуется указать имя коллекции и JSON документа")
	}

	collectionName := parts[0]
	jsonStr := parts[1]

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Разбор JSON
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		return fmt.Errorf("неверный формат JSON: %v", err)
	}

	// Проверить наличие _id
	id, ok := doc["_id"].(string)
	if !ok {
		return fmt.Errorf("документ должен содержать поле _id строкового типа")
	}

	// Удалить _id из content
	delete(doc, "_id")

	// Создать документ
	document := storage.Document{
		ID:      id,
		Content: doc,
	}

	// Вставить документ
	if err := collection.InsertDocument(document); err != nil {
		return err
	}

	fmt.Printf("Документ с ID %s успешно вставлен в коллекцию %s\n", id, collectionName)
	return nil
}

// getDocumentCommand получает документ по ID
func (cli *CLI) getDocumentCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return fmt.Errorf("требуется указать имя коллекции и ID документа")
	}

	collectionName := parts[0]
	id := parts[1]

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Получить документ
	doc, err := collection.GetDocument(id)
	if err != nil {
		return err
	}

	// Добавить _id к content для вывода
	result := make(map[string]interface{})
	for k, v := range doc.Content {
		result[k] = v
	}
	result["_id"] = doc.ID

	// Вывести документ
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// updateDocumentCommand обновляет документ
func (cli *CLI) updateDocumentCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 3)
	if len(parts) < 3 {
		return fmt.Errorf("требуется указать имя коллекции, ID документа и JSON обновления")
	}

	collectionName := parts[0]
	id := parts[1]
	jsonStr := parts[2]

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Получить существующий документ
	doc, err := collection.GetDocument(id)
	if err != nil {
		return err
	}

	// Разбор JSON обновления
	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &updateData); err != nil {
		return fmt.Errorf("неверный формат JSON: %v", err)
	}

	// Удалить _id из обновления, если есть
	delete(updateData, "_id")

	// Обновить содержимое документа
	for k, v := range updateData {
		doc.Content[k] = v
	}

	// Обновить документ
	if err := collection.UpdateDocument(doc); err != nil {
		return err
	}

	fmt.Printf("Документ с ID %s успешно обновлен\n", id)
	return nil
}

// deleteDocumentCommand удаляет документ
func (cli *CLI) deleteDocumentCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return fmt.Errorf("требуется указать имя коллекции и ID документа")
	}

	collectionName := parts[0]
	id := parts[1]

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Удалить документ
	if err := collection.DeleteDocument(id); err != nil {
		return err
	}

	fmt.Printf("Документ с ID %s успешно удален\n", id)
	return nil
}

// listDocumentsCommand выводит список документов в коллекции
func (cli *CLI) listDocumentsCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 1 || parts[0] == "" {
		return fmt.Errorf("требуется указать имя коллекции")
	}

	collectionName := parts[0]
	limit := -1

	if len(parts) > 1 && parts[1] != "" {
		fmt.Sscanf(parts[1], "%d", &limit)
	}

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Получить документы
	docs, err := collection.ListDocuments()
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		fmt.Printf("Коллекция %s не содержит документов\n", collectionName)
		return nil
	}

	fmt.Printf("Документы в коллекции %s:\n", collectionName)

	// Применить ограничение, если указано
	if limit > 0 && limit < len(docs) {
		docs = docs[:limit]
	}

	// Вывести документы
	for i, doc := range docs {
		// Добавить _id к content для вывода
		result := make(map[string]interface{})
		for k, v := range doc.Content {
			result[k] = v
		}
		result["_id"] = doc.ID

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%d. %s\n", i+1, string(jsonBytes))
	}

	return nil
}

// createIndexCommand создает индекс
func (cli *CLI) createIndexCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return fmt.Errorf("требуется указать имя коллекции и поле для индексации")
	}

	collectionName := parts[0]
	field := parts[1]

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Создать индекс
	if err := collection.CreateIndex(field, "btree", cli.Config.DefaultBTreeOrder); err != nil {
		return err
	}

	fmt.Printf("Индекс по полю %s успешно создан\n", field)
	return nil
}

// dropIndexCommand удаляет индекс
func (cli *CLI) dropIndexCommand(args string) error {
	// Разбор аргументов
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		return fmt.Errorf("требуется указать имя коллекции и поле индекса")
	}

	collectionName := parts[0]
	field := parts[1]

	// Получить коллекцию
	collection, err := cli.DB.GetCollection(collectionName)
	if err != nil {
		return err
	}

	// Удалить индекс
	if err := collection.DropIndex(field); err != nil {
		return err
	}

	fmt.Printf("Индекс по полю %s успешно удален\n", field)
	return nil
}

// queryCommand выполняет SQL-подобный запрос
func (cli *CLI) queryCommand(queryStr string) error {
	if queryStr == "" {
		return fmt.Errorf("требуется указать запрос")
	}

	// Выполнить запрос
	results, err := cli.DB.Query(queryStr)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("Запрос не вернул результатов")
		return nil
	}

	fmt.Printf("Результаты запроса (%d строк):\n", len(results))

	// Вывести результаты
	for i, result := range results {
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%d. %s\n", i+1, string(jsonBytes))
	}

	return nil
}
