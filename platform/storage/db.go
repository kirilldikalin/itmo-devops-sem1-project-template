package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	"project_sem/platform/config"
)

type Repository interface {
	InsertProductsAndStats(products []Product) (int, int, int, float64, error)
	GetAllProductsFiltered(start, end string, min, max string) ([]Product, error)
	GetAllProducts() ([]Product, error)
	Close() error
}

type repository struct {
	db *sql.DB
}

func NewRepository(cfg config.DatabaseSettings) (Repository, error) {
	log.Println("Connecting to the database...")
	const sslModeDisable = "disable"

	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		sslModeDisable,
	)

	database, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err = database.Ping(); err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to database '%s'\n", cfg.Name)
	return &repository{db: database}, nil
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) InsertProductsAndStats(products []Product) (int, int, int, float64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	inserted := 0
	duplicates := 0

	for _, product := range products {
		var duplicateCount int
		err = tx.QueryRow(checkDuplicateQuery, product.Name, product.Category, product.Price, product.CreateDate).Scan(&duplicateCount)
		if err != nil {
			return 0, 0, 0, 0, err
		}

		if duplicateCount > 0 {
			duplicates++
			continue
		}

		_, err = tx.Exec(insertProductQuery, product.ID, product.Name, product.Category, product.Price, product.CreateDate)
		if err != nil {
			return 0, 0, 0, 0, err
		}

		inserted++
	}

	var totalItems, totalCategories int
	var totalPrice float64

	err = tx.QueryRow(calculateStatsQuery).Scan(&totalItems, &totalCategories, &totalPrice)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return totalItems, duplicates, totalCategories, totalPrice, nil
}

func (r *repository) GetAllProductsFiltered(start, end, min, max string) ([]Product, error) {
	var conditions []string
	var args []interface{}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(baseFilteredQuery)

	argIndex := 1

	if start != "" && end != "" {
		conditions = append(conditions, fmt.Sprintf("create_date BETWEEN $%d AND $%d", argIndex, argIndex+1))
		args = append(args, start, end+fullDayEndSuffix)
		argIndex += 2
	}
	if min != "" && max != "" {
		conditions = append(conditions, fmt.Sprintf("price BETWEEN $%d AND $%d", argIndex, argIndex+1))
		args = append(args, min, max)
		argIndex += 2
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		for i, cond := range conditions {
			if i > 0 {
				queryBuilder.WriteString(" AND ")
			}
			queryBuilder.WriteString(cond)
		}
	}

	queryBuilder.WriteString(" ORDER BY id")

	rows, err := r.db.Query(queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0, 10)

	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.CreateDate)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *repository) GetAllProducts() ([]Product, error) {
	rows, err := r.db.Query(getAllProductsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0, 10)

	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.CreateDate)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}