package auto

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tir/comport"
	"tir/models"
)

// Константы для префиксов автоматизированных сценариев
const (
	AutoPrefix = "AUTO_P" // Префикс автоматических сценариев
)

// FindScenarioByDistanceAndPulse находит сценарий по дистанции и типу пульта
func FindScenarioByDistanceAndPulse(scenarios map[string]models.Scenario, distance int, pulseType byte) (string, bool) {
	// Сначала ищем по специальному формату имени для автоматизации
	autoName := fmt.Sprintf("%s%d_%dM", AutoPrefix, pulseType, distance)
	if scenario, exists := scenarios[autoName]; exists && scenario.PulseType == pulseType {
		return autoName, true
	}

	// Ищем по стандартным форматам имени
	formats := []string{
		fmt.Sprintf("Сценарий %dм пульт %d", distance, pulseType),
		fmt.Sprintf("Сценарий %dм", distance),
	}

	for _, format := range formats {
		for name, scenario := range scenarios {
			if strings.HasPrefix(name, format) && scenario.PulseType == pulseType {
				return name, true
			}
		}
	}

	// Если точное совпадение не найдено, ищем по начальной части имени и типу пульта
	searchName := fmt.Sprintf("%d", distance)
	for name, scenario := range scenarios {
		if strings.Contains(name, searchName) && scenario.PulseType == pulseType {
			return name, true
		}
	}

	return "", false
}

