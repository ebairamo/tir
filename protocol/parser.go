package protocol

import (
	"fmt"
	"tir/models"
)

// Разобрать сырые данные сценария в структурированный вид
func ParseScenarioData(data []byte) (models.Scenario, error) {
	scenario := models.Scenario{}

	// Минимальная проверка
	if len(data) < 16 || data[0] != 0x7E || data[1] != 0x00 {
		return scenario, fmt.Errorf("неверный формат данных сценария")
	}

	// Получаем версию пульта
	scenario.PulseType = data[2]

	// Извлекаем имя сценария
	nameLength := int(data[3])
	if nameLength+4 > len(data) {
		return scenario, fmt.Errorf("неверная длина имени сценария")
	}

	nameBytes := data[4 : 4+nameLength-1] // Без нулевого байта
	scenario.Name = string(nameBytes)

	// Ищем начало команд
	cmdStart := 4 + nameLength + 6 // Заголовок + имя + константные данные
	if cmdStart >= len(data) {
		return scenario, fmt.Errorf("в данных нет команд")
	}

	// Начинаем чтение команд
	for i := cmdStart; i < len(data)-1; { // -1 чтобы не включать контрольную сумму
		if i+1 >= len(data) {
			break
		}

		// Получаем 2 байта команды
		cmdCode := uint16(data[i]) | uint16(data[i+1])<<8
		name, exists := models.ReverseCommandMap[cmdCode]

		if !exists {
			// Пробуем другой порядок байтов
			cmdCode = uint16(data[i]<<8) | uint16(data[i+1])
			name, exists = models.ReverseCommandMap[cmdCode]

			if !exists {
				// Если команда не найдена, пропускаем байт и продолжаем
				i++
				continue
			}
		}

		// Проверяем, имеет ли команда параметры
		_, hasParam := models.ParamCommands[cmdCode]

		command := models.Command{
			Name:     name,
			Code:     cmdCode,
			HasParam: hasParam,
		}

		// Если команда с параметром, читаем параметр
		if hasParam && i+3 < len(data) {
			paramValue := uint16(data[i+2]) | uint16(data[i+3])<<8
			command.ParamValue = paramValue
			command.ParamName = models.ParamCommands[cmdCode]
			i += 4 // Пропускаем 2 байта команды и 2 байта параметра
		} else {
			i += 2 // Пропускаем только 2 байта команды
		}

		scenario.Commands = append(scenario.Commands, command)
	}

	return scenario, nil
}

// Генерировать бинарный пакет из структуры сценария
func GenerateScenarioPacket(scenario models.Scenario) []byte {
	// Имя с нулевым байтом
	nameWithZero := append([]byte(scenario.Name), 0)

	// Заголовок: 7E 00 [версия пульта] [длина имени с нулевым байтом]
	header := []byte{0x7E, 0x00, scenario.PulseType, byte(len(nameWithZero))}

	// Начинаем собирать пакет
	packet := append(header, nameWithZero...)

	// Константные данные
	constantData := []byte{0xFD, 0xFD, 0xFD, 0xFD, 0x00, 0x00}
	packet = append(packet, constantData...)

	// Добавляем переменную часть в зависимости от пульта
	variablePart := getVariablePartForPulse(scenario.PulseType)
	packet = append(packet, variablePart...)

	// Команды сценария в бинарном формате
	commandsData := []byte{}

	for _, cmd := range scenario.Commands {
		// Преобразуем uint16 в два байта
		highByte := byte((cmd.Code >> 8) & 0xFF)
		lowByte := byte(cmd.Code & 0xFF)

		commandsData = append(commandsData, lowByte, highByte)

		// Если команда с параметром, добавляем его
		if cmd.HasParam {
			paramLowByte := byte(cmd.ParamValue & 0xFF)
			paramHighByte := byte((cmd.ParamValue >> 8) & 0xFF)
			commandsData = append(commandsData, paramLowByte, paramHighByte)
		}
	}

	// Добавляем команды к пакету
	packet = append(packet, commandsData...)

	// Вычисляем контрольную сумму (XOR всех байтов команд)
	var checksum byte
	for _, b := range commandsData {
		checksum ^= b
	}

	// Добавляем контрольную сумму
	packet = append(packet, checksum)

	return packet
}

