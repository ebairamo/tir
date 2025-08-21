// firebase/firebase_rest.go
package firebase

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"tir/auto"
	"tir/models"
)

// RestClient клиент для работы с Firebase REST API
type RestClient struct {
	Running    bool
	PortName   string
	BaudRate   uint32
	Scenarios  map[string]models.Scenario
	LastValues map[string]int // lineID -> last distance
	ProjectID  string
	ApiKey     string
}

// NewRestClient создает новый REST клиент
func NewRestClient(projectID, apiKey string, scenarios map[string]models.Scenario) *RestClient {
	return &RestClient{
		Running:    false,
		PortName:   "COM4",
		BaudRate:   4800,
		Scenarios:  scenarios,
		LastValues: make(map[string]int),
		ProjectID:  projectID,
		ApiKey:     apiKey,
	}
}

// SetPortSettings устанавливает настройки порта
func (rc *RestClient) SetPortSettings(portName string, baudRate uint32) {
	rc.PortName = portName
	rc.BaudRate = baudRate
}

// StartAutoSender запускает автоматическое отслеживание и отправку
func (rc *RestClient) StartAutoSender() error {
	if rc.Running {
		return fmt.Errorf("автоматическая отправка уже запущена")
	}

	fmt.Println("Запуск автоматической отправки сценариев...")
	fmt.Printf("Порт: %s, скорость: %d бод\n", rc.PortName, rc.BaudRate)
	fmt.Println("Подключение к Firebase Firestore...")
	fmt.Printf("Проект: %s\n", rc.ProjectID)

	// Запускаем подготовку сценариев для автоматизации
	rc.Scenarios = auto.PrepareForAutomation(rc.Scenarios)

	// Считываем начальные значения
	lines, err := rc.getFirestoreLines()
	if err != nil {
		fmt.Printf("Ошибка при получении начальных значений: %v\n", err)
	} else {
		for lineID, distance := range lines {
			lineNum, err := getLineNumber(lineID)
			if err == nil {
				rc.LastValues[lineID] = distance
				fmt.Printf("Начальное значение для линии %d (ID: %s): дистанция %d м\n",
					lineNum, lineID, distance)
			}
		}
	}

	rc.Running = true

	// Запускаем обработку в отдельной горутине
	go func() {
		for rc.Running {
			// Получаем текущие значения из Firestore
			currentLines, err := rc.getFirestoreLines()
			if err != nil {
				fmt.Printf("Ошибка при запросе к Firebase: %v\n", err)
				time.Sleep(time.Second * 5)
				continue
			}

			// Проверяем изменения по каждой линии
			for lineID, distance := range currentLines {
				lineNum, err := getLineNumber(lineID)
				if err != nil {
					continue // Пропускаем линии с неверным ID
				}

				lastDistance, exists := rc.LastValues[lineID]

				// Если дистанция изменилась или это новая линия
				if !exists || lastDistance != distance {
					fmt.Printf("\n[%s] Обнаружено изменение в линии %d (ID: %s): дистанция изменена с %d на %d\n",
						time.Now().Format("2006-01-02 15:04:05"),
						lineNum, lineID,
						lastDistance, distance)

					// Обновляем последнее известное значение
					rc.LastValues[lineID] = distance

					// Отправляем сценарий
					err := auto.SendScenarioAuto(rc.Scenarios, rc.PortName, rc.BaudRate,
						byte(lineNum), distance)
					if err != nil {
						fmt.Printf("Ошибка при отправке сценария: %v\n", err)
					} else {
						fmt.Printf("Сценарий для линии %d с дистанцией %d м успешно отправлен\n",
							lineNum, distance)
					}
				}
			}

			time.Sleep(time.Second * 2) // Проверка каждые 2 секунды
		}
	}()

	return nil
}

