package app

import (
	"fmt"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/handler"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/repo"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/usecase"
)

func Run() {
	r := repo.New()
	uc := usecase.New(r)
	h := handler.New(uc)
	h.InitRoutes()
	if err := h.Run(":8080"); err != nil {
		fmt.Println(err)
	}
}
