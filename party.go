package party

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type MultipartRequest struct {
	Filepath      string
	FileFieldName string
	Params        map[string]string
}

func (r *MultipartRequest) Body() (*bytes.Buffer, string, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if r.Filepath != "" {
		var fileFieldName string

		f, err := os.Open(r.Filepath)
		if err != nil {
			return nil, "", "", err
		}

		if r.FileFieldName == "" {
			fileFieldName = "file"
		} else {
			fileFieldName = r.FileFieldName
		}

		part, err := writer.CreateFormFile(fileFieldName, r.Filepath)
		if err != nil {
			return nil, "", "", err
		}

		_, err = io.Copy(part, f)
	}

	if r.Params != nil {
		for k, v := range r.Params {
			if err := writer.WriteField(k, v); err != nil {
				return nil, "", "", err
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", "", err
	}

	return body, writer.FormDataContentType(), writer.Boundary(), nil
}

func (r *MultipartRequest) Request(method, url string) (*http.Request, error) {
	body, contentType, _, err := r.Body()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return req, nil
}
