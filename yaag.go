package middlewares
/**
//
// 	e.Static("/docs/", "docs")
//	e.Use(yaagDoc.YaagDoc())
//
**/


import (
	"bytes"
	"net/http/httptest"
	"strings"

	"github.com/betacraft/yaag/yaag"
	"github.com/betacraft/yaag/yaag/models"
	"github.com/labstack/echo"
)

// YaagDoc - middleware для yaag
func YaagDoc() echo.MiddlewareFunc {

	yaag.Init(&yaag.Config{
		On:       true,
		DocTitle: "yaag.documentation",
		DocPath:  "docs/api.html",
		BaseUrls: map[string]string{
			"api": "api.example.com",
		},
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// @todo тут проверку через конфиг надо
			if !yaag.IsOn() {
				return next(c)
			}

			response := c.Response()

			// сохраняем оригинальный writer
			originalWriter := response.Writer()

			// создаём фейковый writer
			fakeWriter := httptest.NewRecorder()

			response.SetWriter(fakeWriter)

			// выполняем действие
			p := next(c)

			// пропихивает в факерный врайтер оригинальный код
			fakeWriter.Code = response.Status()

			// возвращаем оригинальный writer
			response.SetWriter(originalWriter)
			// пишем во writer оригинальный ответ
			response.Write(fakeWriter.Body.Bytes())

			if fakeWriter.Code != 404 {

				request := c.Request()

				requestHeaders := map[string]string{}
				for _, k := range request.Header().Keys() {
					requestHeaders[k] = request.Header().Get(k)
				}

				responseHeaders := map[string]string{}
				for _, k := range response.Header().Keys() {
					responseHeaders[k] = response.Header().Get(k)
				}

				requestURLParams := map[string]string{}
				for k, v := range request.URL().QueryParams() {
					requestURLParams[k] = strings.Join(v, ",")
				}

				requestFormParams := map[string]string{}
				for k, v := range request.FormParams() {
					requestFormParams[k] = strings.Join(v, ",")
				}

				buf := new(bytes.Buffer)
				buf.ReadFrom(request.Body())
				requestBody := buf.String()

				apiCall := models.ApiCall{
					CurrentPath:      request.URI(),
					MethodType:       request.Method(),
					PostForm:         requestFormParams,
					RequestHeader:    requestHeaders,
					ResponseHeader:   responseHeaders,
					RequestUrlParams: requestURLParams,
					RequestBody:      requestBody,
					ResponseBody:     fakeWriter.Body.String(),
					ResponseCode:     fakeWriter.Code,
				}

				go yaag.GenerateHtml(&apiCall)
			}

			return p
		}
	}
}
