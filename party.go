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

type (
	// Multipart request configuration. A file can be included with the request and/or form parameters. An ErrEmptyRequest
	// is returned if neither is present.
	MultipartRequest struct {
		// Path to the file to include in the request (optional)
		Filepath string
		// The field name for the file included in the request (default is "file")
		FileFieldName string
		// The request boundary (automatically generated if none supplied)
		Boundary string
		// Multipart request parameters (optional)
		Params map[string]string
	}

	// Multipart request handler configuration.
	MultipartRequestHandler struct {
		// Maximum allowable bytes in the request
		MaxBytes int64
		// The field name for the file included in the request (default is "file")
		FileFieldName string
	}

	// A multipart response extracted from an http.Request object
	MultipartResponse struct {
		// The file included in the request
		File multipart.File
		// The header metadata describing the file
		Header *multipart.FileHeader
	}
)

// Translate an incoming HTTP request into a multipart file and a multipart file header (or an error).
func (h *MultipartRequestHandler) Handle(w http.ResponseWriter, r *http.Request) (*MultipartResponse, error) {
	if err := h.validate(); err != nil {
		return nil, err
	}

	if err := r.ParseMultipartForm(h.MaxBytes); err != nil {
		return nil, err
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBytes)

	if h.FileFieldName == "" {
		h.FileFieldName = defaultFileFieldName
	}

	file, header, err := r.FormFile(h.FileFieldName)
	if err != nil {
		return nil, err
	}

	return &MultipartResponse{
		File:   file,
		Header: header,
	}, nil
}

// Creates a request body (as a byte buffer) out of the supplied multipart request configuration and also returns the
// content type of the request, and the boundary of the request, largely for testing purposes (or an error).
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

// Translates the multipart request configuration into a pointer to an http.Request or an error.
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

// Validates the multipart request handler config
func (h *MultipartRequestHandler) validate() error {
	if h.FileFieldName == "" {
		h.FileFieldName = defaultFileFieldName
	}

	return nil
}

// Validates the multipart request config
func (c *MultipartRequest) validate() error {
	if c.Filepath == "" && c.Params == nil {
		return ErrEmptyRequest
	}

	return nil
}
