package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"todo-api/models"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitDB connects to the database
func InitDB() {
	// Step 1: Create connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	// Step 2: Open connection
	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error while opening DB connection: ", err)
	}

	// Step 3: Ping to verify connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Cannot connect to DB: ", err)
	}
	fmt.Println("✅ Connected to the database")

	err = MigrateDB()
	if err != nil {
		fmt.Println("Error migrating database: ", err)
		return
	}
}

func MigrateDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS todo (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		status BOOLEAN DEFAULT false
	);`
	_, err := DB.Exec(query)
	return err
}

func InsertTodo(todo models.Todo) error {
	query := `INSERT INTO todo (name, status) VALUES ($1, $2);`
	// PostgreSQL uses numbered placeholders for query parameters.
	//It helps avoid SQL injection and ensures values are escaped properly.
	_, err := DB.Exec(query, todo.Title, todo.Status)
	return err
}

func UpdateTodo(todo models.Todo) error {
	query := `Update todo SET name=$1, status=$2 WHERE id=$3;`
	res, err := DB.Exec(query, todo.Title, todo.Status, todo.Id)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Printf("Updated %d rows for id %d\n", rowsAffected, todo.Id)

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetAllTodos() ([]models.Todo, error) {
	query := `SELECT id,name,status FROM todo;`

	//We avoid using * in queries for better performance as we have to only fetch the data we are handling here which eliminated added columns

	rows, err := DB.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//What this does is create a empty slice instead of a nil slice so returning in response is easier
	todos := make([]models.Todo, 0)

	//Here rows.Next() iterates over the result set, moving to the next row each time it is called.
	//If there are no more rows, it returns false and the loop ends.
	//Also in rows.Scan(), we can use only 3 fields as we are only fetching 3 columns from the database.
	//We use &todo.Id, &todo.Title, &todo.Status to store the values
	//Use the same sequence while storing the values as in the query

	for rows.Next() {
		var todo models.Todo
		//Scan the row into the todo struct
		//This stores the values of the row into the todo struct so we are using & so that we are modifying the actual struct
		err := rows.Scan(&todo.Id, &todo.Title, &todo.Status)
		if err != nil {
			return nil, err
		}
		//Add the scanned todo to the todos slice

		todos = append(todos, todo)
	}
	return todos, nil
}

func GetTodoById(id int) (models.Todo, error) {
	var todo models.Todo
	query := `Select id,name,status from todo where id=$1;`

	//If you are fetching a single row, you can use QueryRow instead of Query
	err := DB.QueryRow(query, id).Scan(&todo.Id, &todo.Title, &todo.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return todo, fmt.Errorf("no todo found with id %d", id)
		}
		return todo, err
	}

	return todo, nil

}

func DeleteTodoById(id int) error {
	query := `DELETE FROM todo WHERE id=$1;`

	res, err := DB.Exec(query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	return nil
}
