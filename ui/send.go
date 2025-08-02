package ui

import (
	"fmt"
	"time"
	"tir/comport"
	"tir/models"
)

// SendScenario подключается к COM-порту и отправляет выбранный сценарий
func SendScenario(scenarios map[string]models.Scenario) {
	// Выбор порта
	portName := "COM4" // По умолчанию COM4
	fmt.Print("Введите имя порта (по умолчанию COM4): ")
	var input string
	fmt.Scanln(&input)
	if input != "" {
		portName = input
	}

	// Выбор скорости
	baudRate := uint32(4800) // По умолчанию 4800 бод
	fmt.Print("Введите скорость порта (по умолчанию 4800): ")
	fmt.Scanln(&input)
	if input != "" {
		var rate int
		fmt.Sscanf(input, "%d", &rate)
		if rate > 0 {
			baudRate = uint32(rate)
		}
	}

	fmt.Printf("Попытка подключения к %s со скоростью %d бод...\n", portName, baudRate)

	// Открываем COM порт
	handle, err := comport.OpenPort(portName)
	if err != nil {
		fmt.Printf("Ошибка открытия порта: %v\n", err)
		return
	}
	defer comport.ClosePort(handle)

	// Установка параметров порта
	err = comport.SetCommParams(handle, baudRate)
	if err != nil {
		fmt.Printf("Ошибка установки параметров: %v\n", err)
		return
	}

	// Установка таймаутов
	err = comport.SetCommTimeouts(handle)
	if err != nil {
		fmt.Printf("Ошибка установки таймаутов: %v\n", err)
		return
	}

	fmt.Println("Порт успешно открыт")

	// Очищаем буферы
	comport.PurgeComm(handle)

	// Имитация цикла инициализации
	buffer := make([]byte, 64)
	fmt.Println("Выполнение последовательности инициализации...")
	for i := 0; i < 50; i++ {
		_, _ = comport.ReadPort(handle, buffer)
		time.Sleep(time.Millisecond * 16)
	}

	// Отправляем инициализационный пакет
	initPacket := []byte{0x7E, 0xAA}
	fmt.Println("Отправка инициализационного пакета...")
	_, err = comport.WritePort(handle, initPacket)
	if err != nil {
		fmt.Printf("Ошибка отправки инициализационного пакета: %v\n", err)
		return
	}

	// Пауза после инициализации
	time.Sleep(time.Millisecond * 500)
	comport.PurgeComm(handle)

	// Выбор сценария для отправки
	fmt.Println("Доступные сценарии:")
	var scenarioNames []string
	i := 1
	for name := range scenarios {
		fmt.Printf("%d. %s\n", i, name)
		scenarioNames = append(scenarioNames, name)
		i++
	}

	if len(scenarioNames) == 0 {
		fmt.Println("Нет доступных сценариев для отправки")
		return
	}

	fmt.Print("Выберите номер сценария для отправки: ")
	var scenarioChoice int
	fmt.Scanln(&scenarioChoice)

	if scenarioChoice < 1 || scenarioChoice > len(scenarioNames) {
		fmt.Println("Неверный выбор сценария")
		return
	}

	selectedScenario := scenarioNames[scenarioChoice-1]
	scenarioObj := scenarios[selectedScenario]

	// Проверяем, есть ли у сценария сырые данные
	var scenarioData []byte
	if len(scenarioObj.RawData) > 0 {
		scenarioData = scenarioObj.RawData
	} else {
		fmt.Println("Ошибка: у сценария отсутствуют данные для отправки")
		return
	}

	fmt.Printf("Отправка сценария '%s'...\n", selectedScenario)
	n, err := comport.WritePort(handle, scenarioData)
	if err != nil {
		fmt.Printf("Ошибка отправки сценария: %v\n", err)
	} else {
		fmt.Printf("Отправлено %d байт\n", n)
		fmt.Println("Сценарий успешно отправлен")

		// Печатаем отправленные данные для отладки
		fmt.Print("Отправленные данные: ")
		for _, b := range scenarioData {
			fmt.Printf("%02X ", b)
		}
		fmt.Println()
	}

	// Ожидаем ответа от устройства
	fmt.Println("Ожидание ответа...")
	for i := 0; i < 10; i++ {
		n, err := comport.ReadPort(handle, buffer)
		if err != nil {
			// Игнорируем ошибки чтения
		} else if n > 0 {
			fmt.Printf("Получен ответ (%d байт): % X\n", n, buffer[:n])
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Println("Закрытие порта...")
}

// DebugSendScenario отправляет сценарий с расширенной отладкой
func DebugSendScenario(scenarios map[string]models.Scenario) {
	fmt.Println("\nОтладочная отправка сценария")
	fmt.Println("==========================")

	// Выбор сценария для отправки
	fmt.Println("Выберите рабочий сценарий (который точно работает):")
	var workingScenarioNames []string

	// Предварительно определенные рабочие сценарии
	knownWorkingScenarios := []string{"test1", "range_3m_pulse1"}

	// Проверяем, существуют ли они
	for _, name := range knownWorkingScenarios {
		if _, exists := scenarios[name]; exists {
			workingScenarioNames = append(workingScenarioNames, name)
		}
	}

	if len(workingScenarioNames) == 0 {
		fmt.Println("Ошибка: не найдены известные рабочие сценарии")
		return
	}

	for i, name := range workingScenarioNames {
		fmt.Printf("%d. %s\n", i+1, name)
	}

	var choice int
	fmt.Print("Выберите номер рабочего сценария: ")
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(workingScenarioNames) {
		fmt.Println("Неверный выбор сценария")
		return
	}

	selectedScenario := workingScenarioNames[choice-1]

	// Выбор порта
	portName := "COM4" // По умолчанию COM4
	fmt.Print("Введите имя порта (по умолчанию COM4): ")
	var input string
	fmt.Scanln(&input)
	if input != "" {
		portName = input
	}

	// Выбор скорости
	baudRate := uint32(4800) // По умолчанию 4800 бод
	fmt.Print("Введите скорость порта (по умолчанию 4800): ")
	fmt.Scanln(&input)
	if input != "" {
		var rate int
		fmt.Sscanf(input, "%d", &rate)
		if rate > 0 {
			baudRate = uint32(rate)
		}
	}

	// Открываем COM порт
	fmt.Printf("Попытка подключения к %s со скоростью %d бод...\n", portName, baudRate)
	handle, err := comport.OpenPort(portName)
	if err != nil {
		fmt.Printf("Ошибка открытия порта: %v\n", err)
		return
	}
	defer comport.ClosePort(handle)

	// Установка параметров порта
	err = comport.SetCommParams(handle, baudRate)
	if err != nil {
		fmt.Printf("Ошибка установки параметров: %v\n", err)
		return
	}

	// Установка таймаутов
	err = comport.SetCommTimeouts(handle)
	if err != nil {
		fmt.Printf("Ошибка установки таймаутов: %v\n", err)
		return
	}

	fmt.Println("Порт успешно открыт")

	// Очищаем буферы
	comport.PurgeComm(handle)

	// Расширенная инициализация
	buffer := make([]byte, 64)
	fmt.Println("Выполнение расширенной последовательности инициализации...")

	// Пауза перед началом
	time.Sleep(time.Second * 1)

	// Читаем возможные данные
	fmt.Println("Чтение возможных данных из порта...")
	for i := 0; i < 10; i++ {
		n, _ := comport.ReadPort(handle, buffer)
		if n > 0 {
			fmt.Printf("Получены данные (%d байт): % X\n", n, buffer[:n])
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Отправляем несколько инициализационных пакетов с паузами
	initPackets := [][]byte{
		{0x7E, 0xAA},
		{0x7E, 0x5B},
		{0x7E, 0xAA},
	}

	for i, packet := range initPackets {
		fmt.Printf("Отправка инициализационного пакета %d: % X\n", i+1, packet)
		_, err = comport.WritePort(handle, packet)
		if err != nil {
			fmt.Printf("Ошибка отправки инициализационного пакета: %v\n", err)
			continue
		}

		// Ожидаем ответа
		fmt.Println("Ожидание ответа...")
		for j := 0; j < 10; j++ {
			n, _ := comport.ReadPort(handle, buffer)
			if n > 0 {
				fmt.Printf("Получен ответ (%d байт): % X\n", n, buffer[:n])
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		// Пауза между пакетами
		time.Sleep(time.Millisecond * 500)
	}

	// Получаем данные сценария
	scenarioData := scenarios[selectedScenario].RawData

	fmt.Printf("Отправка оригинального сценария '%s'...\n", selectedScenario)
	fmt.Printf("Данные сценария (%d байт): % X\n", len(scenarioData), scenarioData)

	// Отправляем сценарий
	n, err := comport.WritePort(handle, scenarioData)
	if err != nil {
		fmt.Printf("Ошибка отправки сценария: %v\n", err)
	} else {
		fmt.Printf("Отправлено %d байт\n", n)
		fmt.Println("Сценарий успешно отправлен")
	}

	// Ожидаем ответа от устройства
	fmt.Println("Ожидание ответа (увеличенное время)...")
	for i := 0; i < 30; i++ { // 3 секунды ожидания
		n, _ := comport.ReadPort(handle, buffer)
		if n > 0 {
			fmt.Printf("Получен ответ (%d байт): % X\n", n, buffer[:n])
		}
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Println("Закрытие порта...")
}
