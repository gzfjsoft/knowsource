package knowsource

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"

	"github.com/zeromicro/go-zero/core/logx"
)

type ColumnInfo struct {
	ColumnName string
	DataType   string
	IsNullable string
	MaxLength  int
	Precision  int
	Scale      int
}

type Config struct {
	MSSQL struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
		Encrypt  bool   `yaml:"encrypt"`
	} `yaml:"mssql"`
	MySQL struct {
		Host      string `yaml:"host"`
		Port      string `yaml:"port"`
		User      string `yaml:"user"`
		Password  string `yaml:"password"`
		Database  string `yaml:"database"`
		Charset   string `yaml:"charset"`
		ParseTime bool   `yaml:"parseTime"`
		Loc       string `yaml:"loc"`
	} `yaml:"mysql"`
	Views                 []string            `yaml:"views"`
	TableNames            map[string]string   `yaml:"table_names"` // Maps view name to MySQL table name
	PrimaryKeys           map[string]string   `yaml:"primary_keys"`
	AllowedColumns        map[string][]string `yaml:"allowed_columns"`
	DropTableBeforeImport bool                `yaml:"drop_table_before_import"`
}

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

// getTableName returns the MySQL table name for a view.
// If a mapping is configured, use it; otherwise use the view name.
func getTableName(viewName string, config *Config) string {
	if config.TableNames != nil {
		if tableName, ok := config.TableNames[viewName]; ok {
			return tableName
		}
	}
	return viewName
}

func ImportHrUserDept() error {
	// 跟踪是否有错误发生
	hasError := false

	// Load configuration from YAML file
	// If dev.ini exists, use syncHR.dev.yaml, otherwise use syncHR.yaml
	configPath := "syncHR.yaml"

	// Check if dev.ini exists
	if _, err := os.Stat("dev.ini"); err == nil {
		configPath = "syncHR.dev.yaml"
	}

	logx.Infof("Loading config from: %s", configPath)

	config, err := loadConfig(configPath)
	if err != nil {
		logx.Infof("Error loading config: %v", err)
		return fmt.Errorf("error loading config: %v", err)
	}

	// Build SQL Server connection string
	encryptOption := "encrypt=disable"
	if config.MSSQL.Encrypt {
		encryptOption = "encrypt=true"
	}
	connString := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;%s",
		config.MSSQL.Server, config.MSSQL.Port, config.MSSQL.User, config.MSSQL.Password,
		config.MSSQL.Database, encryptOption)

	// Connect to SQL Server
	mssqlDB, err := sql.Open("sqlserver", connString)
	if err != nil {
		logx.Infof("Error connecting to SQL Server: %v", err)
		return fmt.Errorf("error connecting to SQL Server: %v", err)
	}
	defer mssqlDB.Close()

	// Test SQL Server connection
	err = mssqlDB.Ping()
	if err != nil {
		logx.Infof("Error pinging SQL Server: %v", err)
		return fmt.Errorf("error pinging SQL Server: %v", err)
	}

	fmt.Println("Successfully connected to SQL Server!")

	// Build MySQL connection string
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		config.MySQL.User, config.MySQL.Password, config.MySQL.Host, config.MySQL.Port,
		config.MySQL.Database, config.MySQL.Charset, config.MySQL.ParseTime,
		strings.ReplaceAll(config.MySQL.Loc, "/", "%2F"))

	// Connect to MySQL
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		logx.Infof("Error connecting to MySQL: %v", err)
		return fmt.Errorf("error connecting to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	// Test MySQL connection
	err = mysqlDB.Ping()
	if err != nil {
		logx.Infof("Error pinging MySQL: %v", err)
		return fmt.Errorf("error pinging MySQL: %v", err)
	}

	fmt.Println("Successfully connected to MySQL!")

	// Views to export from config
	views := config.Views

	info := ""

	for _, viewName := range views {
		fmt.Printf("\nProcessing view: %s\n", viewName)

		// Get MySQL table name (may be different from view name)
		tableName := getTableName(viewName, config)
		if tableName != viewName {
			fmt.Printf("Mapping view %s to table %s\n", viewName, tableName)
		}

		// Get view structure
		columns, err := getViewStructure(mssqlDB, viewName)
		if err != nil {
			logx.Infof("Error getting structure for view %s: %v", viewName, err)
			hasError = true
			continue
		}

		// Get primary key for this table (use viewName for config lookup)
		primaryKey := ""
		if config.PrimaryKeys != nil {
			primaryKey = config.PrimaryKeys[viewName]
		}
		if primaryKey != "" {
			fmt.Printf("Primary key for table %s: %s\n", tableName, primaryKey)
		} else {
			fmt.Printf("No primary key configured for table %s\n", tableName)
		}

		// Create table in MySQL (use tableName for actual table creation)
		err = createMySQLTable(mysqlDB, tableName, columns, primaryKey, config, viewName)
		if err != nil {
			logx.Errorf("Error creating table for view %s: %v", viewName, err)
			info += err.Error() + "\n"
			hasError = true
			continue
		}

		// Export data
		err = exportViewDataToMySQL(mssqlDB, mysqlDB, viewName, tableName, columns, config)
		if err != nil {
			logx.Errorf("Error exporting data for view %s: %v", viewName, err)
			info += err.Error() + "\n"
			hasError = true
			continue
		}

		fmt.Printf("View %s exported successfully to MySQL table %s!\n", viewName, tableName)
	}

	if hasError {

		return errors.New(info)
	} else {

		logx.Infof("All views exported to MySQL database!")
		return nil
	}
}

