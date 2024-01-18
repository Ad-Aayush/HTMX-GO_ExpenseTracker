package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

// type date_type struct {
// 	month time.Month
// 	year  int
// 	day   int
// 	str   string
// }

// func (x *date_type) convert() {
// 	x.str = fmt.Sprintf("%v-%d-%d", x.day, x.month, x.year)
// }

type expense struct {
	date_str    string
	date        time.Time
	description string
	category    string
	amount      int
}

type send_data struct {
	array []expense
}

func (x send_data) total() int {
	ans := 0
	for _, exp := range x.array {
		ans += exp.amount
	}
	return ans
}

func (x send_data) total_curr_month() int {
	ans := 0
	for _, exp := range x.array {
		if exp.date.Month() == time.Now().Month() {
			ans += exp.amount
		}
	}
	return ans
}

func main() {
	godotenv.Load(".env")
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	db_port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")
	db_driver := os.Getenv("DB_DRIVER")

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, db_port, user, password, db_name)
	db, err := sql.Open(db_driver, psqlconn)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM expense ORDER BY date DESC")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var expenses []expense

	for rows.Next() {
		var exp expense

		err = rows.Scan(&exp.date, &exp.description, &exp.category, &exp.amount)

		if err != nil {
			log.Fatal(err)
		}

		exp.date_str = exp.date.Format("02-01-2006")

		expenses = append(expenses, exp)
		// fmt.Print(exp.date.String())
	}

	e := echo.New()
	var sent send_data
	sent.array = expenses

	current_month := time.Now().Month()

	if current_month == 1 {
		fmt.Printf("%d\n", current_month)
	}

	e.GET("/", func(c echo.Context) error {
		component := index(sent)
		return component.Render(context.Background(), c.Response().Writer)
	})

	e.POST("/add", func(c echo.Context) error {
		date := c.FormValue("date")
		description := c.FormValue("description")
		category := c.FormValue("category")
		filter := c.FormValue("filter")
		amount := c.FormValue("amount")
		// fmt.Print("hi\n")
		// fmt.Printf(date, description, category)
		fmt.Printf("\n%v\n", filter)
		// * Insert
		sqlQuery := fmt.Sprintf("INSERT INTO expense(date, description, type, amount) VALUES('%v', '%v', '%v', %v)", date, description, category, amount)
		db.Query(sqlQuery)

		// * Query
		rows, err := db.Query("SELECT * FROM expense ORDER BY date DESC")
		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()

		var new_expenses []expense

		for rows.Next() {
			var exp expense

			err = rows.Scan(&exp.date, &exp.description, &exp.category, &exp.amount)

			if err != nil {
				log.Fatal(err)
			}

			exp.date_str = exp.date.Format("02-01-2006")

			new_expenses = append(new_expenses, exp)
			// fmt.Print(exp.date.String())
		}
		sent.array = new_expenses
		fmt.Printf("HELLO")
		return table_out(new_expenses).Render(context.Background(), c.Response().Writer)
	})

	e.GET("/update", func(c echo.Context) error {
		// Wait 20ms
		fmt.Printf("Request Recieved %d", sent.total())

		return totalEx(sent).Render(context.Background(), c.Response().Writer)
	})

	e.GET("/updateMonth", func(c echo.Context) error {
		// Wait 20ms
		fmt.Printf("Request Recieved %d", sent.total_curr_month())

		return totalMonthlyEx(sent).Render(context.Background(), c.Response().Writer)
	})

	log.Fatal(e.Start(fmt.Sprintf(":%v", port)))
}
