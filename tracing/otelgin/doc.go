package otelgin

import ()

/*
Package otelgin provides OpenTelemetry tracing middleware for Gin.

This package follows the naming convention of official OpenTelemetry contrib packages
(otelhttp, otelgin, etc.) to avoid collisions with framework packages.

Example:

	package main

	import (
		"github.com/gin-gonic/gin"
		"github.com/QuickBill-Engineering/qbp-lib/tracing"
		"github.com/QuickBill-Engineering/qbp-lib/tracing/otelgin"
	)

	func main() {
		shutdown, err := tracing.InitFromEnv()
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown(context.Background())

		r := gin.Default()
		r.Use(otelgin.RequestID())
		r.Use(otelgin.Middleware())

		r.GET("/users/:id", func(c *gin.Context) {
			// Your handler...
		})

		r.Run(":8080")
	}
*/
