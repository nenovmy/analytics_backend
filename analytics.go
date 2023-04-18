package main

import (
	"fmt"

	"bitbucket.org/analytics-backend/conf"

	"bitbucket.org/analytics-backend/server"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s "+
		"dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Password, conf.Dbname)
	s := server.NewServer(psqlInfo, 100, "localhost:12345")
	fmt.Println("Analytics initialized...")

	err := s.StartServer()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
