package party

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
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

func ExampleMultipartRequest_Request() {
	params := &MultipartRequest{
		Filepath: "./dissertation.pdf",
		FileFieldName: "file",
		Params: map[string]string{
			"Author": "Luc Perkins",
			"Title": "The purposive Prometheus: re-imagining practical reason beyond homo œconomicus",
		},
	}

	req, err := params.Request(http.MethodPost, "https://example.com/dissertations")
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	log.Println(res.StatusCode)
}

func ExampleMultipartRequestHandler_Handle() {
	handler := &MultipartRequestHandler{
		MaxBytes: 32 << 20,
	}

	srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, err := handler.Handle(w, r)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Filename:", res.Header.Filename)
			bs, err := ioutil.ReadAll(res.File)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("File contents:")
			log.Print(string(bs))
		}),
	}

	log.Fatal(srv.ListenAndServe())
}