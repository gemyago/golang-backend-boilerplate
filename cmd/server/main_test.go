package main

import (
	"errors"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	t.Run("run", func(t *testing.T) {
		t.Run("should initialize the app", func(t *testing.T) {
			assert.NotPanics(t, func() {
				run(runOpts{
					rootLogger:     diag.RootTestLogger(),
					noopHTTPListen: true,
					cfg:            &config{},
				})
			})
		})
	})

	t.Run("mustNoErrors", func(t *testing.T) {
		t.Run("should not panic if no errors", func(t *testing.T) {
			assert.NotPanics(t, func() {
				mustNoErrors(nil, nil, nil)
			})
		})

		t.Run("should panic if error", func(t *testing.T) {
			assert.Panics(t, func() {
				mustNoErrors(nil, nil, errors.New(faker.Sentence()))
			})
		})
	})
}
