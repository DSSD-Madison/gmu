package web

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// TODO: change to context.Context
func Render(ctx echo.Context, statusCode int, comps ...templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	for _, t := range comps {
		if err := t.Render(ctx.Request().Context(), buf); err != nil {
			return err
		}
	}

	return ctx.HTML(statusCode, buf.String())
}
