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
	"strings"
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
	e.GET("tk/all", GetAll)
	e.POST("tk/new", CreateTicket)
	e.POST("tk/update/:token", Update)
	/*run a server*/
	e.Logger.Print((e.Start(":2332")))
}

/* controller GetAll
- this controller returns info for all the users which are registered on to this service
it does a fetch all function,
- TODO it will perform pagination, if given the range
- TODO it will filter queries according email, lattitude and longtitude
*/
func GetAll(c echo.Context) error {
	tokens, err := GetAllTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}
	c.JSON(http.StatusOK, tokens)
	return nil
}

/* this is an all return function, it will return an array of tokens*/
func GetAllTokens() ([]Tk, error) {
	rows, err := dbConn.Query(`SELECT email,ticketno,location FROM userToken`)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var list []Tk

	var email string
	var location string
	var tokenId int

	for rows.Next() {
		rows.Scan(&email, &tokenId, &location)
		list = append(list, Tk{tokenId, email, location})
	}
	return list, nil
}

/*function to validate before entering new or updated
  values*/

func Validate(to Tk) error {
	return nil
}

/*
update controller, which is called when user wants to update details about oneself
a user can want to update their email  or location
*/
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
	err = Validate(ticket)
	if err != nil {
		c.JSON(http.StatusBadRequest, 0)
		return err
	}

	err = UpdateToken(ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}
	c.JSON(http.StatusOK, "updated")
	return nil
}

/*
this is the function that updates all the variables to database, this is called by
an edititng or updating controller
*/
func UpdateToken(to Tk) error {
	location := to.Location
	locationSplit := strings.Split(location, ",")
	lat := strings.TrimPrefix(locationSplit[0], "{")
	long := strings.TrimSuffix(locationSplit[1], "}")
	_, err := dbConn.Exec(fmt.Sprintf(`UPDATE userToken SET email='%s', lat='%s',long='%s' WHERE ticketno='%d'`, to.Email, lat, long, to.Id))
	if err != nil {
		return err
	}
	return nil
}

/* Create token controller, this is the initial token create controller*/
func CreateTicket(c echo.Context) error {
	defer c.Request().Body.Close() /* body will close when response is ended*/
	var ticket Tk
	tk, err := GetReqJSON(c.Request())
	ticket = tk.(Tk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 0)
		return err
	}

	err = Validate(ticket)
	if err != nil {
		c.JSON(http.StatusBadRequest, 0)
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

/*
the ticket creation struct, it holds a token number, whic is the Id and the email as well as the
location of the token creation
*/
type Tk struct {
	Id       int    `json:id`
	Email    string `json:email`
	Location string `json:location`
}

func CreateToken(to Tk) error {
	location := to.Location
	locationSplit := strings.Split(location, ",")
	lat := strings.TrimPrefix(locationSplit[0], "{")
	long := strings.TrimSuffix(locationSplit[1], "}")
	_, err := dbConn.Exec(fmt.Sprintf(`INSERT INTO userToken(ticketno, email,lat,long) values(%d,'%s','%s','%s')`, to.Id, to.Email, lat, long))
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
