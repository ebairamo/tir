package ui

import (
	"bufio"
	"fmt"
	"os"
	"tir/models"
	"tir/protocol"
)

// EditScenario позволяет редактировать существующий сценарий
func EditScenario(scenarios map[string]models.Scenario) {
	fmt.Println("\nРедактирование сценария")
	fmt.Println("======================")

	// Список доступных сценариев
	fmt.Println("Доступные сценарии:")
	var scenarioNames []string
	i := 1
	for name := range scenarios {
		fmt.Printf("%d. %s\n", i, name)
		scenarioNames = append(scenarioNames, name)
		i++
	}

	if len(scenarioNames) == 0 {
		fmt.Println("Нет доступных сценариев для редактирования")
		return
	}

	// Выбор сценария
	var choice int
	fmt.Print("Выберите сценарий для редактирования (номер): ")
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(scenarioNames) {
		fmt.Println("Неверный выбор сценария")
		return
	}

	// Получаем выбранный сценарий
	selectedName := scenarioNames[choice-1]
	scenario := scenarios[selectedName]

	// Если сценарий был импортирован только как сырые данные, пробуем разобрать его
	if len(scenario.Commands) == 0 && len(scenario.RawData) > 0 {
		parsedScenario, err := protocol.ParseScenarioData(scenario.RawData)
		if err != nil {
			fmt.Printf("Не удалось разобрать сценарий: %v\n", err)
			fmt.Println("Будет создан новый пустой сценарий с тем же именем")
			scenario = models.Scenario{
				Name:      selectedName,
				PulseType: scenario.PulseType,
				Commands:  []models.Command{},
				RawData:   scenario.RawData, // Сохраняем оригинальные данные
			}
		} else {
			scenario = parsedScenario
		}
	}

	// Показываем текущие команды
	fmt.Println("\nТекущие команды в сценарии:")
	for i, cmd := range scenario.Commands {
		if cmd.HasParam {
			fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
		} else {
			fmt.Printf("%d. %s\n", i+1, cmd.Name)
		}
	}

	// Меню редактирования
	for {
		fmt.Println("\nМеню редактирования:")
		fmt.Println("1. Добавить команду")
		fmt.Println("2. Удалить команду")
		fmt.Println("3. Изменить порядок команд")
		fmt.Println("4. Изменить имя сценария")
		fmt.Println("5. Изменить тип пульта")
		fmt.Println("6. Сохранить и выйти")
		fmt.Println("7. Выйти без сохранения")

		var editChoice string
		fmt.Print("Выберите действие: ")
		fmt.Scanln(&editChoice)

		switch editChoice {
		case "1":
			// Добавить команду
			addCommandToScenario(&scenario)
		case "2":
			// Удалить команду
			deleteCommandFromScenario(&scenario)
		case "3":
			// Изменить порядок команд
			reorderCommandsInScenario(&scenario)
		case "4":
			// Изменить имя сценария
			fmt.Print("Введите новое имя сценария: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			newName := scanner.Text()

			if newName != "" && newName != selectedName {
				// Проверяем, не существует ли уже сценарий с таким именем
				_, exists := scenarios[newName]
				if exists {
					fmt.Printf("Сценарий с именем '%s' уже существует. Хотите перезаписать? (да/нет): ", newName)
					var confirm string
					fmt.Scanln(&confirm)
					if confirm != "да" && confirm != "д" && confirm != "yes" && confirm != "y" {
						fmt.Println("Изменение имени отменено")
						continue
					}
				}

				// Обновляем имя
				delete(scenarios, selectedName)
				selectedName = newName
				scenario.Name = newName
				scenarios[newName] = scenario
				fmt.Printf("Имя сценария изменено на '%s'\n", newName)
			} else {
				fmt.Println("Имя сценария не изменено")
			}
		case "5":
			// Изменить тип пульта
			fmt.Print("Введите новый номер пульта (1-6): ")
			var newPulseType int
			fmt.Scanln(&newPulseType)

			if newPulseType >= 1 && newPulseType <= 6 {
				scenario.PulseType = byte(newPulseType)
				fmt.Printf("Тип пульта изменен на %d\n", newPulseType)
			} else {
				fmt.Println("Некорректный номер пульта")
			}
		case "6":
			// Сохранить и выйти
			// Генерируем бинарный пакет
			scenario.RawData = protocol.GenerateScenarioPacket(scenario)
			scenarios[selectedName] = scenario
			fmt.Printf("Сценарий '%s' успешно сохранен\n", selectedName)
			return
		case "7":
			// Выйти без сохранения
			fmt.Println("Изменения отменены")
			return
		default:
			fmt.Println("Неверный выбор, попробуйте снова")
		}
	}
}

// Добавить команду в сценарий
func addCommandToScenario(scenario *models.Scenario) {
	fmt.Println("\nДоступные команды:")

	// Вывод списка простых команд
	i := 1
	var commandNames []string
	for name := range models.CommandMap {
		fmt.Printf("%d. %s\n", i, name)
		commandNames = append(commandNames, name)
		i++
	}

	// Выбор команды
	var cmdChoice int
	fmt.Print("Выберите команду (номер): ")
	fmt.Scanln(&cmdChoice)

	if cmdChoice < 1 || cmdChoice > len(commandNames) {
		fmt.Println("Неверный выбор, команда не добавлена")
		return
	}

	// Получаем выбранную команду
	selectedCmd := commandNames[cmdChoice-1]
	cmdCode := models.CommandMap[selectedCmd]

	// Создаем объект команды
	command := models.Command{
		Name: selectedCmd,
		Code: cmdCode,
	}

	// Проверяем, требует ли команда параметр
	_, hasParam := models.ParamCommands[cmdCode]
	if hasParam {
		command.HasParam = true
		command.ParamName = models.ParamCommands[cmdCode]

		// Запрашиваем значение параметра
		fmt.Printf("Введите %s: ", command.ParamName)
		var paramValue uint16
		fmt.Scanln(&paramValue)
		command.ParamValue = paramValue
	}

	// Выбираем позицию для вставки
	if len(scenario.Commands) > 0 {
		fmt.Printf("\nВыберите позицию для вставки (1-%d) или 0 для добавления в конец: ", len(scenario.Commands))
		var position int
		fmt.Scanln(&position)

		if position < 0 || position > len(scenario.Commands) {
			fmt.Println("Неверная позиция, добавляем в конец")
			position = 0
		}

		if position == 0 {
			// Добавляем в конец
			scenario.Commands = append(scenario.Commands, command)
		} else {
			// Вставляем в указанную позицию
			// Вставляем в указанную позицию
			newCommands := make([]models.Command, 0, len(scenario.Commands)+1)
			newCommands = append(newCommands, scenario.Commands[:position-1]...)
			newCommands = append(newCommands, command)
			newCommands = append(newCommands, scenario.Commands[position-1:]...)
			scenario.Commands = newCommands
		}
	} else {
		// Если сценарий пуст, просто добавляем команду
		scenario.Commands = append(scenario.Commands, command)
	}

	// Выводим обновленный список команд
	fmt.Println("\nОбновленный список команд:")
	for i, cmd := range scenario.Commands {
		if cmd.HasParam {
			fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
		} else {
			fmt.Printf("%d. %s\n", i+1, cmd.Name)
		}
	}
}

// Удалить команду из сценария
func deleteCommandFromScenario(scenario *models.Scenario) {
	if len(scenario.Commands) == 0 {
		fmt.Println("Сценарий не содержит команд")
		return
	}

	fmt.Println("\nТекущие команды в сценарии:")
	for i, cmd := range scenario.Commands {
		if cmd.HasParam {
			fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
		} else {
			fmt.Printf("%d. %s\n", i+1, cmd.Name)
		}
	}

	fmt.Print("Выберите номер команды для удаления: ")
	var position int
	fmt.Scanln(&position)

	if position < 1 || position > len(scenario.Commands) {
		fmt.Println("Неверный номер команды")
		return
	}

	// Удаляем команду
	deletedCmd := scenario.Commands[position-1]
	scenario.Commands = append(scenario.Commands[:position-1], scenario.Commands[position:]...)

	if deletedCmd.HasParam {
		fmt.Printf("Удалена команда: %s (%s: %d)\n", deletedCmd.Name, deletedCmd.ParamName, deletedCmd.ParamValue)
	} else {
		fmt.Printf("Удалена команда: %s\n", deletedCmd.Name)
	}

	// Выводим обновленный список команд
	fmt.Println("\nОбновленный список команд:")
	for i, cmd := range scenario.Commands {
		if cmd.HasParam {
			fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
		} else {
			fmt.Printf("%d. %s\n", i+1, cmd.Name)
		}
	}
}

// Изменить порядок команд в сценарии
func reorderCommandsInScenario(scenario *models.Scenario) {
	if len(scenario.Commands) < 2 {
		fmt.Println("Недостаточно команд для изменения порядка")
		return
	}

	fmt.Println("\nТекущие команды в сценарии:")
	for i, cmd := range scenario.Commands {
		if cmd.HasParam {
			fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
		} else {
			fmt.Printf("%d. %s\n", i+1, cmd.Name)
		}
	}

	fmt.Print("Выберите номер команды для перемещения: ")
	var sourcePos int
	fmt.Scanln(&sourcePos)

	if sourcePos < 1 || sourcePos > len(scenario.Commands) {
		fmt.Println("Неверный номер команды")
		return
	}

	fmt.Print("Выберите новую позицию для команды: ")
	var targetPos int
	fmt.Scanln(&targetPos)

	if targetPos < 1 || targetPos > len(scenario.Commands) {
		fmt.Println("Неверная позиция")
		return
	}

	if sourcePos == targetPos {
		fmt.Println("Команда уже находится в указанной позиции")
		return
	}

	// Перемещаем команду
	cmd := scenario.Commands[sourcePos-1]

	// Удаляем с исходной позиции
	scenario.Commands = append(scenario.Commands[:sourcePos-1], scenario.Commands[sourcePos:]...)

	// Вставляем в новую позицию
	if targetPos > len(scenario.Commands) {
		// Если указана позиция за пределами массива, добавляем в конец
		scenario.Commands = append(scenario.Commands, cmd)
	} else {
		// Корректируем позицию, если команда перемещается вниз
		if sourcePos < targetPos {
			targetPos--
		}

		newCommands := make([]models.Command, 0, len(scenario.Commands)+1)
		newCommands = append(newCommands, scenario.Commands[:targetPos-1]...)
		newCommands = append(newCommands, cmd)
		newCommands = append(newCommands, scenario.Commands[targetPos-1:]...)
		scenario.Commands = newCommands
	}

	fmt.Printf("Команда перемещена с позиции %d на позицию %d\n", sourcePos, targetPos)

	// Выводим обновленный список команд
	fmt.Println("\nОбновленный список команд:")
	for i, cmd := range scenario.Commands {
		if cmd.HasParam {
			fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
		} else {
			fmt.Printf("%d. %s\n", i+1, cmd.Name)
		}
	}
}
