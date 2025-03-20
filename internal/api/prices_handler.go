package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"project_sem/pkg/archiver"
	"project_sem/pkg/serializer"
	"project_sem/platform/storage"
)

type PriceStatistics struct {
	TotalCount      int     `json:"total_count"`
	DuplicatesCount int     `json:"duplicates_count"`
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}

func determineArchiver(r *http.Request) (interface {
	Extract(input io.Reader) (io.ReadCloser, error)
	Archive(output io.Writer, fileName string, data []byte) error
}, error) {
	q := r.URL.Query()
	archiveType := q.Get("type")
	if archiveType == "tar" {
		return archiver.NewTarArchiver(), nil
	}
	return archiver.NewZipArchiver(), nil
}

func PostPrices(repo storage.Repository, maxFileSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
		defer cancel()
		r = r.WithContext(ctx)

		if r.ContentLength > maxFileSize && maxFileSize > 0 {
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			log.Printf("failed to read file: %v\n", err)
			http.Error(w, "failed to process file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		limitedReader := io.LimitReader(file, maxFileSize+1)

		arch, err := determineArchiver(r)
		if err != nil {
			log.Printf("failed to determine archiver: %v\n", err)
			http.Error(w, "failed to determine archive type", http.StatusBadRequest)
			return
		}

		csvFile, err := arch.Extract(limitedReader)
		if err != nil {
			log.Printf("failed to extract archive: %v\n", err)
			http.Error(w, "failed to extract archive", http.StatusBadRequest)
			return
		}
		defer csvFile.Close()

		products, totalCount, _, err := serializer.DeserializeProducts(csvFile)
		if err != nil {
			log.Printf("failed to deserialize products: %v\n", err)
			http.Error(w, "failed to parse CSV", http.StatusBadRequest)
			return
		}

		duplicatesCount, totalItems, totalCategories, totalPrice, err := repo.InsertProductsAndStats(products)
		if err != nil {
			log.Printf("failed to insert products: %v\n", err)
			http.Error(w, "failed to insert products", http.StatusInternalServerError)
			return
		}

		stats := PriceStatistics{
			TotalCount:      totalCount,
			DuplicatesCount: duplicatesCount,
			TotalItems:      totalItems,
			TotalCategories: totalCategories,
			TotalPrice:      totalPrice,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			log.Printf("failed to encode JSON: %v\n", err)
		}
	}
}

func GetPrices(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		start := q.Get("start")
		end := q.Get("end")
		minPrice := q.Get("min")
		maxPrice := q.Get("max")

		var products []storage.Product
		var err error

		if start != "" && end != "" && minPrice != "" && maxPrice != "" {
			products, err = repo.GetAllProductsFiltered(start, end, minPrice, maxPrice)
		} else {
			products, err = repo.GetAllProducts()
		}
		if err != nil {
			log.Printf("failed to get products: %v\n", err)
			http.Error(w, "failed to get products", http.StatusInternalServerError)
			return
		}

		buffer, err := serializer.SerializeProducts(products)
		if err != nil {
			log.Printf("failed to serialize products: %v\n", err)
			http.Error(w, "failed to serialize products", http.StatusInternalServerError)
			return
		}

		zipArchiver := archiver.NewZipArchiver()
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="data.zip"`)

		if err := zipArchiver.Archive(w, "data.csv", buffer.Bytes()); err != nil {
			log.Printf("failed to create zip archive: %v\n", err)
			http.Error(w, "failed to create zip archive", http.StatusInternalServerError)
			return
		}
	}
}