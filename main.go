// main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"tir/auto"
	"tir/firebase" // Импортируем новый пакет
	"tir/models"
	"tir/protocol"
	"tir/storage"
	"tir/ui"
)

// Хранилище сценариев
var scenarios = map[string]models.Scenario{}

// Клиент автоматической отправки
var restClient *firebase.RestClient

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
		fmt.Println("11. Запустить отслеживание изменений в Firebase")      // Мониторинг
		fmt.Println("12. Автоматическая отправка при изменении в Firebase") // Автоматическая отправка
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
		case "11":
			// Запускаем отслеживание изменений
			startMonitoring()
		case "12":
			// Запускаем автоматическую отправку при изменении
			startAutoSender()
		case "0":
			fmt.Println("Завершение работы...")
			// Закрываем соединение, если оно открыто
			if restClient != nil {
				restClient.Close()
			}
			return
		default:
			fmt.Println("Неверный выбор, попробуйте снова")
		}
	}
}

// getFirebaseCredentials возвращает учетные данные Firebase
func getFirebaseCredentials() (string, string) {
	// Правильные значения для вашего проекта
	projectID := "keiki-mergen-c4f1b"
	apiKey := "AIzaSyDNa90KsHCh9LGG0qpUJFD0GwH5McQK-X4"

	return projectID, apiKey
}

// startMonitoring запускает отслеживание изменений
func startMonitoring() {
	fmt.Println("\nЗапуск отслеживания изменений в Firebase")
	fmt.Println("=========================================")

	// Получаем учетные данные Firebase
	projectID, apiKey := getFirebaseCredentials()

	fmt.Println("Инициализация клиента мониторинга...")
	client := firebase.NewRestClient(projectID, apiKey, scenarios)

	fmt.Println("Клиент мониторинга успешно инициализирован")
	fmt.Println("Запуск отслеживания изменений...")
	fmt.Println("Для возврата в главное меню нажмите Ctrl+C")

	// Запускаем отслеживание изменений
	client.ListenToTargetLines()

	// Закрываем соединение
	client.Close()
}

// startAutoSender запускает автоматическую отправку при изменении
func startAutoSender() {
	fmt.Println("\nАвтоматическая отправка при изменении в Firebase")
	fmt.Println("==============================================")

	// Если уже запущен, останавливаем
	if restClient != nil && restClient.Running {
		restClient.StopAutoSender()
		fmt.Println("Автоматическая отправка остановлена")
		return
	}

	// Получаем учетные данные Firebase
	projectID, apiKey := getFirebaseCredentials()

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

	fmt.Println("Инициализация клиента автоматической отправки...")

	// Инициализируем клиент
	restClient = firebase.NewRestClient(projectID, apiKey, scenarios)

	// Устанавливаем настройки порта
	restClient.SetPortSettings(portName, baudRate)

	// Запускаем автоматическую отправку
	err := restClient.StartAutoSender()
	if err != nil {
		fmt.Printf("Ошибка запуска: %v\n", err)
		return
	}

	fmt.Println("Автоматическая отправка успешно запущена")
	fmt.Println("Программа будет автоматически отправлять сценарии при изменении в Firebase")
	fmt.Println("Для возврата в главное меню нажмите Enter (автоматическая отправка продолжится в фоне)")
	fmt.Println("Для остановки автоматической отправки выберите пункт 12 повторно")

	// Настраиваем обработку сигналов для корректного завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем нажатия Enter или сигнала завершения
	go func() {
		fmt.Scanln()
		// Возврат в главное меню, но автоматическая отправка продолжается
	}()

	select {
	case <-sigChan:
		// Получен сигнал завершения
		fmt.Println("\nПолучен сигнал завершения, останавливаем автоматическую отправку...")
		restClient.StopAutoSender()
		restClient.Close()
		restClient = nil
		os.Exit(0)
	}
}
