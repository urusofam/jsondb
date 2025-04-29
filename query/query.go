package query

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yourusername/jsondb/storage"
)

// Query представляет запрос к базе данных
type Query struct {
	Select []string
	From   string
	Where  *Condition
	Limit  int
	Offset int
}

// Condition представляет условие запроса
type Condition struct {
	Left     interface{}
	Operator string
	Right    interface{}
	ChildOp  string
	Children []*Condition
}

// QueryParser разбирает строки запросов в объекты Query
type QueryParser struct {
	// Регулярные выражения для токенизации
	selectRegex *regexp.Regexp
	fromRegex   *regexp.Regexp
	whereRegex  *regexp.Regexp
	limitRegex  *regexp.Regexp
	offsetRegex *regexp.Regexp
}

// NewQueryParser создает новый парсер запросов
func NewQueryParser() *QueryParser {
	return &QueryParser{
		selectRegex: regexp.MustCompile(`(?i)SELECT\s+(.*?)\s+FROM`),
		fromRegex:   regexp.MustCompile(`(?i)FROM\s+(.*?)(\s+WHERE|\s+LIMIT|\s+OFFSET|$)`),
		whereRegex:  regexp.MustCompile(`(?i)WHERE\s+(.*?)(\s+LIMIT|\s+OFFSET|$)`),
		limitRegex:  regexp.MustCompile(`(?i)LIMIT\s+(\d+)`),
		offsetRegex: regexp.MustCompile(`(?i)OFFSET\s+(\d+)`),
	}
}

// Parse разбирает строку запроса в объект Query
func (qp *QueryParser) Parse(queryStr string) (*Query, error) {
	query := &Query{
		Limit:  -1,
		Offset: 0,
	}
	
	// Разбор SELECT
	selectMatches := qp.selectRegex.FindStringSubmatch(queryStr)
	if len(selectMatches) < 2 {
		return nil, errors.New("неверный оператор SELECT")
	}
	
	fields := strings.Split(selectMatches[1], ",")
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}
	query.Select = fields
	
	// Разбор FROM
	fromMatches := qp.fromRegex.FindStringSubmatch(queryStr)
	if len(fromMatches) < 2 {
		return nil, errors.New("неверный оператор FROM")
	}
	
	query.From = strings.TrimSpace(fromMatches[1])
	
	// Разбор WHERE
	whereMatches := qp.whereRegex.FindStringSubmatch(queryStr)
	if len(whereMatches) >= 2 {
		condition, err := qp.parseCondition(whereMatches[1])
		if err != nil {
			return nil, err
		}
		query.Where = condition
	}
	
	// Разбор LIMIT
	limitMatches := qp.limitRegex.FindStringSubmatch(queryStr)
	if len(limitMatches) >= 2 {
		limit, err := strconv.Atoi(limitMatches[1])
		if err != nil {
			return nil, err
		}
		query.Limit = limit
	}
	
	// Разбор OFFSET
	offsetMatches := qp.offsetRegex.FindStringSubmatch(queryStr)
	if len(offsetMatches) >= 2 {
		offset, err := strconv.Atoi(offsetMatches[1])
		if err != nil {
			return nil, err
		}
		query.Offset = offset
	}
	
	return query, nil
}

// parseCondition разбирает строку условия в объект Condition
func (qp *QueryParser) parseCondition(condStr string) (*Condition, error) {
	// Проверка на AND или OR
	if strings.Contains(strings.ToUpper(condStr), " AND ") {
		parts := strings.Split(condStr, " AND ")
		conditions := make([]*Condition, len(parts))
		
		for i, part := range parts {
			cond, err := qp.parseCondition(part)
			if err != nil {
				return nil, err
			}
			conditions[i] = cond
		}
		
		return &Condition{
			ChildOp:  "AND",
			Children: conditions,
		}, nil
	}
	
	if strings.Contains(strings.ToUpper(condStr), " OR ") {
		parts := strings.Split(condStr, " OR ")
		conditions := make([]*Condition, len(parts))
		
		for i, part := range parts {
			cond, err := qp.parseCondition(part)
			if err != nil {
				return nil, err
			}
			conditions[i] = cond
		}
		
		return &Condition{
			ChildOp:  "OR",
			Children: conditions,
		}, nil
	}
	
	// Парсинг простого условия
	for _, op := range []string{">=", "<=", "!=", ">", "<", "="} {
		if strings.Contains(condStr, op) {
			parts := strings.SplitN(condStr, op, 2)
			if len(parts) != 2 {
				return nil, errors.New("неверное условие")
			}
			
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			
			// Разбор правого значения
			var rightVal interface{}
			
			if strings.HasPrefix(right, "'") && strings.HasSuffix(right, "'") {
				// Строка
				rightVal = right[1 : len(right)-1]
			} else if strings.ToLower(right) == "true" || strings.ToLower(right) == "false" {
				// Булево значение
				rightVal = (strings.ToLower(right) == "true")
			} else if strings.Contains(right, ".") {
				// Число с плавающей точкой
				f, err := strconv.ParseFloat(right, 64)
				if err != nil {
					return nil, err
				}
				rightVal = f
			} else {
				// Целое число
				i, err := strconv.Atoi(right)
				if err != nil {
					return nil, err
				}
				rightVal = i
			}
			
			return &Condition{
				Left:     left,
				Operator: op,
				Right:    rightVal,
			}, nil
		}
	}
	
	return nil, errors.New("неверное условие")
}

