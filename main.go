package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"todo-api/constants"
	"todo-api/database"
	"todo-api/models"
	"todo-api/rabbitmq"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type ApiResponse struct {
	StatusCode    int    `json:"status_code"`
	StatusMessage string `json:"status_message"`
}

func handleGetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		todoList, err := database.GetAllTodos()
		if err != nil {
			fmt.Println("Error fetching todos: ", err)
			json.NewEncoder(w).Encode(ApiResponse{
				StatusCode:    0,
				StatusMessage: constants.DbFetchError,
			})
			return
		}
		json.NewEncoder(w).Encode(todoList)
		return
	} else {
		id, err := strconv.Atoi(idStr)

		if err != nil {
			fmt.Println("Error converting id to integer: ", err)
			json.NewEncoder(w).Encode(ApiResponse{
				StatusCode:    0,
				StatusMessage: "Please provide a valid id",
			})
			return
		}

		todo, err := database.GetTodoById(id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No todo found with the given id")
				json.NewEncoder(w).Encode(ApiResponse{
					StatusCode:    0,
					StatusMessage: "No record found with the given id",
				})
				return
			}
			fmt.Println("Error fetching todo by id: ", err)
			json.NewEncoder(w).Encode(ApiResponse{
				StatusCode:    0,
				StatusMessage: constants.DbFetchError,
			})
			return

		}
		json.NewEncoder(w).Encode(todo)
	}
}

func handlePostTodos(w http.ResponseWriter, r *http.Request) {
	var newTodo models.Todo
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&newTodo)

	if err != nil {
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: "Error decoding request body",
		})
		return
	}

	if newTodo.Title == "" {
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: "Title is required",
		})
		return
	}
	if newTodo.Status == "" {
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: "Status is required",
		})
		return
	}
	err = database.InsertTodo(newTodo)

	if err != nil {
		fmt.Println("Error inserting todo: ", err)
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: "Error inserting todo ",
		})
		return
	}

	json.NewEncoder(w).Encode(ApiResponse{
		StatusCode:    1,
		StatusMessage: "Todo added successfully",
	})

}

func handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// mux.Vars(r) is a function provided by the Gorilla Mux router.\
	// It extracts path parameters from the URL.
	// These path parameters are defined in the route using curly braces: {}.
	// We are passing the request pointer to the function to extract the parameters.
	vars := mux.Vars(r)
	id, erro := strconv.Atoi(vars["id"])
	// So you can extract the value of the parameter used in the route using vars["parameter_name"].
	// First check and parse the id and check any errors
	if erro != nil {
		fmt.Println("Error converting id to integer: ", erro)
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: "Invalid ID format",
		})
		return
	}

	//Then parse the new todo from the request Body

	var updatedTodo models.Todo
	err := json.NewDecoder(r.Body).Decode(&updatedTodo)

	if err != nil {
		fmt.Println("Error decoding request body: ", err)
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: constants.DbUpdateError,
		})
		return
	}

	if updatedTodo.Title == "" || updatedTodo.Status == "" {

		var message string

		if updatedTodo.Title == "" {
			message = "Title is required"
		} else {
			message = "Status is required"
		}

		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: message,
		})
		return
	}

	//Then find the todo to update by matching the id
	updatedTodo.Id = id
	err = database.UpdateTodo(updatedTodo)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("No rows affected, todo with given id does not exist")
		} else {
			fmt.Println("Error updating todo ", err)
		}

		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: constants.DbUpdateError,
		})
		return
	}

	json.NewEncoder(w).Encode(ApiResponse{
		StatusCode:    1,
		StatusMessage: constants.DbUpdateSuccess,
	})

}

func handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Println("Error parsing your ID")
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: "Invalid ID format",
		})
		return
	}

	err = database.DeleteTodoById(id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("No todo found with the given id")
			w.WriteHeader(http.StatusBadRequest)
			//Send 400 Bad Request status code when no todo is found
			//This is a common practice to send a 400 Bad Request status code when the request is invalid
			//or the resource is not found
			json.NewEncoder(w).Encode(ApiResponse{
				StatusCode:    0,
				StatusMessage: "No record found with the given id",
			})
			return
		}
		fmt.Println("Error Deleting Rows: ", err)
		json.NewEncoder(w).Encode(ApiResponse{
			StatusCode:    0,
			StatusMessage: constants.DBDeleteError,
		})
		return

	}
	json.NewEncoder(w).Encode(ApiResponse{
		StatusCode:    1,
		StatusMessage: constants.DBDeleteSuccess,
	})

}

func main() {
	rabbitmq.ConnectRabbitMQ()

	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file: ", err)
		return
	}

	database.InitDB()

	

	router := mux.NewRouter()

	router.HandleFunc("/todo", handleGetTodos).Methods(http.MethodGet)
	router.HandleFunc("/todo/{id}", handleGetTodos).Methods(http.MethodGet)
	router.HandleFunc("/todo", handlePostTodos).Methods(http.MethodPost)
	router.HandleFunc("/todo/{id}", handleUpdateTodo).Methods(http.MethodPut)
	router.HandleFunc("/todo/{id}", handleDeleteTodo).Methods(http.MethodDelete)

	err = http.ListenAndServe(":8080", router)

	if err != nil {
		fmt.Println("Error starting server: ", err)
		return
	}

}
