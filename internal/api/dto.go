package api

type CreateLinkPayload struct {
	OriginalUrl string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

type RedirectUriParams struct {
	ShortName string `uri:"code" binding:"required"`
}

type GetEntityUriParams struct {
	ID int `uri:"id" binding:"required"`
}

type Range struct {
	Begin int
	End   int
}

type GetMultipleEntityQueryParams struct {
	Range Range `form:"range" binding:"required"`
}