// QueryExecutor выполняет запросы к базе данных
type QueryExecutor struct {
	DB map[string]Collection
}

// Collection представляет коллекцию документов
type Collection struct {
	Storage storage.Storage
}

// NewQueryExecutor создает новый исполнитель запросов
func NewQueryExecutor(collections map[string]Collection) *QueryExecutor {
	return &QueryExecutor{
		DB: collections,
	}
}

// Execute выполняет запрос к базе данных
func (qe *QueryExecutor) Execute(query *Query) ([]map[string]interface{}, error) {
	collection, ok := qe.DB[query.From]
	if !ok {
		return nil, fmt.Errorf("коллекция %s не найдена", query.From)
	}
	
	// Получить все документы
	docs, err := collection.Storage.List()
	if err != nil {
		return nil, err
	}
	
	results := make([]map[string]interface{}, 0)
	
	// Применить условия
	for _, doc := range docs {
		if query.Where == nil || qe.evalCondition(doc, query.Where) {
			result := make(map[string]interface{})
			
			if len(query.Select) == 1 && query.Select[0] == "*" {
				// Выбрать все поля
				for k, v := range doc.Content {
					result[k] = v
				}
				result["_id"] = doc.ID
			} else {
				// Выбрать определенные поля
				for _, field := range query.Select {
					if field == "_id" {
						result["_id"] = doc.ID
					} else {
						value, ok := getNestedValue(doc.Content, field)
						if ok {
							result[field] = value
						}
					}
				}
			}
			
			results = append(results, result)
		}
	}
	
	// Применить OFFSET и LIMIT
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	}
	
	if query.Limit >= 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}
	
	return results, nil
}

// getNestedValue получает вложенные значения из JSON документа
func getNestedValue(content map[string]interface{}, field string) (interface{}, bool) {
	// Для простоты предположим, что вложенность отсутствует
	val, ok := content[field]
	return val, ok
}

// evalCondition оценивает условие для документа
func (qe *QueryExecutor) evalCondition(doc storage.Document, cond *Condition) bool {
	if len(cond.Children) > 0 {
		// Оценить дочерние условия
		results := make([]bool, len(cond.Children))
		
		for i, child := range cond.Children {
			results[i] = qe.evalCondition(doc, child)
		}
		
		if cond.ChildOp == "AND" {
			for _, result := range results {
				if !result {
					return false
				}
			}
			return true
		} else if cond.ChildOp == "OR" {
			for _, result := range results {
				if result {
					return true
				}
			}
			return false
		}
	}
	
	// Оценить простое условие
	fieldName, ok := cond.Left.(string)
	if !ok {
		return false
	}
	
	var value interface{}
	if fieldName == "_id" {
		value = doc.ID
	} else {
		value, ok = getNestedValue(doc.Content, fieldName)
		if !ok {
			return false
		}
	}
	
	return qe.compareValues(value, cond.Operator, cond.Right)
}

// compareValues сравнивает два значения с использованием указанного оператора
func (qe *QueryExecutor) compareValues(left interface{}, op string, right interface{}) bool {
	switch op {
	case "=":
		return qe.equals(left, right)
	case "!=":
		return !qe.equals(left, right)
	case ">":
		return qe.compare(left, right) > 0
	case ">=":
		return qe.compare(left, right) >= 0
	case "<":
		return qe.compare(left, right) < 0
	case "<=":
		return qe.compare(left, right) <= 0
	default:
		return false
	}
}

// equals проверяет, равны ли два значения
func (qe *QueryExecutor) equals(a, b interface{}) bool {
	return qe.compare(a, b) == 0
}

// compare сравнивает два значения
func (qe *QueryExecutor) compare(a, b interface{}) int {
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
		var v2 float64
		switch b := b.(type) {
		case float64:
			v2 = b
		case int:
			v2 = float64(b)
		default:
			return 0
		}
		
		if v1 < v2 {
			return -1
		} else if v1 > v2 {
			return 1
		}
		return 0
	case int:
		var v2 int
		switch b := b.(type) {
		case int:
			v2 = b
		case float64:
			v2 = int(b)
		default:
			return 0
		}
		
		if v1 < v2 {
			return -1
		} else if v1 > v2 {
			return 1
		}
		return 0
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