func getViewStructure(db *sql.DB, viewName string) ([]ColumnInfo, error) {
	query := `
        SELECT 
            COLUMN_NAME,
            DATA_TYPE,
            IS_NULLABLE,
            CHARACTER_MAXIMUM_LENGTH,
            NUMERIC_PRECISION,
            NUMERIC_SCALE
        FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_NAME = @p1
        ORDER BY ORDINAL_POSITION
    `

	rows, err := db.Query(query, viewName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var maxLength, precision, scale sql.NullInt32

		err := rows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable,
			&maxLength, &precision, &scale)
		if err != nil {
			return nil, err
		}

		if maxLength.Valid {
			col.MaxLength = int(maxLength.Int32)
		}
		if precision.Valid {
			col.Precision = int(precision.Int32)
		}
		if scale.Valid {
			col.Scale = int(scale.Int32)
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func createSQLiteTable(db *sql.DB, tableName string, columns []ColumnInfo) error {
	var columnDefs []string

	for _, col := range columns {
		sqlType := getSQLiteType(col.DataType, col.MaxLength, col.Precision, col.Scale)
		nullable := ""
		if col.IsNullable == "NO" {
			nullable = " NOT NULL"
		}
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s%s", col.ColumnName, sqlType, nullable))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columnDefs, ", "))

	_, err := db.Exec(createSQL)
	return err
}

func getSQLiteType(mssqlType string, maxLength, precision, scale int) string {
	switch strings.ToUpper(mssqlType) {
	case "INT", "SMALLINT", "TINYINT", "BIGINT":
		return "INTEGER"
	case "DECIMAL", "NUMERIC", "MONEY", "SMALLMONEY":
		return "REAL"
	case "FLOAT", "REAL":
		return "REAL"
	case "DATE", "TIME", "DATETIME", "DATETIME2", "SMALLDATETIME":
		return "TEXT"
	case "BIT":
		return "INTEGER"
	case "CHAR", "VARCHAR", "TEXT", "NCHAR", "NVARCHAR", "NTEXT":
		return "TEXT"
	case "BINARY", "VARBINARY", "IMAGE":
		return "BLOB"
	default:
		return "TEXT"
	}
}

func exportViewData(mssqlDB, sqliteDB *sql.DB, viewName string, columns []ColumnInfo) error {
	// Get column names for SELECT and INSERT
	var colNames []string
	var placeholders []string

	for _, col := range columns {
		colNames = append(colNames, col.ColumnName)
		placeholders = append(placeholders, "?")
	}

	selectSQL := fmt.Sprintf("SELECT %s FROM %s", strings.Join(colNames, ", "), viewName)
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		viewName, strings.Join(colNames, ", "), strings.Join(placeholders, ", "))

	// Read data from SQL Server
	rows, err := mssqlDB.Query(selectSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Prepare SQLite insert statement
	stmt, err := sqliteDB.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Export data
	var count int
	for rows.Next() {
		// Create slice of interfaces for scanning
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}

		// Insert into SQLite
		_, err = stmt.Exec(values...)
		if err != nil {
			return err
		}

		count++
	}

	fmt.Printf("Exported %d rows from view %s\n", count, viewName)
	return nil
}

// getAllowedColumns returns the list of allowed column names for a table from config
// Returns both a case-insensitive map and the original column names map
// viewName is used to look up the config (since config uses view names as keys)
func getAllowedColumns(viewName string, config *Config) (map[string]bool, map[string]string) {
	allowed := make(map[string]bool)
	allowedLower := make(map[string]string) // Maps lowercase to original case

	if config.AllowedColumns != nil {
		if columns, ok := config.AllowedColumns[viewName]; ok {
			for _, col := range columns {
				allowed[col] = true
				allowedLower[strings.ToLower(col)] = col
			}
		}
	}

	return allowed, allowedLower
}

func createMySQLTable(db *sql.DB, tableName string, columns []ColumnInfo, primaryKey string, config *Config, viewName string) error {
	// Drop table if configured
	if config.DropTableBeforeImport {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		_, err := db.Exec(dropSQL)
		if err != nil {
			logx.Infof("Warning: Failed to drop table %s: %v", tableName, err)
		} else {
			fmt.Printf("Dropped table %s (if existed)\n", tableName)
		}
	}

	var columnDefs []string
	var primaryKeyCol *ColumnInfo

	// Get allowed columns for this table (use viewName for config lookup)
	allowedColumns, allowedLower := getAllowedColumns(viewName, config)

	// Filter columns to only include allowed ones (case-insensitive matching)
	var filteredColumns []ColumnInfo
	var foundColumnNames []string
	for _, col := range columns {
		colLower := strings.ToLower(col.ColumnName)
		if allowedColumns[col.ColumnName] || allowedLower[colLower] != "" {
			filteredColumns = append(filteredColumns, col)
			foundColumnNames = append(foundColumnNames, col.ColumnName)
		}
	}

	if len(filteredColumns) == 0 {
		// Build error message with details
		var actualColumns []string
		for _, col := range columns {
			actualColumns = append(actualColumns, col.ColumnName)
		}
		var expectedColumns []string
		if config.AllowedColumns != nil {
			if cols, ok := config.AllowedColumns[viewName]; ok {
				expectedColumns = cols
			}
		}
		return fmt.Errorf("no allowed columns found for table %s. Found columns: %v, Expected columns: %v",
			tableName, actualColumns, expectedColumns)
	}

	// Find primary key column if specified (case-insensitive matching)
	if primaryKey != "" {
		primaryKeyLower := strings.ToLower(primaryKey)
		for i := range filteredColumns {
			if strings.ToLower(filteredColumns[i].ColumnName) == primaryKeyLower {
				primaryKeyCol = &filteredColumns[i]
				break
			}
		}
		if primaryKeyCol == nil {
			logx.Infof("Warning: Primary key column '%s' not found in table %s, skipping primary key constraint", primaryKey, tableName)
		} else {
			fmt.Printf("Primary key column '%s' found, will be set as PRIMARY KEY\n", primaryKeyCol.ColumnName)
		}
	}

	for _, col := range filteredColumns {
		mysqlType := getMySQLType(col.DataType, col.MaxLength, col.Precision, col.Scale)

		// Special handling for FAge and Fdept_id - convert to INT
		if col.ColumnName == "FAge" || col.ColumnName == "Fdept_id" {
			mysqlType = "INT"
		}

		nullable := ""
		// Primary key columns should be NOT NULL
		// String columns should be NOT NULL
		isStringType := isStringType(col.DataType)
		if col.IsNullable == "NO" || (primaryKeyCol != nil && col.ColumnName == primaryKeyCol.ColumnName) || isStringType {
			nullable = " NOT NULL"
		}
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s%s", col.ColumnName, mysqlType, nullable))
	}

	// Add PRIMARY KEY constraint if specified
	if primaryKeyCol != nil {
		// Use the actual column name from the filtered columns (not the config key)
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", primaryKeyCol.ColumnName))
		fmt.Printf("Added PRIMARY KEY constraint on column '%s'\n", primaryKeyCol.ColumnName)
	}

	// Use IF NOT EXISTS only if we're not dropping the table first
	createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(columnDefs, ", "))
	if !config.DropTableBeforeImport {
		createSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columnDefs, ", "))
	}

	fmt.Printf("Creating table %s with SQL: %s\n", tableName, createSQL)
	_, err := db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}
	fmt.Printf("Table %s created successfully\n", tableName)
	return nil
}

