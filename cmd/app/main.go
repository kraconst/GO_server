package main

import (
	"log"

	"github.com/kraconst/GO_server/internal/config"
	"github.com/kraconst/GO_server/internal/handler"
	"github.com/kraconst/GO_server/internal/storage/postgres"
	"github.com/kraconst/GO_server/internal/usecase"
)

func main() {
	cfg := config.Must()

	db, err := postgres.New(cfg.DB.DSN())
	var accountUsecase *usecase.AccountUsecase
	var employeeUsecase *usecase.EmployeeUsecase
	var messageUsecase *usecase.MessageUsecase
	var contactUsecase *usecase.ContactUsecase
	var chatUsecase *usecase.ChatUsecase

	if err != nil {
		log.Printf("PostgreSQL connection failed: %v (run 'docker compose up -d' to start database)", err)
		log.Println("Starting server with mock fallback...")
	} else {
		defer db.Close()
		log.Println("Connected to PostgreSQL successfully")

		accountRepo := postgres.NewAccountStorage(db)
		employeeRepo := postgres.NewEmployeeStorage(db)
		messageRepo := postgres.NewMessageStorage(db)
		contactRepo := postgres.NewContactStorage(db)
		chatRepo := postgres.NewChatStorage(db)

		accountUsecase = usecase.NewAccountUsecase(accountRepo)
		employeeUsecase = usecase.NewEmployeeUsecase(employeeRepo)
		messageUsecase = usecase.NewMessageUsecase(messageRepo)
		contactUsecase = usecase.NewContactUsecase(contactRepo)
		chatUsecase = usecase.NewChatUsecase(chatRepo)
	}

	h := handler.New(accountUsecase, employeeUsecase, messageUsecase, contactUsecase, chatUsecase)
	log.Fatal(h.Run(cfg.Addr))
}
