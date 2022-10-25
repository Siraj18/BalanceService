package utils

import (
	"fmt"
	"github.com/siraj18/balance-service-new/internal/models"
	"github.com/siraj18/balance-service-new/pkg/csvtool"
)

const folder = "./files/reports/"

func GenerateReportsLink(reserves *[]models.Reserve, host string) (string, error) {
	serviceGain := make(map[string]float64)

	for _, j := range *reserves {
		serviceGain[j.ServiceId] += j.Amount
	}

	data := make([][]string, len(serviceGain))

	count := 0

	for i, j := range serviceGain {
		data[count] = []string{i, fmt.Sprintf("%f", j)}
		count++
	}

	fileId, err := csvtool.CreateFile(data, folder)

	if err != nil {
		return "", err
	}

	return host + "/reports/" + fileId, nil
}
