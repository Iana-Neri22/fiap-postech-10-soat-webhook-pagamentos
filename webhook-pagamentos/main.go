package main

import (
	"context"
	"fmt"
	"net/http"

	"webhook-pagamentos/payment"
	"webhook-pagamentos/sender"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

type PagamentoInput struct {
	IdPagamento int     `json:"id_pagamento" binding:"required"`
	Valor       float32 `json:"valor" binding:"required"`
}

type PagamentoResponse struct {
	IdPagamento int    `json:"id_pagamento"`
	Status      string `json:"status"`
}

type WebhookPayload struct {
	IdPagamento int    `json:"id_pagamento"`
	Status      string `json:"status"`
}

var ginLambda *ginadapter.GinLambda

func init() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/pagamento", func(c *gin.Context) {
		fmt.Println("🔥 Rota /pagamento chamada")

		var input PagamentoInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		status := payment.ProcessPayment()

		webhookPayload := WebhookPayload{
			IdPagamento: input.IdPagamento,
			Status:      status,
		}

		webhookURL := "https://qjy8d5de2c.execute-api.us-east-1.amazonaws.com/v1/pagamentos/confirmar"
		err := sender.SendWebhook(webhookPayload, webhookURL, "")
		if err != nil {
			c.Header("X-Webhook-Status", "failed")
		} else {
			c.Header("X-Webhook-Status", "sent")
		}

		c.JSON(http.StatusOK, PagamentoResponse{
			IdPagamento: input.IdPagamento,
			Status:      status,
		})
	})

	// ✅ transforma o Gin em um handler compatível com Lambda
	ginLambda = ginadapter.New(r)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
