package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	"database/sql"
	"log"
	"os"
	"strconv"
	_"github.com/lib/pq" 
	
)


type Customer struct {
	ID int  `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Status string `json:"status"`
}

var db 	*sql.DB

func authMiddleware(c *gin.Context)  {
	
	token := c.GetHeader("Authorization")
	if token != "token2019"{
		c.JSON(http.StatusUnauthorized,gin.H{"error":"unautorized."})
		c.Abort()
		return
	} 
	c.Next()
	
}
 
func creatCustomersHandler(c *gin.Context)  {    
	var cust Customer
	err := c.ShouldBindJSON(&cust)   

	if  err != nil {
		c.JSON(http.StatusBadRequest,err.Error())
		return
	}
	fmt.Println("Database : ",db)

	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id ",cust.Name,cust.Email,cust.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't scan id"}) 
		return
	}
	cust.ID = id
	c.JSON(http.StatusCreated,cust) 
}

func getOneCustomerHandler(c *gin.Context) {
	
	id := c.Param("id") 
	stmt,err := db.Prepare("SELECT id,name,email,status FROM customers where id=$1")
	
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't prepare query one row statment"}) 
		return
	}
	rowId,err := strconv.Atoi(id)  
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't convert string to int"}) 
		return
	}
	row := stmt.QueryRow(rowId)

	var custs Customer

	err = row.Scan(&custs.ID,&custs.Name,&custs.Email,&custs.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't Scan row into varibles"}) 
		return
	}
	c.JSON(http.StatusOK,custs)
}

func getAllCustomersHandler(c *gin.Context)  {    
	stmt,err := db.Prepare("SELECT id,name,email,status FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status": "can't prepare query all customers"})
		return

	}
	row,err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status": "Error!!! Query"})
		return

	}
	var custs []Customer
	for row.Next(){
		var id int
		var name ,email ,status string
		err := row.Scan(&id,&name,&email,&status)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"status":"can't scan row into variable"}) 
			return
		}
		custs = append(custs,Customer{id,name,email,status})
	}
	c.JSON(http.StatusOK,custs)

}

func updateCustomerhandler(c *gin.Context) {
	id := c.Param("id") 
	rowId,err := strconv.Atoi(id) 
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't convert string to int"}) 
		return
	}

	var cust Customer
	err = c.ShouldBindJSON(&cust) 

	stmt ,err := db.Prepare("UPDATE customers SET name=$2,email=$3,status=$4 where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't prepare statment update"}) 
		return

	}

	if _,err := stmt.Exec(rowId,cust.Name,cust.Email,cust.Status); err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error execute update"}) 
		return
	}
	c.JSON(http.StatusOK,cust)

}

func deleteCustomerhandler(c *gin.Context) {
	id := c.Param("id") 
	rowId,err := strconv.Atoi(id) 
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't convert string to int"}) 
		return
	}

	var cust Customer
	err = c.ShouldBindJSON(&cust) 
	stmt ,err := db.Prepare("DELETE FROM customers where id=$1")
	
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't prepare statment delete"+err.Error()}) 
		return
	}

	if _,err := stmt.Exec(rowId); err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error execute delete"+err.Error()}) 
		return
	}
	
	c.JSON(http.StatusOK,map[string]string{"message": "customer deleted"}) 
}

func main() {
	var err error
	db,err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal("Connect to database error",err)
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF  NOT EXISTS customers (    
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`  
	_,err= db.Exec(createTb) 
	if err != nil {
		log.Fatal("Connect to database error",err)
	}
	

	r := gin.Default()
	r.Use(authMiddleware)
	r.POST("/customers",creatCustomersHandler)
	r.GET("/customers/:id",getOneCustomerHandler)
	r.GET("/customers",getAllCustomersHandler) 
	r.PUT("/customers/:id",updateCustomerhandler) 
	r.DELETE("/customers/:id",deleteCustomerhandler) 
	r.Run(":2019")

}