package internal

import "testing"

func TestPostTitle(t *testing.T) {
	input := "Test quote \""
	expectedOutput := "Test quote"

	output := PostTitle(input)

	if expectedOutput != output {
		t.Errorf("Failed ! got '%s' want '%s'", output, expectedOutput)
	} else {
		t.Log("Success !")
	}
}
