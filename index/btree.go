package index

import (
	"errors"

	"github.com/urusofam/jsondb/storage"
)

// Index определяет интерфейс для индексирования
type Index interface {
	// Add добавляет документ в индекс
	Add(doc storage.Document) error
	
	// Remove удаляет документ из индекса
	Remove(id string) error
	
	// Search ищет документы по полю и значению
	Search(field string, value interface{}) ([]string, error)
}

// BTreeNode представляет узел в B-дереве
type BTreeNode struct {
	Keys     []interface{}
	Values   [][]string
	Children []*BTreeNode
	IsLeaf   bool
}

// BTreeIndex реализует интерфейс Index используя B-дерево
type BTreeIndex struct {
	Root     *BTreeNode
	Field    string
	Order    int
	DocIDs   map[string]interface{} // Сопоставляет ID документов со значениями полей
}

// NewBTreeIndex создает новый индекс B-дерева
func NewBTreeIndex(field string, order int) *BTreeIndex {
	return &BTreeIndex{
		Root: &BTreeNode{
			Keys:     make([]interface{}, 0),
			Values:   make([][]string, 0),
			Children: make([]*BTreeNode, 0),
			IsLeaf:   true,
		},
		Field:  field,
		Order:  order,
		DocIDs: make(map[string]interface{}),
	}
}

// Add добавляет документ в индекс
func (bt *BTreeIndex) Add(doc storage.Document) error {
	value, ok := getNestedValue(doc.Content, bt.Field)
	if !ok {
		return nil // Поле не существует, нечего индексировать
	}
	
	bt.DocIDs[doc.ID] = value
	
	// Найти, куда вставить значение в B-дереве
	ids, _ := bt.Search(bt.Field, value)
	
	// Проверяем, есть ли уже этот ID в списке
	found := false
	for _, id := range ids {
		if id == doc.ID {
			found = true
			break
		}
	}
	
	// Если ID ещё нет в списке, добавляем его
	if !found {
		ids = append(ids, doc.ID)
	}
	
	// Обновить или вставить значение
	return bt.insert(bt.Root, value, ids)
}

// getNestedValue получает вложенные значения из JSON документа
func getNestedValue(content map[string]interface{}, field string) (interface{}, bool) {
	// Для простоты предположим, что вложенность отсутствует
	val, ok := content[field]
	return val, ok
}

// insert добавляет значение в B-дерево
func (bt *BTreeIndex) insert(node *BTreeNode, key interface{}, docIDs []string) error {
	// Реализация логики вставки B-дерева
	// Это упрощенная версия
	
	if node.IsLeaf {
		// Вставить ключ и docIDs в правильную позицию
		pos := 0
		for pos < len(node.Keys) && compare(node.Keys[pos], key) < 0 {
			pos++
		}
		
		if pos < len(node.Keys) && compare(node.Keys[pos], key) == 0 {
			// Ключ уже существует, обновляем docIDs
			node.Values[pos] = docIDs
			return nil
		}
		
		// Вставить новый ключ и docIDs
		node.Keys = append(node.Keys, nil)
		node.Values = append(node.Values, nil)
		
		for i := len(node.Keys) - 1; i > pos; i-- {
			node.Keys[i] = node.Keys[i-1]
			node.Values[i] = node.Values[i-1]
		}
		
		node.Keys[pos] = key
		node.Values[pos] = docIDs
		
		// Разделить узел при необходимости
		return bt.splitIfNeeded(node)
	} else {
		// Найти дочерний узел для вставки
		pos := 0
		for pos < len(node.Keys) && compare(node.Keys[pos], key) < 0 {
			pos++
		}
		
		if pos < len(node.Keys) && compare(node.Keys[pos], key) == 0 {
			// Ключ уже существует, обновляем docIDs
			node.Values[pos] = docIDs
			return nil
		}
		
		// Вставка в дочерний узел
		childPos := pos
		if pos == len(node.Keys) {
			childPos = len(node.Children) - 1
		}
		return bt.insert(node.Children[childPos], key, docIDs)
	}
}

