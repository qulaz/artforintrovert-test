package logging

type StructureLogger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	// With Возвращает логгер с добавленными тегами
	With(keysAndValues ...interface{}) ContextLogger
}

type (
	loggerLevel uint8
	loggerMode  uint8
)

const (
	InfoLevel  loggerLevel = iota
	DebugLevel loggerLevel = iota
	WarnLevel  loggerLevel = iota
	ErrorLevel loggerLevel = iota
)

const (
	// ProductionMode логи пишутся в json-формате.
	ProductionMode loggerMode = iota
	// DevelopmentMode логи пишутся в строчном формате, оформленном для нормального восприятия человеком.
	DevelopmentMode loggerMode = iota
)

type LoggerConfig struct {
	// Mode режим логера. По-умолчанию инициализируется логгер в ProductionMode
	Mode loggerMode

	// Level уровень логирования. По-умолчанию установлен InfoLevel
	Level loggerLevel
}
