package config

// StorageType определяет тип хранения
type StorageType string

const (
	// StorageTypeMemory для хранения в памяти
	StorageTypeMemory StorageType = "memory"
	// StorageTypeFile для файлового хранения
	StorageTypeFile StorageType = "file"
)

// DBConfig содержит конфигурацию базы данных
type DBConfig struct {
	// Тип хранения (память или файл)
	StorageType StorageType
	
	// DataDir используется для хранения файлов (только для StorageTypeFile)
	DataDir string
	
	// UseCache указывает, должен ли файловый механизм хранения использовать кэш
	UseCache bool
	
	// DefaultBTreeOrder определяет порядок B-дерева по умолчанию для индексов
	DefaultBTreeOrder int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *DBConfig {
	return &DBConfig{
		StorageType:      StorageTypeMemory,
		DataDir:          "./data",
		UseCache:         true,
		DefaultBTreeOrder: 5,
	}
}

// NewFileStorageConfig создает конфигурацию для файлового хранения
func NewFileStorageConfig(dataDir string, useCache bool) *DBConfig {
	return &DBConfig{
		StorageType:      StorageTypeFile,
		DataDir:          dataDir,
		UseCache:         useCache,
		DefaultBTreeOrder: 5,
	}
}

// NewMemoryStorageConfig создает конфигурацию для хранения в памяти
func NewMemoryStorageConfig() *DBConfig {
	return &DBConfig{
		StorageType:      StorageTypeMemory,
		DefaultBTreeOrder: 5,
	}
}