func isStringType(mssqlType string) bool {
	switch strings.ToUpper(mssqlType) {
	case "CHAR", "NCHAR", "VARCHAR", "NVARCHAR", "TEXT", "NTEXT":
		return true
	default:
		return false
	}
}

func getMySQLType(mssqlType string, maxLength, precision, scale int) string {
	switch strings.ToUpper(mssqlType) {
	case "INT", "SMALLINT", "TINYINT":
		return "INT"
	case "BIGINT":
		return "BIGINT"
	case "DECIMAL", "NUMERIC":
		if precision > 0 {
			if scale > 0 {
				return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
			}
			return fmt.Sprintf("DECIMAL(%d)", precision)
		}
		return "DECIMAL"
	case "MONEY", "SMALLMONEY":
		return "DECIMAL(19,4)"
	case "FLOAT", "REAL":
		return "FLOAT"
	case "DATE":
		return "DATE"
	case "TIME":
		return "TIME"
	case "DATETIME", "DATETIME2", "SMALLDATETIME":
		return "DATETIME"
	case "BIT":
		return "TINYINT(1)"
	case "CHAR", "NCHAR":
		if maxLength > 0 {
			return fmt.Sprintf("CHAR(%d)", maxLength)
		}
		return "CHAR(255)"
	case "VARCHAR", "NVARCHAR":
		if maxLength > 0 {
			return fmt.Sprintf("VARCHAR(%d)", maxLength)
		}
		return "VARCHAR(255)"
	case "TEXT", "NTEXT":
		return "TEXT"
	case "BINARY":
		if maxLength > 0 {
			return fmt.Sprintf("BINARY(%d)", maxLength)
		}
		return "BINARY(255)"
	case "VARBINARY":
		if maxLength > 0 {
			return fmt.Sprintf("VARBINARY(%d)", maxLength)
		}
		return "VARBINARY(255)"
	case "IMAGE":
		return "LONGBLOB"
	default:
		return "VARCHAR(255)"
	}
}

