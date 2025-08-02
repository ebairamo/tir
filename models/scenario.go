package models

// Структура сценария
type Scenario struct {
	Name      string
	PulseType byte
	Commands  []Command
	RawData   []byte // Опционально, для хранения сырых данных при импорте
}

// Структура команды
type Command struct {
	Name       string
	Code       uint16
	HasParam   bool
	ParamName  string
	ParamValue uint16
}
