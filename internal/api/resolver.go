package api

import (
	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel/trace"

	apigen "atomys.codes/stud42/internal/api/generated"
	modelgen "atomys.codes/stud42/internal/models/generated"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client *modelgen.Client
	tracer trace.Tracer
}

type contextKey string

// NewSchema creates a graphql executable schema.
func NewSchema(client *modelgen.Client, tr trace.Tracer) graphql.ExecutableSchema {
	return apigen.NewExecutableSchema(apigen.Config{
		Resolvers: &Resolver{
			client: client,
			tracer: tr,
		},
		Directives: apigen.DirectiveRoot{
			AuthzByPolicy: directiveAuthzByPolicy,
			Authenticated: directiveAuthorization(client),
		},
	})
}
