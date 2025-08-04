package main

import (
	"fmt"
	"tir/auto"
	"tir/models"
	"tir/protocol"
	"tir/storage"
	"tir/ui"
)

// Хранилище сценариев
var scenarios = map[string]models.Scenario{}

func main() {
	fmt.Println("Монорельсовая управляющая программа")
	fmt.Println("====================================")

	// Импортируем сохраненные сценарии
	protocol.ImportDefaultScenarios(scenarios)

	// Проверяем наличие файла сохраненных сценариев
	storage.LoadScenariosFromFile(scenarios)

	for {
		fmt.Println("\nГлавное меню:")
		fmt.Println("1. Подключиться к порту и отправить сценарий")
		fmt.Println("2. Конструктор сценариев")
		fmt.Println("3. Показать сохраненные сценарии")
		fmt.Println("4. Редактировать существующий сценарий")
		fmt.Println("5. Генерировать множество сценариев с рубежами")
		fmt.Println("6. Сохранить сценарии в файл")
		fmt.Println("7. Импорт сценария из HEX-строки")
		fmt.Println("8. Быстрое создание сценария (гарантированно работающего)")
		fmt.Println("9. Отладочная отправка сценария")
		fmt.Println("10. Автоматический режим (по типу пульта и дистанции)")
		fmt.Println("0. Выход")

		var choice string
		fmt.Print("Выберите действие: ")
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			ui.SendScenario(scenarios)
		case "2":
			ui.ScenarioConstructor(scenarios)
		case "3":
			ui.ShowScenarios(scenarios)
		case "4":
			ui.EditScenario(scenarios)
		case "5":
			ui.GenerateRangeScenarios(scenarios)
		case "6":
			storage.SaveScenariosToFile(scenarios)
		case "7":
			ui.ImportScenarioFromHex(scenarios)
		case "8":
			ui.QuickCreateScenario(scenarios)
		case "9":
			ui.DebugSendScenario(scenarios)
		case "10":
			auto.AutoModeMenu(scenarios)
		case "0":
			fmt.Println("Завершение работы...")
			return
		default:
			fmt.Println("Неверный выбор, попробуйте снова")
		}
	}
}
