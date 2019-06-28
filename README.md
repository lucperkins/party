# Party

[![](https://godoc.org/github.com/lucperkins/party?status.svg)](http://godoc.org/github.com/lucperkins/party)
[![Actions Status](https://action-badges.now.sh/lucperkins/party?action=test)](https://github.com/lucperkins/party/actions)

A Go library for working with multipart form requests.

## Purpose

Go's core libraries are extremely good in general but not always terribly ergonomic and downright clunky in some places. One area of frustration for me recently has been the [`mime/multipart`](https://godoc.org/mime/multipart) library and dealing with multipart form requests. I created this library to enable you to do two things easily:

* Create an [`http.Request`](https://godoc.org/net/http#Request) object containing a file and a map of multipart form data. Here's an example:

    ```go
    multipartRequest := &party.MultipartRequest{
        Filepath: "./article.pdf", // Path to the file to include
        FileFieldName: "pdf",      // Defaults to "file"
        Boundary: "asdf4321",      // Optional. This is set automatically if none is supplied
        Params: map[string]string{ // The form data params to include
            "Author": "Luc Perkins",
        }
    }

    // Now create an HTTP request
    req, err := multipartRequest.Request(http.MethodPost, "https://example/com")
    if err != nil {
        handleError(err)
    }

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        handleError(err)
    }
    ```

* Parse an [`http.Request`](https://godoc.org/net/http#Request) in an HTTP handler into a [`multipart.File`](https://godoc.org/mime/multipart#File) and [`multipart.FileHeader`](https://godoc.org/mime/multipart#FileHeader). Here's an example:

    ```go
    multipartHandler := &party.MultipartRequestHandler{
        MaxBytes:      32 << 20,    // 32 MB max
        FileFieldName: "text-file", // Defaults to "file"
    }

    func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
        payload, err := multipartHandler.Handle(w, r)
        if err != nil {
            handleError(err)
        }

        file := payload.File
        header := payload.Header

        // Do something with the file and header
    }
    ```

## API

You can find full API docs on [GoDoc](https://godoc.org/github.com/lucperkins/party).