// getFirestoreLines получает информацию о линиях из Firestore
func (rc *RestClient) getFirestoreLines() (map[string]int, error) {
	result := make(map[string]int)

	// Формируем URL
	collectionUrl := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents/target_lines?key=%s",
		url.QueryEscape(rc.ProjectID), url.QueryEscape(rc.ApiKey))

	// Отправляем запрос
	resp, err := http.Get(collectionUrl)
	if err != nil {
		return result, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем содержимое ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Проверяем код ответа
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("неверный код ответа: %d", resp.StatusCode)
	}

	// Структура ответа Firestore
	var response struct {
		Documents []struct {
			Name   string                            `json:"name"`
			Fields map[string]map[string]interface{} `json:"fields"`
		} `json:"documents"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return result, fmt.Errorf("ошибка при разборе JSON: %v", err)
	}

	// Обрабатываем документы
	for _, doc := range response.Documents {
		// Извлекаем ID документа из пути
		parts := strings.Split(doc.Name, "/")
		if len(parts) == 0 {
			continue
		}

		lineID := parts[len(parts)-1]

		// Проверяем наличие поля distance
		distanceField, hasDistance := doc.Fields["distance"]
		if !hasDistance {
			continue
		}

		// Извлекаем значение distance в зависимости от типа
		var distanceValue interface{}
		var distanceType string

		for fieldType, fieldValue := range distanceField {
			distanceType = fieldType
			distanceValue = fieldValue
			break
		}

		// Преобразуем значение в число
		var distance int

		switch distanceType {
		case "integerValue":
			if strValue, ok := distanceValue.(string); ok {
				distance, _ = strconv.Atoi(strValue)
			} else if floatValue, ok := distanceValue.(float64); ok {
				distance = int(floatValue)
			}
		case "stringValue":
			if strValue, ok := distanceValue.(string); ok {
				distance, _ = strconv.Atoi(strValue)
			}
		case "doubleValue":
			if floatValue, ok := distanceValue.(float64); ok {
				distance = int(floatValue)
			}
		}

		if distance > 0 {
			result[lineID] = distance
		}
	}

	return result, nil
}

// getLineNumber извлекает номер линии из ID
func getLineNumber(lineID string) (int, error) {
	// Проверяем, соответствует ли ID формату "line_X"
	if strings.HasPrefix(lineID, "line_") {
		// Извлекаем номер из ID формата "line_X"
		lineNumStr := strings.TrimPrefix(lineID, "line_")
		lineNum, err := strconv.Atoi(lineNumStr)
		if err == nil && lineNum >= 1 && lineNum <= 6 {
			return lineNum, nil
		}
		return 0, fmt.Errorf("неверный номер линии: %s", lineNumStr)
	}

	// Проверяем, соответствует ли ID формату "lineX"
	if strings.HasPrefix(lineID, "line") {
		// Извлекаем номер из ID формата "lineX"
		lineNumStr := strings.TrimPrefix(lineID, "line")
		lineNum, err := strconv.Atoi(lineNumStr)
		if err == nil && lineNum >= 1 && lineNum <= 6 {
			return lineNum, nil
		}
		return 0, fmt.Errorf("неверный номер линии: %s", lineNumStr)
	}

	// Если ID не имеет префикса "line" или "line_", пробуем напрямую преобразовать в число
	lineNum, err := strconv.Atoi(lineID)
	if err == nil && lineNum >= 1 && lineNum <= 6 {
		return lineNum, nil
	}

	return 0, fmt.Errorf("неверный формат ID линии: %s", lineID)
}

// StopAutoSender останавливает автоматическую отправку
func (rc *RestClient) StopAutoSender() {
	rc.Running = false
	fmt.Println("Остановка автоматической отправки...")
}

// Close закрывает соединение
func (rc *RestClient) Close() {
	rc.Running = false
}

// ListenToTargetLines отслеживает изменения в target_lines
func (rc *RestClient) ListenToTargetLines() {
	fmt.Println("Начинаем отслеживание изменений в target_lines в Firestore...")

	for {
		lines, err := rc.getFirestoreLines()
		if err != nil {
			fmt.Printf("Ошибка при запросе к Firebase: %v\n", err)
		} else if len(lines) > 0 {
			fmt.Printf("[%s] Обнаружено %d линий в Firestore:\n",
				time.Now().Format("2006-01-02 15:04:05"), len(lines))

			for lineID, distance := range lines {
				lineNum, _ := getLineNumber(lineID)
				fmt.Printf("  Линия %d (ID: %s): Дистанция %d м\n",
					lineNum, lineID, distance)
			}
		} else {
			fmt.Println("Линии в Firestore не найдены")
		}

		time.Sleep(time.Second * 5)
	}
}
