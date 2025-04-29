package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Document представляет JSON-документ
type Document struct {
	ID      string                 `json:"_id"`
	Content map[string]interface{} `json:"content"`
}

// Storage определяет интерфейс для механизмов хранения
type Storage interface {
	// Save сохраняет документ
	Save(doc Document) error
	
	// Get извлекает документ по ID
	Get(id string) (Document, error)
	
	// Delete удаляет документ по ID
	Delete(id string) error
	
	// List возвращает все документы
	List() ([]Document, error)
}

// FileStorage реализует Storage используя файловую систему
type FileStorage struct {
	Dir      string
	Mutex    sync.RWMutex
	UseCache bool
	Cache    map[string]Document
}

// NewFileStorage создает новое файловое хранилище
func NewFileStorage(dir string, useCache bool) (*FileStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	fs := &FileStorage{
		Dir:      dir,
		UseCache: useCache,
		Cache:    make(map[string]Document),
	}
	
	return fs, nil
}

// Save сохраняет документ
func (fs *FileStorage) Save(doc Document) error {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	
	filePath := filepath.Join(fs.Dir, doc.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	
	if fs.UseCache {
		fs.Cache[doc.ID] = doc
	}
	
	return nil
}

// Get извлекает документ по ID
func (fs *FileStorage) Get(id string) (Document, error) {
	fs.Mutex.RLock()
	defer fs.Mutex.RUnlock()
	
	if fs.UseCache {
		if doc, ok := fs.Cache[id]; ok {
			return doc, nil
		}
	}
	
	filePath := filepath.Join(fs.Dir, id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, err
	}
	
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return Document{}, err
	}
	
	if fs.UseCache {
		fs.Cache[id] = doc
	}
	
	return doc, nil
}

// Delete удаляет документ по ID
func (fs *FileStorage) Delete(id string) error {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	
	filePath := filepath.Join(fs.Dir, id+".json")
	if err := os.Remove(filePath); err != nil {
		return err
	}
	
	if fs.UseCache {
		delete(fs.Cache, id)
	}
	
	return nil
}

// List возвращает все документы
func (fs *FileStorage) List() ([]Document, error) {
	fs.Mutex.RLock()
	defer fs.Mutex.RUnlock()
	
	files, err := os.ReadDir(fs.Dir)
	if err != nil {
		return nil, err
	}
	
	docs := make([]Document, 0, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		id := filepath.Base(file.Name())
		id = id[:len(id)-5] // Удаляем расширение .json
		
		doc, err := fs.Get(id)
		if err != nil {
			continue
		}
		
		docs = append(docs, doc)
	}
	
	return docs, nil
}

// MemoryStorage реализует Storage используя память
type MemoryStorage struct {
	Docs  map[string]Document
	Mutex sync.RWMutex
}

// NewMemoryStorage создает новое хранилище в памяти
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Docs: make(map[string]Document),
	}
}

// Save сохраняет документ
func (ms *MemoryStorage) Save(doc Document) error {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	
	ms.Docs[doc.ID] = doc
	return nil
}

// Get извлекает документ по ID
func (ms *MemoryStorage) Get(id string) (Document, error) {
	ms.Mutex.RLock()
	defer ms.Mutex.RUnlock()
	
	doc, ok := ms.Docs[id]
	if !ok {
		return Document{}, errors.New("документ не найден")
	}
	
	return doc, nil
}

// Delete удаляет документ по ID
func (ms *MemoryStorage) Delete(id string) error {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	
	delete(ms.Docs, id)
	return nil
}

// List возвращает все документы
func (ms *MemoryStorage) List() ([]Document, error) {
	ms.Mutex.RLock()
	defer ms.Mutex.RUnlock()
	
	docs := make([]Document, 0, len(ms.Docs))
	for _, doc := range ms.Docs {
		docs = append(docs, doc)
	}
	
	return docs, nil
}
