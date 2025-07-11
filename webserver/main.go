package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/microwatcher/shared/pkg/clickhouse"
	"github.com/microwatcher/shared/pkg/logger"
	"github.com/microwatcher/webserver/internal/graph"
	"github.com/microwatcher/webserver/internal/otlp"
	"github.com/vektah/gqlparser/v2/ast"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	logger := logger.NewDefaultLogger()

	otelShutdown := otlp.InitLocalTracer(context.Background(), logger)
	defer otelShutdown()

	localSource, err := clickhouse.NewLocalConnection(logger)
	if err != nil {
		logger.Error("failed to create local clickhouse connection",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	r := gin.Default()
	r.Use(otelgin.Middleware(otlp.ServiceName))

	r.POST("/query", graphqlHandler(localSource))
	r.GET("/", playgroundHandler())
	r.Run()
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func graphqlHandler(chSource *clickhouse.ClickhouseSource) gin.HandlerFunc {
	h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		ChSource: chSource,
	}}))

	// Server setup:
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
