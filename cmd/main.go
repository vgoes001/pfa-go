package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	_ "github.com/vgoes001/pfa-go/internal/order/entity"
	"github.com/vgoes001/pfa-go/internal/order/infra/database"
	"github.com/vgoes001/pfa-go/internal/order/usecase"
	"github.com/vgoes001/pfa-go/pkg/rabbitmq"
)


func main(){
	maxWorkers := 1
	wg := sync.WaitGroup{}
	db, err := sql.Open("mysql", "root:root@tcp(mysql:3306)/orders")

	if err != nil {
		panic(err)
	}

	defer db.Close() // fecha a conexão após executar todo o código

	repository := database.NewOrderRepository(db)
	u := usecase.NewCalculateFinalPriceUseCase(repository)


	http.HandleFunc("/total", func(w http.ResponseWriter, r *http.Request){
		uc := usecase.NewGetTotalUseCase(repository)
		output, err :=uc.Execute()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(output)
	})
	go http.ListenAndServe(":8080", nil)


	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()
	out := make(chan amqp.Delivery)
	go rabbitmq.Consume(ch, out)
	wg.Add(maxWorkers)
	for i:=0; i < maxWorkers; i++{
		defer wg.Done()
		go worker(out, u, i)
	}
	wg.Wait()
}

func worker(deliveryMessage <-chan amqp.Delivery, uc *usecase.CalculateFinalPriceUseCase, workerId int){
	for msg := range deliveryMessage{
		var input usecase.OrderInputDTO
		err := json.Unmarshal(msg.Body, &input)
		if err != nil{
			fmt.Println("Error unmarshalling message", err)
		}
		input.Tax = 10
		_, err = uc.Execute(input)
		if err != nil{
			fmt.Println("Error unmarshalling message", err)
		}
		msg.Ack(false) // retira a mensagem da fila
		fmt.Println("Worker", workerId, "processedd order", input.ID)
	}
}