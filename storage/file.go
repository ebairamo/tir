package storage

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"tir/models"
	"tir/protocol"
)

// Сохранить сценарии в файл
func SaveScenariosToFile(scenarios map[string]models.Scenario) {
	fileName := "scenarios.txt"
	fmt.Print("Введите имя файла для сохранения (по умолчанию 'scenarios.txt'): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	customFileName := scanner.Text()

	if customFileName != "" {
		fileName = customFileName
		// Добавляем расширение, если его нет
		if !strings.HasSuffix(fileName, ".txt") {
			fileName += ".txt"
		}
	}

	fmt.Printf("Сохранение сценариев в файл %s...\n", fileName)

	var content strings.Builder

	// Сохраняем каждый сценарий в формате:
	// [имя]:[тип пульта]:[HEX-данные]
	for name, scenario := range scenarios {
		content.WriteString(name)
		content.WriteString(":")
		content.WriteString(fmt.Sprintf("%d", scenario.PulseType))
		content.WriteString(":")

		// Получаем данные сценария
		var data []byte
		if len(scenario.RawData) > 0 {
			data = scenario.RawData
		} else {
			data = protocol.GenerateScenarioPacket(scenario)
		}

		// Сохраняем данные в HEX-формате
		for _, b := range data {
			content.WriteString(fmt.Sprintf(" %02X", b))
		}
		content.WriteString("\n")
	}

	err := ioutil.WriteFile(fileName, []byte(content.String()), 0644)
	if err != nil {
		fmt.Printf("Ошибка сохранения файла: %v\n", err)
		return
	}

	fmt.Printf("Сценарии успешно сохранены в файл %s\n", fileName)
}

// Загрузить сценарии из файла
func LoadScenariosFromFile(scenarios map[string]models.Scenario) {
	fileName := "scenarios.txt"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Файл %s не найден, будут использоваться только встроенные сценарии\n", fileName)
		} else {
			fmt.Printf("Ошибка чтения файла %s: %v\n", fileName, err)
		}
		return
	}

	lines := strings.Split(string(data), "\n")
	loadedCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Разбираем строку формата: [имя]:[тип пульта]:[HEX-данные]
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			fmt.Printf("Некорректный формат строки: %s\n", line)
			continue
		}

		name := parts[0]

		// Извлекаем тип пульта
		var pulseType byte
		fmt.Sscanf(parts[1], "%d", &pulseType)
		if pulseType < 1 || pulseType > 6 {
			fmt.Printf("Некорректный тип пульта в сценарии %s: %s\n", name, parts[1])
			continue
		}

		// Декодируем HEX-данные
		hexData := strings.TrimSpace(parts[2])
		hexBytes := strings.Fields(hexData)

		var scenarioData []byte
		for _, h := range hexBytes {
			b, err := hex.DecodeString(h)
			if err != nil {
				fmt.Printf("Ошибка декодирования байта %s в сценарии %s: %v\n", h, name, err)
				continue
			}
			scenarioData = append(scenarioData, b[0])
		}

		// Создаем новый сценарий
		scenario := models.Scenario{
			Name:      name,
			PulseType: pulseType,
			RawData:   scenarioData,
		}

		// Пытаемся разобрать команды
		parsedScenario, err := protocol.ParseScenarioData(scenarioData)
		if err == nil && len(parsedScenario.Commands) > 0 {
			scenario.Commands = parsedScenario.Commands
		}

		// Сохраняем сценарий
		scenarios[name] = scenario
		loadedCount++
	}

	fmt.Printf("Загружено %d сценариев из файла %s\n", loadedCount, fileName)
}
