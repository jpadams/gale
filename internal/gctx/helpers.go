package gctx

import (
	"fmt"
	"reflect"

	"dagger.io/dagger"

	"github.com/caarlos0/env/v9"

	"github.com/aweris/gale/internal/dagger/helpers"
)

// NewContextFromEnv initializes a new context from environment variables. It works with structs having exported fields
// tagged with the `env` tag.
func NewContextFromEnv[T any]() (T, error) {
	val := new(T)

	if err := env.Parse(val); err != nil {
		return *val, err
	}

	return *val, nil
}

const trueStr = "true"

// WithContainerEnv loads context fields into the container as environment variables or secrets.
// It expects a struct with exported fields. Fields tagged with `container_env` or `container_secret` are loaded
// using the `env` tag value as their name.
func WithContainerEnv[T any](client *dagger.Client, t *T) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return loadFieldsIntoContainer(client, container, t)
	}
}

func loadFieldsIntoContainer(client *dagger.Client, container *dagger.Container, t interface{}) *dagger.Container {
	val := reflect.ValueOf(t).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		containerEnvTag := typ.Field(i).Tag.Get("container_env")
		containerSecretTag := typ.Field(i).Tag.Get("container_secret")

		// skip if the field is not tagged with container_env or container_secret
		if containerEnvTag == "" && containerSecretTag == "" {
			continue
		}

		if containerEnvTag == trueStr && containerSecretTag == trueStr {
			return helpers.FailPipeline(container, fmt.Errorf("field %s is tagged with both container_env and container_secret", field.Name))
		}

		var (
			fieldVal = val.Field(i)
			envTag   = field.Tag.Get("env")
		)

		// if env tag is empty and the field is not a struct, fail the pipeline. We're using env tag as the name of the
		// environment variable or secret. If it's empty, we can't load the field.
		if envTag == "" && val.Field(i).Kind() != reflect.Struct {
			return helpers.FailPipeline(container, fmt.Errorf("field %s is tagged with container_env or container_secret but not tagged with env", field.Name))
		}

		switch {
		case fieldVal.Kind() == reflect.Struct:
			container = loadFieldsIntoContainer(client, container, fieldVal.Addr().Interface())
		case containerEnvTag == trueStr:
			container = container.WithEnvVariable(envTag, fmt.Sprintf("%v", fieldVal.Interface()))
		case containerSecretTag == trueStr:
			container = container.WithSecretVariable(envTag, client.SetSecret(envTag, fmt.Sprintf("%v", fieldVal.Interface())))
		default:
			return helpers.FailPipeline(container, fmt.Errorf("unsupported field type: %s", fieldVal.Kind()))
		}
	}

	return container
}