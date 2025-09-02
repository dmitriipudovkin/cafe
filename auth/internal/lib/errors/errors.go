package errfmt

import (
	"fmt"
)

// ErrorFormatter предоставляет метод для форматирования ошибок
type ErrorFormatter interface {
	Format(err error) error
}

// OpErrorFormatter форматирует ошибки с указанием операции
type OpErrorFormatter struct {
	Op string
}

// NewOpErrorFormatter создает новый экземпляр форматировщика ошибок с операцией
func NewOpErrorFormatter(op string) *OpErrorFormatter {
	return &OpErrorFormatter{Op: op}
}

// Format добавляет к ошибке контекст операции и дополнительные параметры
func (f *OpErrorFormatter) Format(err error, args ...any) error {
	if err == nil {
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("%s -> %w", f.Op, err)
	}

	// Если первый аргумент - строка, используем её как форматную строку
	if format, ok := args[0].(string); ok {
		return fmt.Errorf("%s -> %w", f.Op, fmt.Errorf(format, args[1:]))
	}

	// Иначе добавляем параметры как дополнительные сведения
	return fmt.Errorf("%s -> %v, %w", f.Op, args, err)
}
