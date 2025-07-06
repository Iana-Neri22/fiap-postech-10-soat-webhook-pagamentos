package main

import (
	"net/http"
	"webhook-pagamentos/payment"
	"webhook-pagamentos/sender"

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

func main() {
	r := gin.Default()

	r.POST("/pagamento", func(c *gin.Context) {
		var input PagamentoInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		status := payment.ProcessPayment()

		// Prepare webhook payload
		webhookPayload := WebhookPayload{
			IdPagamento: input.IdPagamento,
			Status:      status,
		}

		// Send webhook
		webhookURL := "https://qjy8d5de2c.execute-api.us-east-1.amazonaws.com/v1/pagamentos/confirmar"
		err := sender.SendWebhook(webhookPayload, webhookURL, "")
		if err != nil {
			// Log the error but don't fail the request
			// You might want to handle this differently based on your requirements
			c.Header("X-Webhook-Status", "failed")
		} else {
			c.Header("X-Webhook-Status", "sent")
		}

		c.JSON(http.StatusOK, PagamentoResponse{
			IdPagamento: input.IdPagamento,
			Status:      status,
		})
	})

	r.Run(":8080")
}

// This is a simple HTTP server that processes payment requests.
// It listens for POST requests on the `/pagamento` endpoint, validates the input
