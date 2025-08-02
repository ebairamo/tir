package ui

import (
	"bufio"
	"fmt"
	"os"
	"tir/models"
)

// QuickCreateScenario создает новый сценарий на основе рабочего шаблона
func QuickCreateScenario(scenarios map[string]models.Scenario) {
	fmt.Println("\nБыстрое создание сценария")
	fmt.Println("========================")

	fmt.Println("Внимание! Предыдущие попытки показали, что устройство может требовать")
	fmt.Println("точно такие же сценарии, как оригинальные, включая имя.")

	fmt.Println("\nВыберите действие:")
	fmt.Println("1. Создать точную копию рабочего сценария range_3m_pulse1")
	fmt.Println("2. Создать точную копию рабочего сценария test1")

	var choice int
	fmt.Print("Выберите действие (1-2): ")
	fmt.Scanln(&choice)

	if choice != 1 && choice != 2 {
		fmt.Println("Неверный выбор. Используем range_3m_pulse1")
		choice = 1
	}

	var scenarioName string
	var originalScenario models.Scenario
	var exists bool

	if choice == 1 {
		scenarioName = "range_3m_pulse1"
	} else {
		scenarioName = "test1"
	}

	// Проверяем, существует ли оригинальный сценарий
	originalScenario, exists = scenarios[scenarioName]
	if !exists {
		fmt.Printf("Ошибка: оригинальный сценарий %s не найден в базе данных\n", scenarioName)
		return
	}

	// Предлагаем сохранить копию под новым именем
	fmt.Printf("Создание точной копии сценария %s\n", scenarioName)
	fmt.Print("Желаете также сохранить копию под другим именем? (да/нет): ")
	var saveAlso string
	fmt.Scanln(&saveAlso)

	if saveAlso == "да" || saveAlso == "д" || saveAlso == "yes" || saveAlso == "y" {
		fmt.Print("Введите имя для копии: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		copyName := scanner.Text()

		if copyName != "" && copyName != scenarioName {
			// Создаем глубокую копию сценария
			scenarioCopy := models.Scenario{
				Name:      copyName,
				PulseType: originalScenario.PulseType,
				RawData:   make([]byte, len(originalScenario.RawData)),
			}

			// Копируем данные
			copy(scenarioCopy.RawData, originalScenario.RawData)

			// Сохраняем копию
			scenarios[copyName] = scenarioCopy
			fmt.Printf("Копия сценария сохранена под именем '%s'\n", copyName)
		}
	}

	fmt.Printf("\nВНИМАНИЕ: Используйте сценарий '%s' для отправки на устройство\n", scenarioName)
	fmt.Println("Этот сценарий точно такой же, как в оригинальном работающем примере.")
}
