package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func DocumentView(c echo.Context) error {
	return c.Render(http.StatusOK, "document", nil)
}

type newResponse struct {
	Title  string
	Status string
}

func DocumentNew(c echo.Context) error {
	return c.Render(http.StatusOK, "document-new", nil)
}

func DocumentNewPut(c echo.Context) error {
	resp := editResponse{
		Title:  c.FormValue("title"),
		Status: "Success",
	}

	return c.Render(http.StatusOK, "document-new/put", resp)
}

type editResponse struct {
	Title  string
	Status string
	Edit   string
}

func DocumentEdit(c echo.Context) error {
	return c.Render(http.StatusOK, "document-edit", nil)
}

func DocumentEditPatch(c echo.Context) error {
	resp := editResponse{
		Title:  c.FormValue("title"),
		Status: "Success",
		Edit:   c.FormValue("edit"),
	}

	return c.Render(http.StatusOK, "document-edit/patch", resp)
}

type deleteResponse struct {
	Title  string
	Status string
}

func DocumentDelete(c echo.Context) error {
	return c.Render(http.StatusOK, "document-delete", nil)
}

func DocumentDeleteDelete(c echo.Context) error {
	resp := editResponse{
		Title:  c.FormValue("title"),
		Status: "Success",
	}

	return c.Render(http.StatusOK, "document-delete/delete", resp)
}
