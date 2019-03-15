package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

var db *sqlx.DB

type Superhero struct {
	ID 		uint 	`db:"superhero_id" json:"id"`
	Name 	string 	`json:"name"`
	Gender 	string 	`json:"gender"`
}

type Address struct {
	ID			uint 	`db:"address_id" json:"id"`
	Street 		string 	`json:"street"`
	City		string	`json:"city"`
	Country		string	`json:"country"`
}

type Power struct {
	ID 		uint 	`db:"power_id" json:"id"`
	Name 	string 	`json:"name"`
}

func InitDB() {
	db = sqlx.MustConnect("sqlite3", "superhero.db")

	schema := `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS superhero (
		superhero_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		gender TEXT
	);

	CREATE TABLE IF NOT EXISTS address (
		address_id INTEGER PRIMARY KEY AUTOINCREMENT,
		street TEXT,
		city TEXT,
		country TEXT,

		superhero_id INTEGER,
		FOREIGN KEY(superhero_id) REFERENCES superhero(superhero_id) ON DELETE CASCADE ON UPDATE CASCADE 
	);

	CREATE TABLE IF NOT EXISTS power (
		power_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	);

	CREATE TABLE IF NOT EXISTS superheropower (
		power_id INTEGER,
		superhero_id INTEGER,
		
		FOREIGN KEY(power_id) REFERENCES power(power_id) ON DELETE CASCADE ON UPDATE CASCADE ,
        FOREIGN KEY(superhero_id) REFERENCES superhero(superhero_id) ON DELETE CASCADE ON UPDATE CASCADE 
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		panic(err)
	}
}

func CreateSuperhero(c *gin.Context) {
	superheroSQL := "INSERT INTO superhero (name, gender) VALUES(?, ?)"

	name 	:= c.PostForm("name")
	gender 	:= c.PostForm("gender")
	db.MustExec(superheroSQL, name, gender)

	c.JSON(http.StatusCreated, gin.H{
		"name": name,
		"gender": gender,
	})
}

func RetrieveSuperhero(c *gin.Context) {
	superheroID  := c.Param("id")
	superheroSQL := `
	SELECT
    	sh.superhero_id,
	    sh.name,
	    sh.gender,
	    pw.power_id,
    	pw.name
	FROM
    	superhero sh
	INNER JOIN
    	superheropower shp on sh.superhero_id = shp.superhero_id
	INNER JOIN
    	power pw on shp.power_id = pw.power_id
	WHERE sh.superhero_id=?
	`

	var id 		  int
	var name 	  string
	var gender 	  string
	var powers	  []Power
	rows, _ := db.Query(superheroSQL, superheroID)
	for rows.Next() {
		var powerID 	int
		var powerName 	string
		err := rows.Scan(&id, &name, &gender, &powerID, &powerName)
		powers = append(powers, Power{ID:uint(powerID), Name:powerName})
		if err != nil {
			panic(err)
		}
	}

	type SuperheroDetail struct {
		Id			int       	`json:"id"`
		Name  		string    	`json:"name"`
		Gender		string    	`json:"gender"`
		Powers 		[]Power		`json:"powers"`
	}
	c.JSON(http.StatusOK, SuperheroDetail{Id:id, Name:name, Gender:gender, Powers:powers})
}

func ListSuperhero(c *gin.Context) {
	var superheroes []Superhero
	err := db.Select(&superheroes, "SELECT * FROM superhero ORDER BY superhero_id DESC")
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"results": superheroes,
	})
}

func RemoveSuperhero(c *gin.Context) {
	superheroID 	:= c.Param("id")
	superheroSQL 	:= "DELETE FROM superhero WHERE superhero_id=?"

	db.MustExec(superheroSQL, superheroID)
	c.JSON(http.StatusNoContent, nil)
}

func ApplySuperheroPower(c *gin.Context) {
	superheroID := c.Param("id")
	powerID		:= c.PostForm("power")

	superheroPowerSQL := "INSERT INTO superheropower (superhero_id, power_id) VALUES (?, ?)"
	db.MustExec(superheroPowerSQL, superheroID, powerID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func CreateAddress(c *gin.Context) {
	addressSQL := "INSERT INTO address (street, city, country, superhero_id) VALUES(?, ?, ?, ?)"

	street 		:= c.PostForm("street")
	city 		:= c.PostForm("city")
	country 	:= c.PostForm("country")
	superhero 	:= c.PostForm("superhero_id")
	db.MustExec(addressSQL, street, city, country, superhero)

	c.JSON(http.StatusCreated, gin.H{
		"street": street,
		"city": city,
		"country": country,
	})
}

func ListAddress(c *gin.Context) {
	// Only works if filtered by superhero `?superhero_id=`

	var listAddress []Address

	superheroID := c.DefaultQuery("superhero_id", "-1")
	addressSQL 	:= "SELECT address_id, street, city, country FROM address WHERE superhero_id=?"
	err := db.Select(&listAddress, addressSQL, superheroID)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"results": listAddress,
	})
}

func RemoveAddress(c *gin.Context) {
	addressID 	:= c.Param("id")
	addressSQL 	:= "DELETE FROM address WHERE address_id=?"

	db.MustExec(addressSQL, addressID)
	c.JSON(http.StatusNoContent, nil)
}

func CreatePower(c *gin.Context) {
	powerSQL := "INSERT INTO power (name) VALUES (?)"

	name := c.PostForm("name")
	db.MustExec(powerSQL, name)

	c.JSON(http.StatusCreated, gin.H{
		"name": name,
	})
}

func ListPower(c *gin.Context) {
	var powers []Power
	powerSQL := "SELECT * FROM power ORDER BY power_id DESC"
	err := db.Select(&powers, powerSQL)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"results": powers,
	})
}

func main() {
	InitDB()

	router := gin.Default()
	v1 := router.Group("/v1")
	{
		// Superhero.
		v1.POST("/superheroes", CreateSuperhero)
		v1.GET("/superheroes", ListSuperhero)
		v1.GET("/superheroes/:id", RetrieveSuperhero)
		v1.DELETE("/superheroes/:id", RemoveSuperhero)
		v1.POST("/superheroes/:id/powers", ApplySuperheroPower)

		// Address.
		v1.POST("/addresses", CreateAddress)
		v1.GET("/addresses", ListAddress)
		v1.DELETE("/addresses/:id", RemoveAddress)

		// Power.
		v1.POST("powers", CreatePower)
		v1.GET("powers", ListPower)
	}
	router.Run()
}