// compare сравнивает два значения
func compare(a, b interface{}) int {
	// Реализация логики сравнения для разных типов
	// Это упрощенная версия
	
	switch v1 := a.(type) {
	case string:
		if v2, ok := b.(string); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	case float64:
		if v2, ok := b.(float64); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
		// Преобразование int в float64 для сравнения
		if v2, ok := b.(int); ok {
			v2f := float64(v2)
			if v1 < v2f {
				return -1
			} else if v1 > v2f {
				return 1
			}
			return 0
		}
	case int:
		if v2, ok := b.(int); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
		// Преобразование float64 в int для сравнения
		if v2, ok := b.(float64); ok {
			v1f := float64(v1)
			if v1f < v2 {
				return -1
			} else if v1f > v2 {
				return 1
			}
			return 0
		}
	case bool:
		if v2, ok := b.(bool); ok {
			if !v1 && v2 {
				return -1
			} else if v1 && !v2 {
				return 1
			}
			return 0
		}
	}
	
	return 0
}

// splitIfNeeded разделяет узел, если он имеет слишком много ключей
func (bt *BTreeIndex) splitIfNeeded(node *BTreeNode) error {
	if len(node.Keys) <= bt.Order {
		return nil // Нет необходимости разделять
	}
	
	// Реализация разделения узла B-дерева
	// Для простоты сейчас оставим как есть
	// В реальной реализации здесь должно быть разделение по правилам B-дерева
	
	return nil
}

// Remove удаляет документ из индекса
func (bt *BTreeIndex) Remove(id string) error {
	value, ok := bt.DocIDs[id]
	if !ok {
		return nil // Документ не проиндексирован
	}
	
	// Найти IDs документов для этого значения
	ids, _ := bt.Search(bt.Field, value)
	
	// Удалить ID документа
	newIDs := make([]string, 0, len(ids)-1)
	for _, docID := range ids {
		if docID != id {
			newIDs = append(newIDs, docID)
		}
	}
	
	delete(bt.DocIDs, id)
	
	if len(newIDs) > 0 {
		// Обновить значение
		return bt.insert(bt.Root, value, newIDs)
	} else {
		// Удалить значение
		return bt.remove(bt.Root, value)
	}
}

// remove удаляет значение из B-дерева
func (bt *BTreeIndex) remove(node *BTreeNode, key interface{}) error {
	// Реализация логики удаления B-дерева
	// Для простоты сейчас реализуем базовую версию
	if !node.IsLeaf {
		// Найти позицию в текущем узле
		pos := 0
		for pos < len(node.Keys) && compare(node.Keys[pos], key) < 0 {
			pos++
		}
		
		if pos < len(node.Keys) && compare(node.Keys[pos], key) == 0 {
			// Ключ найден в этом узле
			// Удаляем ключ и его значения
			copy(node.Keys[pos:], node.Keys[pos+1:])
			copy(node.Values[pos:], node.Values[pos+1:])
			node.Keys = node.Keys[:len(node.Keys)-1]
			node.Values = node.Values[:len(node.Values)-1]
			return nil
		}
		
		// Ключ должен быть в дочернем узле
		childPos := pos
		if childPos >= len(node.Children) {
			childPos = len(node.Children) - 1
		}
		return bt.remove(node.Children[childPos], key)
	} else {
		// Поиск ключа в листовом узле
		pos := 0
		for pos < len(node.Keys) && compare(node.Keys[pos], key) < 0 {
			pos++
		}
		
		if pos < len(node.Keys) && compare(node.Keys[pos], key) == 0 {
			// Удаляем ключ и его значения
			copy(node.Keys[pos:], node.Keys[pos+1:])
			copy(node.Values[pos:], node.Values[pos+1:])
			node.Keys = node.Keys[:len(node.Keys)-1]
			node.Values = node.Values[:len(node.Values)-1]
			return nil
		}
	}
	return nil
}

// Search ищет документы по полю и значению
func (bt *BTreeIndex) Search(field string, value interface{}) ([]string, error) {
	if field != bt.Field {
		return nil, errors.New("несоответствие поля индекса")
	}
	
	return bt.search(bt.Root, value), nil
}

// search находит ID документов в B-дереве
func (bt *BTreeIndex) search(node *BTreeNode, key interface{}) []string {
	// Реализация логики поиска B-дерева
	// Это упрощенная версия
	
	pos := 0
	for pos < len(node.Keys) && compare(node.Keys[pos], key) < 0 {
		pos++
	}
	
	if pos < len(node.Keys) && compare(node.Keys[pos], key) == 0 {
		return node.Values[pos]
	}
	
	if node.IsLeaf {
		return []string{}
	}
	
	// Спуск в соответствующий дочерний узел
	childPos := pos
	if childPos >= len(node.Children) {
		childPos = len(node.Children) - 1
	}
	return bt.search(node.Children[childPos], key)
}