// Получить переменную часть заголовка для разных типов пультов
func getVariablePartForPulse(pulseType byte) []byte {
	switch pulseType {
	case models.PULSE_1:
		return []byte{0x00, 0xA6, 0x40, 0xA9, 0x00, 0x00, 0x00}
	case models.PULSE_2:
		return []byte{0x20, 0x00, 0x80, 0x59, 0xD4, 0xB4, 0x00, 0x00, 0x00}
	case models.PULSE_3:
		return []byte{0x00, 0x00, 0x00, 0x5D, 0xC6, 0x83, 0x00, 0x00, 0x00}
	case models.PULSE_4:
		return []byte{0x00, 0x00, 0x00, 0x6B, 0x47, 0x59, 0x00, 0x00, 0x00}
	case models.PULSE_5:
		return []byte{0x55, 0x17, 0x0A, 0x97, 0x47, 0x9A, 0x00, 0x00, 0x00}
	case models.PULSE_6:
		return []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // Предполагаемые данные для пульта 6
	default:
		return []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}
}

// Импортировать сохраненные ранее сценарии в новый формат
func ImportDefaultScenarios(scenarios map[string]models.Scenario) {
	savedScenarios := map[string][]byte{
		"test1": {
			0x7e, 0x00, 0x01, 0x13, 0x74, 0x65, 0x73, 0x74, 0x31, 0x00, 0xfd, 0xfd, 0xfd, 0xfd, 0x00, 0x00,
			0xff, 0xff, 0x00, 0x00, 0xd8, 0x72, 0x85, 0x00, 0x00, 0x00, 0x01, 0x13, 0xe8, 0x03, 0x15, 0x32,
			0x00, 0x14, 0x2c, 0x01, 0x11, 0x02, 0x03, 0x0a, 0x05, 0x00, 0x04, 0x01, 0x09, 0xca,
		},
		"test1_bez_park": {
			0x7e, 0x00, 0x01, 0x1b, 0x74, 0x65, 0x73, 0x74, 0x31, 0x20, 0x62, 0x65, 0x7a, 0x20, 0x70, 0x61,
			0x72, 0x6b, 0x00, 0xfd, 0xfd, 0xfd, 0xfd, 0x00, 0x00, 0x00, 0x01, 0x13, 0xe8, 0x03, 0x15, 0x32,
			0x00, 0x14, 0x2c, 0x01, 0x11, 0x02, 0x03, 0x0a, 0x05, 0x00, 0x04, 0x01, 0x15, 0x32, 0x00, 0x14,
			0xc8, 0x00, 0x1c, 0x12, 0x02, 0x55,
		},
		"test5_30m_park": {
			0x7e, 0x00, 0x05, 0x0d, 0x74, 0x65, 0x73, 0x74, 0x35, 0x20, 0x33, 0x30, 0x6d, 0x20, 0x70, 0x61,
			0x72, 0x6b, 0x00, 0xfd, 0xfd, 0xfd, 0xfd, 0x00, 0x00, 0x00, 0x01, 0x13, 0xb8, 0x0b, 0x15, 0x32,
			0x00, 0x14, 0x2c, 0x01, 0x11, 0x02, 0x03, 0x65,
		},
		"range_3m_pulse1": {
			0x7e, 0x00, 0x01, 0x0d, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x20, 0x33, 0x6d, 0x00, 0xfd,
			0xfd, 0xfd, 0xfd, 0x00, 0x00, 0x00, 0xa6, 0x40, 0xa9, 0x00, 0x00, 0x00, 0x01, 0x13, 0x2c, 0x01,
			0x15, 0x32, 0x00, 0x14, 0x2c, 0x01, 0x11, 0x02, 0x03, 0xc1,
		},
		"range_3m_pulse5": {
			0x7e, 0x00, 0x05, 0x0d, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x20, 0x33, 0x6d, 0x00, 0xfd,
			0xfd, 0xfd, 0xfd, 0x55, 0x17, 0x0a, 0x97, 0x47, 0x9a, 0x00, 0x00, 0x00, 0x01, 0x13, 0x2c, 0x01,
			0x15, 0x32, 0x00, 0x14, 0x2c, 0x01, 0x11, 0x02, 0x03, 0x56,
		},
	}

	for name, data := range savedScenarios {
		// Импортируем как сырые данные, без разбора
		scenario := models.Scenario{
			Name:    name,
			RawData: data,
		}

		// Извлекаем версию пульта из пакета
		if len(data) > 3 {
			scenario.PulseType = data[2]
		}

		// Пытаемся разобрать команды
		parsedScenario, err := ParseScenarioData(data)
		if err == nil && len(parsedScenario.Commands) > 0 {
			scenario.Commands = parsedScenario.Commands
		}

		// Сохраняем сценарий
		scenarios[name] = scenario
	}

	fmt.Printf("Импортировано %d встроенных сценариев\n", len(savedScenarios))
}
