//     connPool, err := pgx.Connect(context.Background(), "postgres://root@192.168.86.74:26257/tickets?application_name=rc-app")

/*
export OTEL_EXPORTER_OTLP_ENDPOINT="http://192.168.86.74:4317"
export OTEL_EXPORTER_OTLP_PROTOCOL="grpc"
export OTEL_SERVICE_NAME="ie_app"

*/

package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	// "os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/XSAM/otelsql"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc"
)

// User represents a user in the system.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Initialize OpenTelemetry
	tp, err := initTracing()
	if err != nil {
		log.Fatalf("failed to initialize tracing: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	// Initialize Gin router
	r := gin.Default()

	// Define database connection
	connStr := "postgres://root@192.168.86.74:26257/tickets?application_name=rc-app"
	// db, err := sql.Open(otelsql.DriverName("pgx"), connStr)
	db, err := otelsql.Open("pgx", connStr, otelsql.WithAttributes(
	    semconv.DBSystemPostgreSQL, semconv.DBSystemCockroachdb,
    ))
	if err != nil {
		log.Fatalf("Unable to open database: %v\n", err)
	}
	defer db.Close()

	// Define API endpoints
	r.GET("/implicit/users/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		user, err := getUserImplicit(context.Background(), db, uuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	r.GET("/explicit/users/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		user, err := getUserExplicit(context.Background(), db, uuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// Start the Gin server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v\n", err)
	}
}

func initTracing() (*sdktrace.TracerProvider, error) {
	// Set up the connection to the collector.
	conn, err := grpc.DialContext(context.Background(), "192.168.86.74:4317", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to collector: %v", err)
	}

	// Set up the OTLP trace exporter.
	exporter, err := otlptrace.New(context.Background(),
		otlptracegrpc.NewClient(otlptracegrpc.WithGRPCConn(conn)),
	)
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}

	// Set up the trace provider.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("user-service"),
		attribute.String("environment", "development"),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	// Set up the global trace provider.
	otel.SetTracerProvider(tp)

	return tp, nil
}

func getUserImplicit(ctx context.Context, db *sql.DB, uuid string) (*User, error) {
	ctx, span := otel.Tracer("user-service").Start(ctx, "getUserImplicit")
	defer span.End()

	var user User
	err := db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id=$1", uuid).Scan(&user.ID, &user.Name)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	return &user, nil
}

// this uses database/sql directly which should use pgx
func getUserExplicit(ctx context.Context, db *sql.DB, uuid string) (*User, error) {
	ctx, span := otel.Tracer("user-service").Start(ctx, "getUserExplicit")
	defer span.End()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer tx.Rollback()

	var user User
	err = tx.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id=$1", uuid).Scan(&user.ID, &user.Name)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return nil, err
	}
	return &user, nil
}

