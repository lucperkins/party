package party

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"mime/multipart"
	"testing"
)

const (
	badFile           = "does-not-exist.txt"
	testFile          = "party.go"
	testFileFieldName = "upload-file"
)

func TestRequestBodyCreation(t *testing.T) {
	is := assert.New(t)

	req := &MultipartRequest{}
	_, _, _, err := req.body()
	is.NoError(err)

	req.Filepath = badFile
	_, _, _, err = req.body()
	is.Error(err)

	req = &MultipartRequest{
		Filepath:      testFile,
		FileFieldName: testFileFieldName,
	}

	body, contentType, boundary, err := req.body()
	is.NoError(err)
	is.NotNil(body)
	is.NotEmpty(contentType)
	is.NotEmpty(boundary)

	bs, err := ioutil.ReadFile(testFile)
	is.NoError(err)
	is.NotNil(bs)

	reader := multipart.NewReader(body, boundary)
	is.NoError(err)
	is.NotNil(reader)

	form, err := reader.ReadForm(0)
	is.NoError(err)
	is.NotNil(form)

	file := form.File[testFileFieldName][0]
	is.Equal(file.Filename, testFile)

	f, err := file.Open()
	is.NoError(err)
	is.NotNil(f)
	is.Equal(file.Size, int64(len(bs)))

	fileBytes, err := ioutil.ReadAll(f)
	is.NoError(err)
	is.NotNil(fileBytes)

	is.Equal(bs, fileBytes)
}

func TestMultipartRequestHandler(t *testing.T) {
	is := assert.New(t)

	handler := MultipartRequestHandler{
		MaxBytes: 32 << 20,
	}
	is.NoError(handler.validate())

	is.Equal(handler.FileFieldName, defaultFileFieldName)
	is.Equal(handler.MaxBytes, int64(32 << 20))
}