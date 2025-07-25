// Code generated by ogen, DO NOT EDIT.

package api

import (
	"net/http"
	"net/url"

	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/conv"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
)

// ListPetsParams is parameters of listPets operation.
type ListPetsParams struct {
	// How many items to return at one time (max 100).
	Limit OptInt32
}

func unpackListPetsParams(packed middleware.Parameters) (params ListPetsParams) {
	{
		key := middleware.ParameterKey{
			Name: "limit",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Limit = v.(OptInt32)
		}
	}
	return params
}

func decodeListPetsParams(args [0]string, argsEscaped bool, r *http.Request) (params ListPetsParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	// Decode query: limit.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "limit",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotLimitVal int32
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt32(val)
					if err != nil {
						return err
					}

					paramsDotLimitVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Limit.SetTo(paramsDotLimitVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "limit",
			In:   "query",
			Err:  err,
		}
	}
	return params, nil
}

// ShowPetByIdParams is parameters of showPetById operation.
type ShowPetByIdParams struct {
	// The id of the pet to retrieve.
	PetId string
}

func unpackShowPetByIdParams(packed middleware.Parameters) (params ShowPetByIdParams) {
	{
		key := middleware.ParameterKey{
			Name: "petId",
			In:   "path",
		}
		params.PetId = packed[key].(string)
	}
	return params
}

func decodeShowPetByIdParams(args [1]string, argsEscaped bool, r *http.Request) (params ShowPetByIdParams, _ error) {
	// Decode path: petId.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "petId",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.PetId = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "petId",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}
