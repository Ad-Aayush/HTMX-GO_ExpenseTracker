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
	"strings"
	"time"
)

func DaysInMonth(t time.Time) int {
	y, m, _ := t.Date()
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

type expense struct {
	date_str    string
	date        time.Time
	description string
	category    string
	amount      int
}

type send_data struct {
	array        []expense
	filter_array []expense
	db           *sql.DB
}

func (x send_data) total() int {
	str := "SELECT SUM(amount) FROM expense"
	fmt.Printf("%v\n", str)
	rows, err := x.db.Query(str)
	if err != nil {
		log.Fatal(err)
	}
	var ans int
	for rows.Next() {
		err := rows.Scan(&ans)
		if err != nil {
			ans = 0
		}

	}

	return ans
}

func (x send_data) total_curr_month() int {
	days := DaysInMonth(time.Now())
	date_str := time.Now().Format("2006-01-02")
	str := fmt.Sprintf("SELECT SUM(amount) FROM expense WHERE '%v'<=date+%d AND date<='%v'", date_str, days, date_str)
	fmt.Printf("%v\n", str)
	rows, err := x.db.Query(str)
	if err != nil {
		log.Fatal(err)
	}
	var ans int
	for rows.Next() {
		err := rows.Scan(&ans)
		if err != nil {
			ans = 0
		}

	}

	return ans
}

func sql_query(db *sql.DB, str string) []expense {
	rows, err := db.Query(str)
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
	return expenses
}

func sql_filter(db *sql.DB, filter, filterDate, search string) []expense{
	var query string
		days := DaysInMonth(time.Now())
		if filter == "All" {
			date_str := time.Now().Format("2006-01-02")
			// query = "SELECT * FROM expense ORDER BY date DESC"
			if filterDate == "All Dates" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", search)
			} else if filterDate == "Last 7 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+7 AND date<='%v'  AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, search)
			} else if filterDate == "Last 30 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+%d AND date<='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, days, date_str, search)
			} else if filterDate == "Last 3 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+90 AND date<='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, search)
			} else if filterDate == "Last 6 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+180 AND date<='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, search)
			} else if filterDate == "Last Year" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+365 AND date<='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, search)
			}

		} else {
			date_str := time.Now().Format("2006-01-02")
			// query = "SELECT * FROM expense ORDER BY date DESC"
			if filterDate == "All Dates" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE type='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", filter, search)
			} else if filterDate == "Last 7 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+7 AND date<='%v' AND type='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, filter, search)
			} else if filterDate == "Last 30 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+%d AND date<='%v' AND type='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, days, date_str, filter, search)
			} else if filterDate == "Last 3 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+90 AND date<='%v' AND type='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, filter, search)
			} else if filterDate == "Last 6 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+180 AND date<='%v' AND type='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, filter, search)
			} else if filterDate == "Last Year" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+365 AND date<='%v' AND type='%v' AND LOWER(description) LIKE '%%%v%%' ORDER BY date DESC", date_str, date_str, filter, search)
			}
		}
		fmt.Printf("ADD: %v\n", query)
		return sql_query(db, query)
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

	e := echo.New()
	var sent send_data

	sent.db = db
	e.GET("/", func(c echo.Context) error {
		sent.array = sql_query(db, "SELECT * FROM expense ORDER BY date DESC")
		component := index(sent)
		return component.Render(context.Background(), c.Response().Writer)
	})

	e.POST("/add", func(c echo.Context) error {
		date := c.FormValue("date")
		description := c.FormValue("description")
		category := c.FormValue("category")
		filter := c.FormValue("filter")
		filterDate := c.FormValue("filterDate")
		amount := c.FormValue("amount")
		search := c.FormValue("search")
		// fmt.Print("hi\n")
		fmt.Printf(filter, filterDate, category)
		fmt.Printf("\n%v\n", filter)
		// * Insert
		sqlQuery := fmt.Sprintf("INSERT INTO expense(date, description, type, amount) VALUES('%v', '%v', '%v', %v)", date, description, category, amount)
		db.Query(sqlQuery)
		sent.array = sql_query(db, "SELECT * FROM expense ORDER BY date DESC")
		// * Query
		sent.filter_array = sql_filter(db, filter, filterDate, search)
		// fmt.Printf("HELLO")
		return table_out(sent.filter_array).Render(context.Background(), c.Response().Writer)
	})

	e.GET("/update", func(c echo.Context) error {
		// Wait 20ms
		fmt.Printf("Request Recieved %d\n", sent.total())

		return totalEx(sent).Render(context.Background(), c.Response().Writer)
	})

	e.GET("/updateMonth", func(c echo.Context) error {
		// Wait 20ms
		fmt.Printf("Request Recieved %d\n", sent.total_curr_month())

		return totalMonthlyEx(sent).Render(context.Background(), c.Response().Writer)
	})

	e.GET("/filter", func(c echo.Context) error {
		fmt.Printf("Request Recieved Filter\n")

		filter := c.FormValue("filter")
		filterDate := c.FormValue("filterDate")
		search := c.FormValue("search")
		// fmt.Printf("%v\n", category)

		sent.filter_array = sql_filter(db, filter, filterDate, search)

		return table_out(sent.filter_array).Render(context.Background(), c.Response().Writer)
	})


	e.GET("/delete-row", func(c echo.Context) error {
		date := c.FormValue("date")
		description := c.FormValue("description")
		category := c.FormValue("category")

		amount := c.FormValue("amount")
		// * DELETE
		sqlQuery := fmt.Sprintf("DELETE FROM expense WHERE date='%v' AND description='%v' AND type='%v' AND amount=%v", date, description, category, amount)
		fmt.Printf("%v\n", sqlQuery)
		_, err := db.Query(sqlQuery)
		if err != nil {
			log.Fatal(err)
		}
		return oob(sent).Render(context.Background(), c.Response().Writer)
	})

	e.POST("/search", func(c echo.Context) error {
		search := c.FormValue("search")
		str := fmt.Sprintf("SELECT * FROM expense WHERE LOWER(description) LIKE '%%%v%%'", strings.ToLower(search))

		fmt.Printf("%v\n", str)
		sent.filter_array = sql_query(db, str)
		return table_out(sent.filter_array).Render(context.Background(), c.Response().Writer)
	})
	log.Fatal(e.Start(fmt.Sprintf(":%v", port)))
}
