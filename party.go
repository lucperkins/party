package party

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const defaultFileFieldName = "file"

var ErrEmptyRequest = errors.New("request has no file and no request params")

type MultipartRequest struct {
	Filepath      string
	FileFieldName string
	Boundary      string
	Params        map[string]string
}

type MultipartRequestHandler struct {
	MaxBytes      int64
	FileFieldName string
}

func (h *MultipartRequestHandler) Handle(w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := h.validate(); err != nil {
		return nil, nil, err
	}

	if err := r.ParseMultipartForm(h.MaxBytes); err != nil {
		return nil, nil, err
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBytes)

	if h.FileFieldName == "" {
		h.FileFieldName = defaultFileFieldName
	}

	return r.FormFile(h.FileFieldName)
}

func (c *MultipartRequest) body() (*bytes.Buffer, string, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if c.Boundary != "" {
		if err := writer.SetBoundary(c.Boundary); err != nil {
			return nil, "", "", err
		}
	}

	if c.Filepath != "" {
		f, err := os.Open(c.Filepath)
		if err != nil {
			return nil, "", "", err
		}

		if c.FileFieldName == "" {
			c.FileFieldName = defaultFileFieldName
		}

		part, err := writer.CreateFormFile(c.FileFieldName, c.Filepath)
		if err != nil {
			return nil, "", "", err
		}

		_, err = io.Copy(part, f)
	}

	if c.Params != nil {
		for k, v := range c.Params {
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

func (c *MultipartRequest) Request(method, url string) (*http.Request, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	body, contentType, _, err := c.body()
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

func (h *MultipartRequestHandler) validate() error {
	if h.FileFieldName == "" {
		h.FileFieldName = defaultFileFieldName
	}

	return nil
}

func (c *MultipartRequest) validate() error {
	if c.Filepath == "" && c.Params == nil {
		return ErrEmptyRequest
	}

	return nil
}