func exportViewDataToMySQL(mssqlDB, mysqlDB *sql.DB, viewName string, tableName string, columns []ColumnInfo, config *Config) error {
	// Clear existing data from table if it exists
	truncateSQL := fmt.Sprintf("TRUNCATE TABLE %s", tableName)
	_, err := mysqlDB.Exec(truncateSQL)
	if err != nil {
		// If table doesn't exist or TRUNCATE fails, try DELETE instead
		deleteSQL := fmt.Sprintf("DELETE FROM %s", tableName)
		_, err = mysqlDB.Exec(deleteSQL)
		if err != nil {
			// Ignore error if table doesn't exist (will be created later)
			logx.Infof("Warning: Could not clear table %s (may not exist yet): %v", tableName, err)
		}
	}

	// Get allowed columns for this table (use viewName for config lookup)
	allowedColumns, allowedLower := getAllowedColumns(viewName, config)

	// Filter columns to only include allowed ones (case-insensitive matching)
	var filteredColumns []ColumnInfo
	for _, col := range columns {
		colLower := strings.ToLower(col.ColumnName)
		if allowedColumns[col.ColumnName] || allowedLower[colLower] != "" {
			filteredColumns = append(filteredColumns, col)
		}
	}

	if len(filteredColumns) == 0 {
		return fmt.Errorf("no allowed columns found for table %s", viewName)
	}

	// Get column names for SELECT and INSERT
	var colNames []string
	var placeholders []string

	for _, col := range filteredColumns {
		colNames = append(colNames, col.ColumnName)
		placeholders = append(placeholders, "?")
	}

	selectSQL := fmt.Sprintf("SELECT %s FROM %s", strings.Join(colNames, ", "), viewName)
	// 只导入 status=0 的人员，status!=0 的不导入
	if viewName == "frEmpAI" {
		selectSQL += " WHERE [status] = 0"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(colNames, ", "), strings.Join(placeholders, ", "))

	// Read data from SQL Server
	rows, err := mssqlDB.Query(selectSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Prepare MySQL insert statement
	stmt, err := mysqlDB.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Export data
	var count int
	for rows.Next() {
		// Create slice of interfaces for scanning
		// For numeric types that might be DECIMAL/NUMERIC, use sql.NullString to capture as string first
		scanValues := make([]interface{}, len(filteredColumns))
		for i, col := range filteredColumns {
			// For FAge and Fdept_id, use sql.NullString to handle DECIMAL/NUMERIC types
			if col.ColumnName == "FAge" || col.ColumnName == "Fdept_id" {
				scanValues[i] = new(sql.NullString)
			} else {
				scanValues[i] = new(interface{})
			}
		}

		err := rows.Scan(scanValues...)
		if err != nil {
			return err
		}

		// Convert scan values to actual values
		values := make([]interface{}, len(filteredColumns))
		for i, col := range filteredColumns {
			if col.ColumnName == "FAge" || col.ColumnName == "Fdept_id" {
				ns := scanValues[i].(*sql.NullString)
				if ns.Valid {
					values[i] = ns.String
				} else {
					values[i] = nil
				}
			} else {
				valPtr := scanValues[i].(*interface{})
				values[i] = *valPtr
			}
		}

		// Process values: convert types and handle NULLs
		for i, col := range filteredColumns {
			// Convert FAge from various numeric types to int64
			if col.ColumnName == "FAge" {
				if values[i] == nil {
					values[i] = nil
				} else {
					var intVal int64
					switch v := values[i].(type) {
					case float64:
						intVal = int64(v)
					case float32:
						intVal = int64(v)
					case int64:
						intVal = v
					case int32:
						intVal = int64(v)
					case int:
						intVal = int64(v)
					case *float64:
						if v != nil {
							intVal = int64(*v)
						} else {
							values[i] = nil
							continue
						}
					case *float32:
						if v != nil {
							intVal = int64(*v)
						} else {
							values[i] = nil
							continue
						}
					case *int64:
						if v != nil {
							intVal = *v
						} else {
							values[i] = nil
							continue
						}
					case *int32:
						if v != nil {
							intVal = int64(*v)
						} else {
							values[i] = nil
							continue
						}
					case []byte:
						// MSSQL DECIMAL/NUMERIC types might come as []byte
						// Try to parse as string then convert
						if len(v) > 0 {
							var f float64
							_, err := fmt.Sscanf(string(v), "%f", &f)
							if err == nil {
								intVal = int64(f)
							} else {
								// Try as int
								var i int64
								_, err := fmt.Sscanf(string(v), "%d", &i)
								if err == nil {
									intVal = i
								} else {
									logx.Infof("Warning: Could not parse FAge value: %v, setting to nil", string(v))
									values[i] = nil
									continue
								}
							}
						} else {
							values[i] = nil
							continue
						}
					case string:
						// Try to parse string as number
						var f float64
						_, err := fmt.Sscanf(v, "%f", &f)
						if err == nil {
							intVal = int64(f)
						} else {
							// Try as int
							var i int64
							_, err := fmt.Sscanf(v, "%d", &i)
							if err == nil {
								intVal = i
							} else {
								logx.Infof("Warning: Could not parse FAge value: %v, setting to nil", v)
								values[i] = nil
								continue
							}
						}
					default:
						// Try to convert using fmt.Sprintf and then parse
						strVal := fmt.Sprintf("%v", v)
						var f float64
						_, err := fmt.Sscanf(strVal, "%f", &f)
						if err == nil {
							intVal = int64(f)
						} else {
							logx.Infof("Warning: FAge type %T value %v could not be converted, setting to nil", v, v)
							values[i] = nil
							continue
						}
					}
					values[i] = intVal
				}
			}

			// Convert Fdept_id to int64 (handle various numeric types and string from DECIMAL/NUMERIC)
			if col.ColumnName == "Fdept_id" {
				if values[i] == nil {
					values[i] = nil
				} else {
					var intVal int64
					switch v := values[i].(type) {
					case int64:
						intVal = v
					case int32:
						intVal = int64(v)
					case int:
						intVal = int64(v)
					case float64:
						intVal = int64(v)
					case float32:
						intVal = int64(v)
					case *int64:
						if v != nil {
							intVal = *v
						} else {
							values[i] = nil
							continue
						}
					case *int32:
						if v != nil {
							intVal = int64(*v)
						} else {
							values[i] = nil
							continue
						}
					case *float64:
						if v != nil {
							intVal = int64(*v)
						} else {
							values[i] = nil
							continue
						}
					case *float32:
						if v != nil {
							intVal = int64(*v)
						} else {
							values[i] = nil
							continue
						}
					case sql.NullInt64:
						if v.Valid {
							intVal = v.Int64
						} else {
							values[i] = nil
							continue
						}
					case string:
						// Parse string as int64 (from DECIMAL/NUMERIC types scanned as string)
						var i int64
						_, err := fmt.Sscanf(v, "%d", &i)
						if err == nil {
							intVal = i
						} else {
							// Try parsing as float first, then convert to int
							var f float64
							_, err := fmt.Sscanf(v, "%f", &f)
							if err == nil {
								intVal = int64(f)
							} else {
								logx.Infof("Warning: Could not parse Fdept_id value: %v, setting to nil", v)
								values[i] = nil
								continue
							}
						}
					case []byte:
						// MSSQL DECIMAL/NUMERIC types might come as []byte
						if len(v) > 0 {
							var i int64
							_, err := fmt.Sscanf(string(v), "%d", &i)
							if err == nil {
								intVal = i
							} else {
								// Try as float
								var f float64
								_, err := fmt.Sscanf(string(v), "%f", &f)
								if err == nil {
									intVal = int64(f)
								} else {
									logx.Infof("Warning: Could not parse Fdept_id value: %v, setting to nil", string(v))
									values[i] = nil
									continue
								}
							}
						} else {
							values[i] = nil
							continue
						}
					default:
						// Try to convert using fmt.Sprintf and then parse
						strVal := fmt.Sprintf("%v", v)
						var i int64
						_, err := fmt.Sscanf(strVal, "%d", &i)
						if err == nil {
							intVal = i
						} else {
							// Try as float
							var f float64
							_, err := fmt.Sscanf(strVal, "%f", &f)
							if err == nil {
								intVal = int64(f)
							} else {
								logx.Infof("Warning: Fdept_id type %T value %v could not be converted, setting to nil", v, v)
								values[i] = nil
								continue
							}
						}
					}
					values[i] = intVal
				}
			}

			// Convert NULL string values to empty strings
			if isStringType(col.DataType) {
				if values[i] == nil {
					values[i] = ""
				} else {
					// Convert to string if not already
					switch v := values[i].(type) {
					case string:
						values[i] = v
					case []byte:
						values[i] = string(v)
					default:
						// Convert other types to string
						values[i] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		// Insert into MySQL
		_, err = stmt.Exec(values...)
		if err != nil {
			return err
		}

		count++
	}

	fmt.Printf("Exported %d rows from view %s to MySQL table %s\n", count, viewName, tableName)
	return nil
}
