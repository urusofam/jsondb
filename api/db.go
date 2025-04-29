package api

import (
	"errors"
	"fmt"
	"sync"

	"github.com/urusofam/jsondb/index"
	"github.com/urusofam/jsondb/query"
	"github.com/urusofam/jsondb/storage"
)

// DB является основным API базы данных
type DB struct {
	Collections map[string]*Collection
	Parser      *query.QueryParser
	Executor    *query.QueryExecutor
	Functions   *query.FunctionRegistry
	Mutex       sync.RWMutex
}

// NewDB создает новую базу данных
func NewDB() *DB {
	collections := make(map[string]*Collection)
	qCollections := make(map[string]query.Collection)
	
	db := &DB{
		Collections: collections,
		Parser:      query.NewQueryParser(),
		Executor:    query.NewQueryExecutor(qCollections),
		Functions:   query.NewFunctionRegistry(),
	}
	
	return db
}

// CreateCollection создает новую коллекцию
func (db *DB) CreateCollection(name string, storage storage.Storage) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	if _, ok := db.Collections[name]; ok {
		return fmt.Errorf("коллекция %s уже существует", name)
	}
	
	collection := &Collection{
		Name:    name,
		Storage: storage,
		Indexes: make(map[string]index.Index),
	}
	
	db.Collections[name] = collection
	
	// Обновить коллекции для исполнителя запросов
	qCollections := make(map[string]query.Collection)
	for name, coll := range db.Collections {
		qCollections[name] = query.Collection{
			Storage: coll.Storage,
		}
	}
	db.Executor = query.NewQueryExecutor(qCollections)
	
	return nil
}

// GetCollection возвращает коллекцию
func (db *DB) GetCollection(name string) (*Collection, error) {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	
	collection, ok := db.Collections[name]
	if !ok {
		return nil, fmt.Errorf("коллекция %s не найдена", name)
	}
	
	return collection, nil
}

// DropCollection удаляет коллекцию
func (db *DB) DropCollection(name string) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	if _, ok := db.Collections[name]; !ok {
		return fmt.Errorf("коллекция %s не найдена", name)
	}
	
	delete(db.Collections, name)
	
	// Обновить коллекции для исполнителя запросов
	qCollections := make(map[string]query.Collection)
	for name, coll := range db.Collections {
		qCollections[name] = query.Collection{
			Storage: coll.Storage,
		}
	}
	db.Executor = query.NewQueryExecutor(qCollections)
	
	return nil
}

// Query выполняет запрос
func (db *DB) Query(queryStr string) ([]map[string]interface{}, error) {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	
	query, err := db.Parser.Parse(queryStr)
	if err != nil {
		return nil, err
	}
	
	return db.Executor.Execute(query)
}

// Collection предоставляет операции над коллекцией
type Collection struct {
	Name    string
	Storage storage.Storage
	Indexes map[string]index.Index
	Mutex   sync.RWMutex
}

// InsertDocument вставляет документ в коллекцию
func (c *Collection) InsertDocument(doc storage.Document) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	if doc.ID == "" {
		return errors.New("ID документа обязателен")
	}
	
	if err := c.Storage.Save(doc); err != nil {
		return err
	}
	
	// Обновить индексы
	for _, idx := range c.Indexes {
		if err := idx.Add(doc); err != nil {
			return err
		}
	}
	
	return nil
}

// GetDocument получает документ из коллекции
func (c *Collection) GetDocument(id string) (storage.Document, error) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	
	return c.Storage.Get(id)
}

// UpdateDocument обновляет документ в коллекции
func (c *Collection) UpdateDocument(doc storage.Document) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	if doc.ID == "" {
		return errors.New("ID документа обязателен")
	}
	
	// Удалить документ из индексов
	oldDoc, err := c.Storage.Get(doc.ID)
	if err == nil {
		for _, idx := range c.Indexes {
			if err := idx.Remove(oldDoc.ID); err != nil {
				return err
			}
		}
	}
	
	if err := c.Storage.Save(doc); err != nil {
		return err
	}
	
	// Обновить индексы
	for _, idx := range c.Indexes {
		if err := idx.Add(doc); err != nil {
			return err
		}
	}
	
	return nil
}

// DeleteDocument удаляет документ из коллекции
func (c *Collection) DeleteDocument(id string) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	// Удалить документ из индексов
	for _, idx := range c.Indexes {
		if err := idx.Remove(id); err != nil {
			return err
		}
	}
	
	return c.Storage.Delete(id)
}

// CreateIndex создает индекс по полю
func (c *Collection) CreateIndex(field string, indexType string, order int) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	if _, ok := c.Indexes[field]; ok {
		return fmt.Errorf("индекс по полю %s уже существует", field)
	}
	
	var idx index.Index
	
	switch indexType {
	case "btree":
		idx = index.NewBTreeIndex(field, order)
	default:
		return fmt.Errorf("неизвестный тип индекса: %s", indexType)
	}
	
	c.Indexes[field] = idx
	
	// Добавить все документы в индекс
	docs, err := c.Storage.List()
	if err != nil {
		return err
	}
	
	for _, doc := range docs {
		if err := idx.Add(doc); err != nil {
			return err
		}
	}
	
	return nil
}

// DropIndex удаляет индекс
func (c *Collection) DropIndex(field string) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	if _, ok := c.Indexes[field]; !ok {
		return fmt.Errorf("индекс по полю %s не найден", field)
	}
	
	delete(c.Indexes, field)
	
	return nil
}

// FindByIndex находит документы используя индекс
func (c *Collection) FindByIndex(field string, value interface{}) ([]storage.Document, error) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	
	idx, ok := c.Indexes[field]
	if !ok {
		return nil, fmt.Errorf("индекс по полю %s не найден", field)
	}
	
	ids, err := idx.Search(field, value)
	if err != nil {
		return nil, err
	}
	
	docs := make([]storage.Document, 0, len(ids))
	
	for _, id := range ids {
		doc, err := c.Storage.Get(id)
		if err != nil {
			continue
		}
		
		docs = append(docs, doc)
	}
	
	return docs, nil
}

// ListDocuments возвращает все документы в коллекции
func (c *Collection) ListDocuments() ([]storage.Document, error) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	
	return c.Storage.List()
}

// Size возвращает количество документов в коллекции
func (c *Collection) Size() (int, error) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	
	docs, err := c.Storage.List()
	if err != nil {
		return 0, err
	}
	
	return len(docs), nil
}
