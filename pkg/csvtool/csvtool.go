package csvtool

import (
	"encoding/csv"
	"github.com/google/uuid"
	"os"
)

func CreateFile(data [][]string, folder string) (string, error) {
	fileId := uuid.New()

	file, err := os.Create(folder + fileId.String() + ".csv")
	defer file.Close()
	if err != nil {
		return "", err
	}

	writer := csv.NewWriter(file)

	writer.Comma = ';'

	err = writer.WriteAll(data)
	if err != nil {
		return "", err
	}

	return fileId.String(), nil
}
