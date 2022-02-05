package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGHRepoChecker(t *testing.T) {
	assert.True(t, githubRepoExists("datarootsio", "cheek"))
	assert.False(t, githubRepoExists("datarootsi000o", "cheek"))
}
