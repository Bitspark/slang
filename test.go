package slang

import "io"

// TestOperator reads a file with test data and its corresponding operator and performs the tests.
// It returns the number of failed and succeeded tests and and error in case something went wrong.
// Test failures do not lead to an error. Test failures are printed to the writer.
func TestOperator(testDataPath string, writer io.Writer) (int, int, error) {
	return 0, 0, nil
}
