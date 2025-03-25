package swagger

import (
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/swaggo/swag"
)

type mockedSwag struct{}

func (s *mockedSwag) ReadDoc() string {
	return `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {}
}`
}

var (
	registrationOnce sync.Once
)

func Test_Swagger(t *testing.T) {
	app := fiber.New()

	registrationOnce.Do(func() {
		swag.Register(swag.Name, &mockedSwag{})
	})

	app.Get("/swag/*", HandlerDefault)

	tests := []struct {
		name        string
		url         string
		statusCode  int
		contentType string
		location    string
	}{
		{
			name:        "Should be returns status 200 with 'text/html' content-type",
			url:         "/swag/index.html",
			statusCode:  200,
			contentType: "text/html",
		},
		{
			name:        "Should be returns status 200 with 'application/json' content-type",
			url:         "/swag/doc.json",
			statusCode:  200,
			contentType: "application/json",
		},
		{
			name:        "Should be returns status 200 with 'image/png' content-type",
			url:         "/swag/favicon-16x16.png",
			statusCode:  200,
			contentType: "image/png",
		},
		{
			name:       "Should return status 301",
			url:        "/swag/",
			statusCode: 301,
			location:   "/swag/index.html",
		},
		{
			name:       "Should return status 404",
			url:        "/swag/notfound",
			statusCode: 404,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != test.statusCode {
				t.Fatalf(`StatusCode: got %v - expected %v`, resp.StatusCode, test.statusCode)
			}

			if test.contentType != "" {
				ct := resp.Header.Get("Content-Type")
				if ct != test.contentType {
					t.Fatalf(`Content-Type: got %s - expected %s`, ct, test.contentType)
				}
			}

			if test.location != "" {
				location := resp.Header.Get("Location")
				if location != test.location {
					t.Fatalf(`Location: got %s - expected %s`, location, test.location)
				}
			}
		})
	}
}

func Test_Swagger_Proxy_Redirect(t *testing.T) {
	app := fiber.New()

	registrationOnce.Do(func() {
		swag.Register(swag.Name, &mockedSwag{})
	})

	// Use new handler since the prefix is created only once per handler
	app.Get("/swag/*", New())

	statusCode := 301
	location := "/custom/path/swag/index.html"

	t.Run("Should return status 301 with proxy redirect", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/swag/", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("X-Forwarded-Prefix", "/custom/path/")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != statusCode {
			t.Fatalf(`StatusCode: got %v - expected %v`, resp.StatusCode, statusCode)
		}

		if location != "" {
			responseLocation := resp.Header.Get("Location")
			if responseLocation != location {
				t.Fatalf(`Location: got %s - expected %s`, responseLocation, location)
			}
		}
	})
}