// SendScenarioAuto отправляет сценарий в автоматическом режиме
func SendScenarioAuto(scenarios map[string]models.Scenario, portName string, baudRate uint32, pulseType byte, distance int) error {
	fmt.Printf("Поиск сценария для дистанции %d м и пульта типа %d...\n", distance, pulseType)

	// Находим подходящий сценарий
	scenarioName, found := FindScenarioByDistanceAndPulse(scenarios, distance, pulseType)
	if !found {
		return fmt.Errorf("сценарий для дистанции %d м и пульта типа %d не найден", distance, pulseType)
	}

	fmt.Printf("Найден сценарий: %s\n", scenarioName)

	// Открываем COM порт
	fmt.Printf("Подключение к %s со скоростью %d бод...\n", portName, baudRate)
	handle, err := comport.OpenPort(portName)
	if err != nil {
		return fmt.Errorf("ошибка открытия порта: %v", err)
	}
	defer comport.ClosePort(handle)

	// Установка параметров порта
	err = comport.SetCommParams(handle, baudRate)
	if err != nil {
		return fmt.Errorf("ошибка установки параметров: %v", err)
	}

	// Установка таймаутов
	err = comport.SetCommTimeouts(handle)
	if err != nil {
		return fmt.Errorf("ошибка установки таймаутов: %v", err)
	}

	fmt.Println("Порт успешно открыт")

	// Очищаем буферы
	comport.PurgeComm(handle)

	// Имитация цикла инициализации
	buffer := make([]byte, 64)
	fmt.Println("Выполнение последовательности инициализации...")
	for i := 0; i < 10; i++ {
		_, _ = comport.ReadPort(handle, buffer)
		time.Sleep(time.Millisecond * 16)
	}

	// Отправляем инициализационный пакет
	initPacket := []byte{0x7E, 0xAA}
	fmt.Println("Отправка инициализационного пакета...")
	_, err = comport.WritePort(handle, initPacket)
	if err != nil {
		return fmt.Errorf("ошибка отправки инициализационного пакета: %v", err)
	}

	// Пауза после инициализации
	time.Sleep(time.Millisecond * 500)
	comport.PurgeComm(handle)

	// Получаем данные сценария
	scenarioObj := scenarios[scenarioName]
	scenarioData := scenarioObj.RawData

	// Проверяем, есть ли у сценария сырые данные
	if len(scenarioData) == 0 {
		return fmt.Errorf("ошибка: у сценария отсутствуют данные для отправки")
	}

	fmt.Printf("Отправка сценария '%s'...\n", scenarioName)
	n, err := comport.WritePort(handle, scenarioData)
	if err != nil {
		return fmt.Errorf("ошибка отправки сценария: %v", err)
	}

	fmt.Printf("Отправлено %d байт\n", n)
	fmt.Println("Сценарий успешно отправлен")

	// Печатаем отправленные данные для отладки
	fmt.Print("Отправленные данные: ")
	for _, b := range scenarioData {
		fmt.Printf("%02X ", b)
	}
	fmt.Println()

	// Ожидаем ответа от устройства
	fmt.Println("Ожидание ответа...")
	for i := 0; i < 10; i++ {
		n, _ := comport.ReadPort(handle, buffer)
		if n > 0 {
			fmt.Printf("Получен ответ (%d байт): % X\n", n, buffer[:n])
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Println("Закрытие порта...")
	return nil
}

// PrepareForAutomation подготавливает сценарии для автоматизации
func PrepareForAutomation(scenarios map[string]models.Scenario) map[string]models.Scenario {
	// Создаем специальные AUTO-сценарии из существующих
	for name, scenario := range scenarios {
		// Извлекаем информацию о дистанции из имени
		if strings.Contains(name, "м") && scenario.PulseType > 0 {
			distStr := ""
			for _, char := range name {
				if char >= '0' && char <= '9' {
					distStr += string(char)
				} else if len(distStr) > 0 && (char == 'м' || char == 'M') {
					break
				}
			}

			// Если удалось извлечь дистанцию, создаем AUTO-сценарий
			if distance, err := strconv.Atoi(distStr); err == nil && distance > 0 {
				autoName := fmt.Sprintf("%s%d_%dM", AutoPrefix, scenario.PulseType, distance)

				// Проверяем, существует ли уже такой сценарий
				if _, exists := scenarios[autoName]; !exists {
					// Создаем копию сценария с AUTO-именем
					autoScenario := models.Scenario{
						Name:      autoName,
						PulseType: scenario.PulseType,
						RawData:   make([]byte, len(scenario.RawData)),
						Commands:  make([]models.Command, len(scenario.Commands)),
					}

					// Копируем данные
					copy(autoScenario.RawData, scenario.RawData)
					copy(autoScenario.Commands, scenario.Commands)

					// Добавляем AUTO-сценарий в коллекцию
					scenarios[autoName] = autoScenario
					fmt.Printf("Создан AUTO-сценарий: %s (на основе %s)\n", autoName, name)
				}
			}
		}
	}

	return scenarios
}

// AutoModeMenu выводит меню автоматического режима
func AutoModeMenu(scenarios map[string]models.Scenario) {
	// Подготовка сценариев для автоматизации
	scenarios = PrepareForAutomation(scenarios)

	fmt.Println("\nАвтоматический режим")
	fmt.Println("===================")

	// Настройки порта
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

	for {
		fmt.Println("\nВыберите действие:")
		fmt.Println("1. Отправить сценарий по параметрам")
		fmt.Println("2. Показать доступные AUTO-сценарии")
		fmt.Println("0. Вернуться в главное меню")

		var choice string
		fmt.Print("Выбор: ")
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			// Ввод параметров
			var pulseType int
			fmt.Print("Введите тип пульта (1-6): ")
			fmt.Scanln(&pulseType)

			if pulseType < 1 || pulseType > 6 {
				fmt.Println("Неверный тип пульта")
				continue
			}

			var distance int
			fmt.Print("Введите дистанцию (метры): ")
			fmt.Scanln(&distance)

			if distance <= 0 {
				fmt.Println("Неверная дистанция")
				continue
			}

			// Отправляем сценарий
			err := SendScenarioAuto(scenarios, portName, baudRate, byte(pulseType), distance)
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}

		case "2":
			// Показываем AUTO-сценарии
			fmt.Println("\nДоступные AUTO-сценарии:")
			count := 0
			for name, scenario := range scenarios {
				if strings.HasPrefix(name, AutoPrefix) {
					fmt.Printf("%s (Пульт: %d)\n", name, scenario.PulseType)
					count++
				}
			}

			if count == 0 {
				fmt.Println("AUTO-сценарии не найдены")
			}

		case "0":
			return

		default:
			fmt.Println("Неверный выбор, попробуйте снова")
		}
	}
}
