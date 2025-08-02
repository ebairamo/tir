package ui

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"tir/models"
	"tir/protocol"
)

// DisplayMainMenu отображает главное меню и возвращает выбор пользователя
func DisplayMainMenu() string {
	fmt.Println("\nГлавное меню:")
	fmt.Println("1. Подключиться к порту и отправить сценарий")
	fmt.Println("2. Конструктор сценариев")
	fmt.Println("3. Показать сохраненные сценарии")
	fmt.Println("4. Редактировать существующий сценарий")
	fmt.Println("5. Генерировать множество сценариев с рубежами")
	fmt.Println("6. Сохранить сценарии в файл")
	fmt.Println("7. Импорт сценария из HEX-строки")
	fmt.Println("8. Выход")

	var choice string
	fmt.Print("Выберите действие: ")
	fmt.Scanln(&choice)

	return choice
}

// ShowScenarios отображает все сохраненные сценарии
func ShowScenarios(scenarios map[string]models.Scenario) {
	fmt.Println("\nСохраненные сценарии:")
	fmt.Println("=====================")

	if len(scenarios) == 0 {
		fmt.Println("Нет сохраненных сценариев")
		return
	}

	for name, scenario := range scenarios {
		fmt.Printf("\nСценарий: %s (Пульт: %d)\n", name, scenario.PulseType)

		// Показываем команды, если они есть
		if len(scenario.Commands) > 0 {
			fmt.Println("Команды:")
			for i, cmd := range scenario.Commands {
				if cmd.HasParam {
					fmt.Printf("  %d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
				} else {
					fmt.Printf("  %d. %s\n", i+1, cmd.Name)
				}
			}
		} else if len(scenario.RawData) > 0 {
			// Если структурированные команды не распознаны, показываем сырые данные
			fmt.Printf("Размер данных: %d байт\n", len(scenario.RawData))
			fmt.Print("Данные: ")

			// Выводим до 32 байт данных
			maxShow := len(scenario.RawData)
			if maxShow > 32 {
				maxShow = 32
			}

			for i := 0; i < maxShow; i++ {
				fmt.Printf("%02X", scenario.RawData[i])
				if i < maxShow-1 {
					fmt.Print(" ")
				}
			}

			if len(scenario.RawData) > 32 {
				fmt.Print("...")
			}
			fmt.Println()
		}
	}
}

// ImportScenarioFromHex импортирует сценарий из HEX-строки
func ImportScenarioFromHex(scenarios map[string]models.Scenario) {
	fmt.Println("\nИмпорт сценария из HEX-строки")
	fmt.Println("============================")

	// Ввод имени сценария
	var name string
	fmt.Print("Введите имя для импортируемого сценария: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name = scanner.Text()

	if name == "" {
		fmt.Println("Имя сценария не может быть пустым")
		return
	}

	// Проверяем, существует ли уже такой сценарий
	_, exists := scenarios[name]
	if exists {
		fmt.Printf("Сценарий '%s' уже существует. Хотите перезаписать? (да/нет): ", name)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "да" && confirm != "д" && confirm != "yes" && confirm != "y" {
			fmt.Println("Операция отменена")
			return
		}
	}

	// Ввод данных сценария
	fmt.Println("Введите данные сценария в шестнадцатеричном формате (например, 7E 00 01 13...):")
	fmt.Println("Для завершения ввода нажмите Enter на пустой строке")

	var hexData string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		hexData += " " + line
	}

	// Очищаем и нормализуем входные данные
	hexData = strings.ToUpper(strings.TrimSpace(hexData))
	hexData = strings.ReplaceAll(hexData, ",", " ")
	hexData = strings.ReplaceAll(hexData, "0X", "")
	hexData = strings.ReplaceAll(hexData, "0x", "")

	// Разбиваем на отдельные байты
	hexBytes := strings.Fields(hexData)

	// Конвертируем в бинарные данные
	var data []byte
	for _, h := range hexBytes {
		if len(h) != 2 {
			fmt.Printf("Ошибка в формате байта: %s. Должно быть ровно 2 символа.\n", h)
			return
		}

		b, err := hex.DecodeString(h)
		if err != nil {
			fmt.Printf("Ошибка декодирования байта %s: %v\n", h, err)
			return
		}

		data = append(data, b[0])
	}

	// Проверяем минимальную длину пакета
	if len(data) < 15 {
		fmt.Println("Ошибка: пакет слишком короткий для действительного сценария")
		return
	}

	// Проверяем правильность заголовка
	if data[0] != 0x7E || data[1] != 0x00 || data[2] < 0x01 || data[2] > 0x06 {
		fmt.Println("Ошибка: неверный формат заголовка сценария")
		return
	}

	// Пытаемся разобрать сценарий
	parsedScenario, err := protocol.ParseScenarioData(data)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось полностью разобрать сценарий: %v\n", err)
		fmt.Println("Сценарий будет импортирован только как сырые данные")

		// Создаем сценарий только с сырыми данными
		scenario := models.Scenario{
			Name:      name,
			PulseType: data[2], // Берем тип пульта из заголовка
			RawData:   data,
		}

		scenarios[name] = scenario
	} else {
		// Если разбор успешен, сохраняем полностью структурированный сценарий
		parsedScenario.Name = name
		parsedScenario.RawData = data
		scenarios[name] = parsedScenario

		// Показываем команды
		fmt.Println("\nРаспознанные команды в сценарии:")
		for i, cmd := range parsedScenario.Commands {
			if cmd.HasParam {
				fmt.Printf("%d. %s (%s: %d)\n", i+1, cmd.Name, cmd.ParamName, cmd.ParamValue)
			} else {
				fmt.Printf("%d. %s\n", i+1, cmd.Name)
			}
		}
	}

	fmt.Printf("Сценарий '%s' успешно импортирован (%d байт)\n", name, len(data))
}
