package parser

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func fromOpenapiInfo(info openapi.Info) ogen.Info {
	result := ogen.Info{
		Title:          info.Title,
		Summary:        info.Summary,
		Description:    info.Description,
		TermsOfService: info.TermsOfService,
		Version:        info.Version,
		Extensions:     info.Extensions,
	}

	if info.Contact != nil {
		val := ogen.Contact(*info.Contact)
		result.Contact = &val
	}

	if info.License != nil {
		val := ogen.License(*info.License)
		result.License = &val
	}

	return result
}

func fromOgenInfo(info ogen.Info) openapi.Info {
	result := openapi.Info{
		Title:          info.Title,
		Summary:        info.Summary,
		Description:    info.Description,
		TermsOfService: info.TermsOfService,
		Version:        info.Version,
		Extensions:     info.Extensions,
	}

	if info.Contact != nil {
		val := openapi.Contact(*info.Contact)
		result.Contact = &val
	}

	if info.License != nil {
		val := openapi.License(*info.License)
		result.License = &val
	}

	return result
}
