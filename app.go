package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
)

/* the idea is to make a ticket posting api
   which will accept a post request and update
   that in the database
*/
var dbConn *sql.DB

func main() {
	/*intialize databse connection*/
	connStr := fmt.Sprintf("host=%s user=%s port=%d password=%s dbname=%s sslmode=disable", "localhost", "app", 5432, "test", "ticket")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	dbConn = db
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	/*create an endpoint for posting a ticket*/
	/*declare a new router*/
	e := echo.New()
	/*declare a new endpoint on the router*/
	e.POST("tk/new", CreateTicket)
	/*run a server*/
	e.Logger.Fatal((e.Start(":2332")))
}

func CreateTicket(c echo.Context) error {
	defer c.Request().Body.Close() /* body will close when response is ended*/
	var ticket Tk
	tk, err := GetReqJSON(c.Request())
	ticket = tk.(Tk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}
	c.JSON(http.StatusOK, ticket.Id)
	return nil
}

type Tk struct {
	Id int
}

func GetReqJSON(r *http.Request) (interface{}, error) {
	/*read json from request*/

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var token Tk
	err = json.Unmarshal(b, &token)
	if err != nil {
		return nil, err
	}
	return token, nil
}
