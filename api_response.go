package do

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rightjoin/fig"
)

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Errors  []ErrorPlus `json:"errors"`
}

func (a *ApiResponse) SetData(data interface{}) error {
	a.Data = data
	a.Success = true
	return nil
}

func (a *ApiResponse) AddError(errs ...error) error {
	if a.Errors == nil {
		a.Errors = []ErrorPlus{}
	}

	tmp := make([]ErrorPlus, len(errs))
	for i := 0; i < len(errs); i++ {
		tmp[i] = ErrorPlus{Message: errs[i].Error()}
	}

	a.Errors = append(a.Errors, tmp...)

	return nil
}

func (a *ApiResponse) AddErrorPlus(errs ...ErrorPlus) error {
	if a.Errors == nil {
		a.Errors = []ErrorPlus{}
	}

	a.Errors = append(a.Errors, errs...)

	return nil
}

func (a *ApiResponse) Scribe(webContext interface{}) {

	c, ok := webContext.(echo.Context)

	if !ok {
		// TODO: panic
	}

	c.JSON(http.StatusOK, a)
}

type Paging struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	Current int `json:"current"`
	Total   int `json:"total"`
	Chunk   int `json:"chunk"`
}

type ApiPageResponse struct {
	ApiResponse
	Paging `json:"paging"`
}

func (a *ApiPageResponse) Scribe(webContext interface{}) {

	c, ok := webContext.(echo.Context)

	if !ok {
		// TODO: panic
	}

	c.JSON(http.StatusOK, a)
}

func (a *ApiPageResponse) SetData(d interface{}, current, total int) error {
	a.Data = d
	a.Success = true

	a.Current = current
	a.Total = total
	a.Pages = int(math.Ceil(float64(total) / float64(a.Chunk)))

	return nil
}

func NewApiPageResponse(page, chunk int) ApiPageResponse {

	if page <= 0 {
		page = 1
	}

	max := fig.IntOr(25, "pagination.chunk")
	if chunk < 1 || chunk > max {
		chunk = max
	}

	return ApiPageResponse{
		Paging: Paging{
			Page:    page,
			Pages:   0,
			Total:   0,
			Current: 0,
			Chunk:   chunk,
		},
	}
}

func NewApiPageResponseFromContext(webContext interface{}) ApiPageResponse {

	page := 0
	chunk := 0

	if c, ok := webContext.(echo.Context); ok {
		page = ParseIntOr(c.QueryParam(":page"), 0)
		chunk = ParseIntOr(c.QueryParam(":chunk"), 0)
	}

	return NewApiPageResponse(page, chunk)
}
