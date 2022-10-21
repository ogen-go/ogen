package integration_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_servers"
)

func TestServers(t *testing.T) {
	t.Run("Template", func(t *testing.T) {
		a := require.New(t)
		s := api.ProductionServer{}

		r := s.MustBuild()
		a.Equal("https://us.example.com/prod/v1", r)
		r = s.MustPath()
		a.Equal("/prod/v1", r)

		s.Region = "eu"
		r = s.MustBuild()
		a.Equal("https://eu.example.com/prod/v1", r)
		r = s.MustPath()
		a.Equal("/prod/v1", r)

		s.Region = "ru"
		_, err := s.Build()
		a.Error(err)
		a.Panics(func() {
			s.MustBuild()
		})
		_, err = s.Path()
		a.Error(err)
		a.Panics(func() {
			s.MustPath()
		})
	})
	t.Run("Const", func(t *testing.T) {
		a := require.New(t)
		s := api.ConstServer

		r := s.MustBuild()
		a.Equal("https://cdn.example.com/v1", r)
		r = s.MustPath()
		a.Equal("/v1", r)
	})
}
