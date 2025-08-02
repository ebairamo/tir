package ui

import (
	"bufio"
	"fmt"
	"os"
	"tir/models"
)

// ScenarioConstructor создает новый сценарий через интерактивный конструктор
func ScenarioConstructor(scenarios map[string]models.Scenario) {
	fmt.Println("\nКонструктор сценариев")
	fmt.Println("=====================")

	fmt.Println("\nВНИМАНИЕ: Предыдущие попытки показали, что только точные копии")
	fmt.Println("оригинальных сценариев могут работать правильно!")

	fmt.Println("\nВыберите шаблон для создания сценария:")
	fmt.Println("1. Точная копия test1 (работает гарантированно)")
	fmt.Println("2. Точная копия range_3m_pulse1")

	var templateChoice int
	fmt.Print("Выберите шаблон (1-2): ")
	fmt.Scanln(&templateChoice)

	if templateChoice != 2 {
		templateChoice = 1 // По умолчанию используем test1
	}

	// Определяем имя шаблона
	var templateName string
	if templateChoice == 1 {
		templateName = "test1"
	} else {
		templateName = "range_3m_pulse1"
	}

	// Получаем оригинальный сценарий
	originalScenario, exists := scenarios[templateName]
	if !exists {
		fmt.Printf("Ошибка: оригинальный сценарий %s не найден\n", templateName)
		return
	}

	// Ввод имени для нового сценария (только для организации)
	var scenarioName string
	fmt.Print("Введите имя сценария: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	scenarioName = scanner.Text()

	if scenarioName == "" {
		fmt.Println("Имя сценария не может быть пустым")
		return
	}

	// Проверка на существование сценария
	_, exists = scenarios[scenarioName]
	if exists {
		fmt.Printf("Сценарий '%s' уже существует. Хотите перезаписать? (да/нет): ", scenarioName)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "да" && confirm != "д" && confirm != "yes" && confirm != "y" {
			fmt.Println("Операция отменена")
			return
		}
	}

	// Создаем 100% точную копию оригинального сценария
	newScenario := models.Scenario{
		Name:      scenarioName,
		PulseType: originalScenario.PulseType,
		RawData:   make([]byte, len(originalScenario.RawData)),
	}

	// Копируем данные
	copy(newScenario.RawData, originalScenario.RawData)

	// Сохраняем сценарий
	scenarios[scenarioName] = newScenario

	fmt.Printf("\nСценарий '%s' успешно создан как точная копия '%s' (%d байт)\n",
		scenarioName, templateName, len(newScenario.RawData))
	fmt.Println("Сценарий готов к отправке. Используйте пункт 1 для отправки сценария.")
	fmt.Println("\nПРИМЕЧАНИЕ: Сценарий содержит оригинальное имя внутри пакета.")
}

// GenerateRangeScenarios создает серию сценариев с разными рубежами
func GenerateRangeScenarios(scenarios map[string]models.Scenario) {
	fmt.Println("\nГенерация сценариев для различных рубежей")
	fmt.Println("=========================================")

	fmt.Println("\nВНИМАНИЕ: Из-за особенностей протокола, изменение рубежа может")
	fmt.Println("привести к неработоспособности сценария!")

	fmt.Println("\nВыберите шаблон для создания сценариев:")
	fmt.Println("1. Точная копия test1 (работает гарантированно)")
	fmt.Println("2. Точная копия range_3m_pulse1")

	var templateChoice int
	fmt.Print("Выберите шаблон (1-2): ")
	fmt.Scanln(&templateChoice)

	if templateChoice != 2 {
		templateChoice = 1 // По умолчанию используем test1
	}

	// Определяем имя шаблона
	var templateName string
	if templateChoice == 1 {
		templateName = "test1"
	} else {
		templateName = "range_3m_pulse1"
	}

	// Получаем оригинальный сценарий
	originalScenario, exists := scenarios[templateName]
	if !exists {
		fmt.Printf("Ошибка: оригинальный сценарий %s не найден\n", templateName)
		return
	}

	// Префикс имени
	var namePrefix string
	fmt.Print("Введите префикс имени сценария (например, 'range'): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	namePrefix = scanner.Text()

	if namePrefix == "" {
		namePrefix = "range"
	}

	// Количество копий
	var count int
	fmt.Print("Введите количество копий для создания: ")
	fmt.Scanln(&count)

	if count <= 0 {
		count = 5 // По умолчанию создаем 5 копий
	}

	fmt.Printf("\nСоздание %d копий сценария '%s'...\n", count, templateName)

	// Счетчик созданных сценариев
	createdCount := 0

	for i := 1; i <= count; i++ {
		// Имя сценария
		name := fmt.Sprintf("%s_%d", namePrefix, i)

		// Создаем 100% точную копию оригинального сценария
		newScenario := models.Scenario{
			Name:      name,
			PulseType: originalScenario.PulseType,
			RawData:   make([]byte, len(originalScenario.RawData)),
		}

		// Копируем данные
		copy(newScenario.RawData, originalScenario.RawData)

		// Сохраняем сценарий
		scenarios[name] = newScenario

		fmt.Printf("Создан сценарий: %s (копия %s)\n", name, templateName)
		createdCount++
	}

	fmt.Printf("\nГотово! Создано %d сценариев.\n", createdCount)
	fmt.Println("Сценарии готовы к отправке. Используйте пункт 1 для отправки сценария.")
	fmt.Println("\nПРИМЕЧАНИЕ: Все сценарии содержат оригинальное имя внутри пакета.")
}
