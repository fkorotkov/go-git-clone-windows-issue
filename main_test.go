package main

import (
	"testing"
)

func TestClone(t *testing.T) {
	dir := "temp-repo"
	clone("https://github.com/lumen/lumen", "windows", "911267b21097ea70bf2ccdfd41152313525237fb", dir)
	t.Logf("Cloned into %s", dir)
}
