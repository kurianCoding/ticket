package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"strconv"
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
	e.POST("tk/update/:token", Update)
	/*run a server*/
	e.Logger.Error((e.Start(":2332")))
}

func Update(c echo.Context) error {
	defer c.Request().Body.Close()
	var ticket Tk
	tk, err := GetReqJSON(c.Request())
	ticket = tk.(Tk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}
	/* set the token*/
	ticket.Id, err = strconv.Atoi(c.Param("token"))
	err = UpdateToken(ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}
	c.JSON(http.StatusOK, "updated")
	return nil
}

func UpdateToken(to Tk) error {
	_, err := dbConn.Exec(fmt.Sprintf(`UPDATE userToken SET email='%s' WHERE ticketno='%d'`, to.Email, to.Id))
	if err != nil {
		return err
	}
	return nil
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
	err = CreateToken(ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}
	c.JSON(http.StatusOK, ticket.Id)
	return nil
}

type Tk struct {
	Id    int    `json:id`
	Email string `json:email`
}

func CreateToken(to Tk) error {
	_, err := dbConn.Exec(fmt.Sprintf(`INSERT INTO userToken(ticketno, email) values(%d,'%s')`, to.Id, to.Email))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
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
