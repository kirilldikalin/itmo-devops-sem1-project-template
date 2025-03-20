package storage

const (
	checkDuplicateQuery = `
		SELECT COUNT(*) 
		FROM prices 
		WHERE name = $1 
		  AND category = $2 
		  AND price = $3 
		  AND create_date = $4;
	`

	insertProductQuery = `
		INSERT INTO prices (id, name, category, price, create_date)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING;
	`

	calculateStatsQuery = `
		SELECT 
			COUNT(*) AS total_items,
			COUNT(DISTINCT category) AS total_categories,
			COALESCE(SUM(price), 0) AS total_price
		FROM prices;
	`

	getAllProductsQuery = "SELECT id, name, category, price, create_date FROM prices ORDER BY id"

	baseFilteredQuery = "SELECT id, name, category, price, create_date FROM prices"

	fullDayEndSuffix = " 23:59:59"
)
