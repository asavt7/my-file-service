package tests

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

const (
	downloadPathTemplate = "%s/api/download/%s"
	uploadPathTemplate   = "%s/api/upload"
	pathToTestFile       = "../tests/data/cat.jpg"
)

func (m *MainTestSuite) TestUploadAndDownload() {
	t := m.T()
	var location string

	t.Run("upload", func(t *testing.T) {

		b, w := createMultipartFormData(t, "file", pathToTestFile)

		req, err := http.NewRequest("POST", fmt.Sprintf(uploadPathTemplate, m.baseurl), &b)
		if err != nil {
			return
		}
		// Don't forget to set the content type, this will contain the boundary.
		req.Header.Set("Content-Type", w.FormDataContentType())

		client := http.Client{}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if response.StatusCode > 299 {
			t.Errorf("Error status code!")
		}

		location = response.Header.Get("Location")
		if len(location) == 0 {
			t.Errorf("Expected location header in response")
		}
	})

	t.Run("download", func(t *testing.T) {
		downloadTest(t, location, func(t *testing.T, response *http.Response) {
			if response.StatusCode != http.StatusOK {
				t.Errorf("expected status code 200, got %d", response.StatusCode)
			}

			responseBody, err := io.ReadAll(response.Body)
			if err != nil {
				t.Errorf("cannot download file %+v", err)
			}

			err = response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			testFileContent := bytes.NewBuffer(nil)
			f, _ := os.Open(pathToTestFile) // Error handling elided for brevity.
			_, err = io.Copy(testFileContent, f)
			if err != nil {
				t.Fatalf("cannot load test file %+v\n", err)
			} // Error handling elided for brevity.
			f.Close()

			if !bytes.Equal(responseBody, testFileContent.Bytes()) {
				t.Errorf("downloaded file != uploaded file")
			}

		})
	})
}

func createMultipartFormData(t *testing.T, fieldName, fileName string) (bytes.Buffer, *multipart.Writer) {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	var fw io.Writer
	file := mustOpen(fileName)
	if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
		t.Errorf("Error creating writer: %v", err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		t.Errorf("Error with io.Copy: %v", err)
	}
	err = w.Close()
	if err != nil {
		t.Errorf("Error with io.Copy: %v", err)
	}
	return b, w
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		pwd, _ := os.Getwd()
		fmt.Println("PWD: ", pwd)
		panic(err)
	}
	return r
}

func (m *MainTestSuite) TestDownloadNotFound() {
	t := m.T()

	t.Run("Not found image", func(t *testing.T) {
		downloadTest(t, fmt.Sprintf(downloadPathTemplate, m.baseurl, "not_found.jpeg"), func(t *testing.T, response *http.Response) {
			if response.StatusCode != http.StatusNotFound {
				t.Errorf("expected status code 404, got %d", response.StatusCode)
			}
			err := response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		})

	})
}

func downloadTest(t *testing.T, url string, responseTestFunc func(t *testing.T, res *http.Response)) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	responseTestFunc(t, response)
}
