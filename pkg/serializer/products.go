package serializer

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"project_sem/platform/config"
	"project_sem/platform/storage"
)

func DeserializeProducts(input io.Reader) ([]storage.Product, int, int, error) {
	reader := csv.NewReader(input)

	header, err := reader.Read()
	if err == io.EOF {
		return nil, 0, 0, nil
	}
	if err != nil {
		return nil, 0, 0, err
	}

	const expectedCSVColumns = 5

	if len(header) < expectedCSVColumns {
		return nil, 0, 0, fmt.Errorf("invalid header format")
	}

	var products []storage.Product
	totalCount := 0
	duplicatesCount := 0
	seenIDs := make(map[int]bool)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		totalCount++

		if len(record) < expectedCSVColumns {
			continue
		}

		idStr, name, category, priceStr, dateStr := record[0], record[1], record[2], record[3], record[4]

		id, errID := strconv.Atoi(strings.TrimSpace(idStr))
		price, errPrice := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
		date, errDate := time.Parse(config.DefaultDateFormat, strings.TrimSpace(dateStr))

		if errID != nil || errPrice != nil || errDate != nil || name == "" || category == "" {
			continue
		}

		if seenIDs[id] {
			duplicatesCount++
			continue
		}
		seenIDs[id] = true

		products = append(products, storage.Product{
			ID:         id,
			Name:       name,
			Category:   category,
			Price:      price,
			CreateDate: date,
		})
	}

	return products, totalCount, duplicatesCount, nil
}

func SerializeProducts(products []storage.Product) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	defer writer.Flush()

	writer.Write([]string{"id", "name", "category", "price", "create_date"})

	for _, p := range products {
		record := []string{
			strconv.Itoa(p.ID),
			p.Name,
			p.Category,
			fmt.Sprintf("%.2f", p.Price),
			p.CreateDate.Format(config.DefaultDateFormat),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	return &buffer, nil
}
