package otelgorm

import ()

/*
Package otelgorm provides OpenTelemetry tracing for GORM.

This plugin automatically traces all database operations (Create, Query, Update,
Delete, Row, Raw) with standard DB semantic conventions.

Example:

	import (
		"gorm.io/gorm"
		"github.com/QuickBill-Engineering/qbp-lib/tracing/otelgorm"
	)

	func main() {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		err = db.Use(otelgorm.NewPlugin(
			otelgorm.WithDBName("mydb"),
			otelgorm.WithLogQueries(true),
		))
		if err != nil {
			log.Fatal(err)
		}

		// All DB operations are now traced...
	}
*/
