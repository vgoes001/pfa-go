package main

import (
	"database/sql"

	_ "github.com/vgoes001/pfa-go/internal/order/entity"
	"github.com/vgoes001/pfa-go/internal/order/infra/database"
	"github.com/vgoes001/pfa-go/internal/order/usecase"
)


func main(){
	db, err := sql.Open("mysql", "root:root@tcp(mysql:3306)/orders")

	if err != nil {
		panic(err)
	}

	defer db.Close() // fecha a conexão após executar todo o código

	repository := database.NewOrderRepository(db)
	u := usecase.NewCalculateFinalPriceUseCase(repository)

	input:= usecase.OrderInputDTO{
		ID: "1",
		Price: 100,
		Tax: 2,
	}
	output, err := u.Execute(input)

	if err != nil{
		panic(err)
	}
	println(output.FinalPrice)

}