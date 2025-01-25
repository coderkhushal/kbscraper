package outputs

import (
	"fmt"
	"os"
)

func WriteJson(content []byte, filename string, filepath string) error {
	if content == nil || filename == "" {
		return fmt.Errorf("content, filename or filepath is empty \n")
	}
	f, err := os.Create(filepath + filename)
	if err != nil {
		return fmt.Errorf("error creating file %s \n", err)
	}
	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("error writing file %s \n", err)
	}

	return nil
}
