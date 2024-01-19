package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
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
		// fmt.Print("hi\n")
		fmt.Printf(filter, filterDate, category)
		fmt.Printf("\n%v\n", filter)
		// * Insert
		sqlQuery := fmt.Sprintf("INSERT INTO expense(date, description, type, amount) VALUES('%v', '%v', '%v', %v)", date, description, category, amount)
		db.Query(sqlQuery)
		sent.array = sql_query(db, "SELECT * FROM expense ORDER BY date DESC")
		// * Query
		var query string
		days := DaysInMonth(time.Now())
		if filter == "All" {
			date_str := time.Now().Format("2006-01-02")
			// query = "SELECT * FROM expense ORDER BY date DESC"
			if filterDate == "All Dates" {
				query = "SELECT * FROM expense ORDER BY date DESC"
			} else if filterDate == "Last 7 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+7 AND date<='%v' ORDER BY date DESC", date_str, date_str)
			} else if filterDate == "Last 30 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+%d AND date<='%v' ORDER BY date DESC", date_str, days, date_str)
			} else if filterDate == "Last 3 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+90 AND date<='%v' ORDER BY date DESC", date_str, date_str)
			} else if filterDate == "Last 6 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+180 AND date<='%v' ORDER BY date DESC", date_str, date_str)
			} else if filterDate == "Last Year" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+365 AND date<='%v' ORDER BY date DESC", date_str, date_str)
			}

		} else {
			date_str := time.Now().Format("2006-01-02")
			// query = "SELECT * FROM expense ORDER BY date DESC"
			if filterDate == "All Dates" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE AND type='%v' ORDER BY date DESC", filter)
			} else if filterDate == "Last 7 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+7 AND date<='%v' AND type='%v' ORDER BY date DESC", date_str, date_str, filter)
			} else if filterDate == "Last 30 Days" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+%d AND date<='%v' AND type='%v' ORDER BY date DESC", date_str, days, date_str, filter)
			} else if filterDate == "Last 3 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+90 AND date<='%v' AND type='%v' ORDER BY date DESC", date_str, date_str, filter)
			} else if filterDate == "Last 6 Months" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+180 AND date<='%v' AND type='%v' ORDER BY date DESC", date_str, date_str, filter)
			} else if filterDate == "Last Year" {
				query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+365 AND date<='%v' AND type='%v' ORDER BY date DESC", date_str, date_str, filter)
			}
		}
		fmt.Printf("ADD: %v\n", query)
		sent.filter_array = sql_query(db, query)
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

	e.GET("/filter-data", func(c echo.Context) error {
		// Wait 20ms
		fmt.Printf("Request Recieved Filter\n")

		category := c.FormValue("filter")
		fmt.Printf("%v\n", category)

		var query string

		if category == "All" {
			query = "SELECT * FROM expense ORDER BY date DESC"
		} else {
			query = fmt.Sprintf("SELECT * FROM expense WHERE type='%v' ORDER BY date DESC", category)
		}
		sent.filter_array = sql_query(db, query)

		return table_out(sent.filter_array).Render(context.Background(), c.Response().Writer)
	})

	e.GET("/filter-date", func(c echo.Context) error {
		// Wait 20ms
		fmt.Printf("Request Recieved filterDate")

		category := c.FormValue("filterDate")
		fmt.Printf("%v\n", category)

		var query string
		date_str := time.Now().Format("2006-01-02")
		days := DaysInMonth(time.Now())
		if category == "All Dates" {
			query = "SELECT * FROM expense ORDER BY date DESC"
		} else if category == "Last 7 Days" {
			query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+7 AND date<='%v' ORDER BY date DESC", date_str, date_str)
		} else if category == "Last 30 Days" {
			query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+%d AND date<='%v' ORDER BY date DESC", date_str, days, date_str)
		} else if category == "Last 3 Months" {
			query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+90 AND date<='%v' ORDER BY date DESC", date_str, date_str)
		} else if category == "Last 6 Months" {
			query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+180 AND date<='%v' ORDER BY date DESC", date_str, date_str)
		} else if category == "Last Year" {
			query = fmt.Sprintf("SELECT * FROM expense WHERE '%v'<=date+365 AND date<='%v' ORDER BY date DESC", date_str, date_str)
		}

		sent.filter_array = sql_query(db, query)

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
	log.Fatal(e.Start(fmt.Sprintf(":%v", port)))
}
