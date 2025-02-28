package txtfile

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

// Проверяем, содержит ли файл заявку
func ContainsRequest(reqID int, filename string, UpdateTime string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, возвращаем false
			return false, nil
		}
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if UpdateTime == "" {
			if strconv.Itoa(reqID) == line {
				return true, nil
			}
		} else {
			if strconv.Itoa(reqID)+" "+UpdateTime == line {
				return true, nil
			}
		}
	}
	return false, scanner.Err()
}

// Добавляем новую заявку в файл
func AddRequestToFile(reqID int, filename string, UpdatedTime string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if UpdatedTime == "" {
		_, err = file.WriteString(fmt.Sprintf("%d\n", reqID))
		return err
	} else {
		_, err = file.WriteString(fmt.Sprintf("%d %s\n", reqID, UpdatedTime))
		return err
	}

}
