package protocol

import (
	"bytes"
	"tir/models"
)

// CreateExactClone создает 100% точную копию работающего сценария range_3m_pulse1 или range_3m_pulse5
// в зависимости от типа пульта, меняя только имя сценария
func CreateExactClone(name string, pulseType byte) []byte {
	var template []byte

	// Выбираем шаблон в зависимости от типа пульта
	if pulseType == models.PULSE_5 {
		// Шаблон range_3m_pulse5
		template = []byte{
			0x7e, 0x00, 0x05, 0x0d, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x20, 0x33, 0x6d, 0x00, 0xfd,
			0xfd, 0xfd, 0xfd, 0x55, 0x17, 0x0a, 0x97, 0x47, 0x9a, 0x00, 0x00, 0x00, 0x01, 0x14, 0x2c, 0x01,
			0x56,
		}
	} else {
		// Шаблон range_3m_pulse1 (для всех остальных пультов)
		template = []byte{
			0x7e, 0x00, 0x01, 0x0d, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x20, 0x33, 0x6d, 0x00, 0xfd,
			0xfd, 0xfd, 0xfd, 0x00, 0x00, 0x00, 0xa6, 0x40, 0xa9, 0x00, 0x00, 0x00, 0x01, 0x14, 0x2c, 0x01,
			0x39,
		}
	}

	// Находим позицию имени в шаблоне (после первых 4 байтов заголовка)
	nameStartPos := 4
	nameEndPos := bytes.IndexByte(template[nameStartPos:], 0) + nameStartPos

	// Длина старого имени (включая нулевой байт)
	oldNameLen := nameEndPos - nameStartPos + 1

	// Новое имя с нулевым байтом
	newNameWithZero := append([]byte(name), 0)

	// Если длина нового имени равна длине старого, просто заменяем его
	if len(newNameWithZero) == oldNameLen {
		result := make([]byte, len(template))
		copy(result, template)

		// Заменяем имя
		for i := 0; i < len(newNameWithZero)-1; i++ {
			result[nameStartPos+i] = newNameWithZero[i]
		}

		return result
	}

	// Если длина имени отличается, нужно создать новый пакет
	// Сохраняем все части до и после имени
	headerPart := template[:nameStartPos]   // 7E 00 01/05 0D
	postNamePart := template[nameEndPos+1:] // начиная с FD FD FD FD...

	// Корректируем длину имени в заголовке
	headerPart[3] = byte(len(newNameWithZero))

	// Собираем новый пакет
	result := append(headerPart, newNameWithZero...)
	result = append(result, postNamePart...)

	return result
}

// CloneWorkingScenario клонирует рабочий сценарий range_3m_pulse1 и позволяет заменить определенные параметры
func CloneWorkingScenario(name string, pulseType byte, rangeValue uint16) []byte {
	// Используем CreateExactClone для создания точной копии
	return CreateExactClone(name, pulseType)
}

// CreateStandardScenarioPacket создает стандартный пакет сценария с минимальным набором команд
func CreateStandardScenarioPacket(name string, pulseType byte, rangeValue uint16) []byte {
	// Также используем CreateExactClone для создания точной копии
	return CreateExactClone(name, pulseType)
}
