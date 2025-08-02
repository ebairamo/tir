package models

// Константы для пультов
const (
	PULSE_1 = 0x01
	PULSE_2 = 0x02
	PULSE_3 = 0x03
	PULSE_4 = 0x04
	PULSE_5 = 0x05
	PULSE_6 = 0x06
)

// Константы для устройства
const (
	CMD_EDGE_POSITION      = 0x1102 // Повернуть мишень в ребро
	CMD_ENEMY_POSITION     = 0x1112 // Повернуть мишень в чужой
	CMD_LIGHT_ON           = 0x090A // Включить подсветку мишени
	CMD_LIGHT_OFF          = 0x0A0A // Выключить подсветку мишени
	CMD_OOP_SIMULATION_ON  = 0x0100 // Включить имитацию ООП
	CMD_OOP_SIMULATION_OFF = 0x0200 // Выключить имитацию ООП
	CMD_HIT_LIGHT_ON       = 0x0300 // Включить подсветку поражений
	CMD_HIT_LIGHT_OFF      = 0x0400 // Выключить подсветку поражений
	CMD_PARKING            = 0x0401 // Парковка
	CMD_MOVE_TO_RANGE      = 0x030A // Старт движения - к рубежу
	CMD_MOVE_TO_SHOOTER    = 0x1213 // Старт движения - к стрелку
	CMD_MANUAL_FEED_OFF    = 0x1619 // Ручная протяжка (откл)
	CMD_ENCODER            = 0xBE00 // Энкодер
	CMD_PAUSE              = 0x0500 // Пауза
	CMD_SET_RANGE          = 0x1300 // Установить рубеж
	CMD_SAFE_ZONE          = 0x1400 // Безопасная зона
	CMD_SET_SPEED          = 0x1500 // Установить скорость
)

// Карта команд для удобного поиска
var CommandMap = map[string]uint16{
	"Повернуть мишень в ребро":      CMD_EDGE_POSITION,
	"Повернуть мишень в чужой":      CMD_ENEMY_POSITION,
	"Включить подсветку мишени":     CMD_LIGHT_ON,
	"Выключить подсветку мишени":    CMD_LIGHT_OFF,
	"Включить имитацию ООП":         CMD_OOP_SIMULATION_ON,
	"Выключить имитацию ООП":        CMD_OOP_SIMULATION_OFF,
	"Включить подсветку поражений":  CMD_HIT_LIGHT_ON,
	"Выключить подсветку поражений": CMD_HIT_LIGHT_OFF,
	"Парковка": CMD_PARKING,
	"Старт движения - к рубежу":  CMD_MOVE_TO_RANGE,
	"Старт движения - к стрелку": CMD_MOVE_TO_SHOOTER,
	"Ручная протяжка (откл)":     CMD_MANUAL_FEED_OFF,
	"Энкодер":             CMD_ENCODER,
	"Пауза":               CMD_PAUSE,
	"Установить рубеж":    CMD_SET_RANGE,
	"Безопасная зона":     CMD_SAFE_ZONE,
	"Установить скорость": CMD_SET_SPEED,
}

// Обратная карта команд для поиска по коду
var ReverseCommandMap = map[uint16]string{
	CMD_EDGE_POSITION:      "Повернуть мишень в ребро",
	CMD_ENEMY_POSITION:     "Повернуть мишень в чужой",
	CMD_LIGHT_ON:           "Включить подсветку мишени",
	CMD_LIGHT_OFF:          "Выключить подсветку мишени",
	CMD_OOP_SIMULATION_ON:  "Включить имитацию ООП",
	CMD_OOP_SIMULATION_OFF: "Выключить имитацию ООП",
	CMD_HIT_LIGHT_ON:       "Включить подсветку поражений",
	CMD_HIT_LIGHT_OFF:      "Выключить подсветку поражений",
	CMD_PARKING:            "Парковка",
	CMD_MOVE_TO_RANGE:      "Старт движения - к рубежу",
	CMD_MOVE_TO_SHOOTER:    "Старт движения - к стрелку",
	CMD_MANUAL_FEED_OFF:    "Ручная протяжка (откл)",
	CMD_ENCODER:            "Энкодер",
	CMD_PAUSE:              "Пауза",
	CMD_SET_RANGE:          "Установить рубеж",
	CMD_SAFE_ZONE:          "Безопасная зона",
	CMD_SET_SPEED:          "Установить скорость",
}

// Определение команд с параметрами
var ParamCommands = map[uint16]string{
	CMD_PAUSE:     "длительность (сек)",
	CMD_SET_RANGE: "рубеж (см)",
	CMD_SAFE_ZONE: "безопасное расстояние (см)",
	CMD_SET_SPEED: "скорость",
}
