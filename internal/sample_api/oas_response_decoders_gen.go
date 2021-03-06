// Code generated by ogen, DO NOT EDIT.

package api

import (
	"bytes"
	"io"
	"mime"
	"net/http"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"go.opentelemetry.io/otel/trace"

	"github.com/ogen-go/ogen/validate"
)

func decodeDataGetFormatResponse(resp *http.Response, span trace.Span) (res string, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response string
			if err := func() error {
				v, err := d.Str()
				response = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeDefaultTestResponse(resp *http.Response, span trace.Span) (res int32, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response int32
			if err := func() error {
				v, err := d.Int32()
				response = int32(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeErrorGetResponse(resp *http.Response, span trace.Span) (res ErrorStatusCode, err error) {
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response Error
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return ErrorStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodeFoobarGetResponse(resp *http.Response, span trace.Span) (res FoobarGetRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response Pet
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	case 404:
		// Code 404.
		return &NotFound{}, nil
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeFoobarPostResponse(resp *http.Response, span trace.Span) (res FoobarPostRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response Pet
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	case 404:
		// Code 404.
		return &NotFound{}, nil
	}
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response Error
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return &ErrorStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodeFoobarPutResponse(resp *http.Response, span trace.Span) (res FoobarPutDef, err error) {
	// Default response.
	return FoobarPutDef{
		StatusCode: resp.StatusCode,
	}, nil
}

func decodeGetHeaderResponse(resp *http.Response, span trace.Span) (res Hash, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response Hash
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeNoAdditionalPropertiesTestResponse(resp *http.Response, span trace.Span) (res NoAdditionalPropertiesTest, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response NoAdditionalPropertiesTest
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeNullableDefaultResponseResponse(resp *http.Response, span trace.Span) (res NilIntStatusCode, err error) {
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response NilInt
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return NilIntStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodeOneofBugResponse(resp *http.Response, span trace.Span) (res OneofBugOK, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		return OneofBugOK{}, nil
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodePatternRecursiveMapGetResponse(resp *http.Response, span trace.Span) (res PatternRecursiveMap, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response PatternRecursiveMap
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodePetCreateResponse(resp *http.Response, span trace.Span) (res Pet, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response Pet
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodePetFriendsNamesByIDResponse(resp *http.Response, span trace.Span) (res []string, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response []string
			if err := func() error {
				response = make([]string, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem string
					v, err := d.Str()
					elem = string(v)
					if err != nil {
						return err
					}
					response = append(response, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodePetGetResponse(resp *http.Response, span trace.Span) (res PetGetRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response Pet
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response PetGetDef
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return &PetGetDefStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodePetGetAvatarByIDResponse(resp *http.Response, span trace.Span) (res PetGetAvatarByIDRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/octet-stream":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			response := PetGetAvatarByIDOK{Data: bytes.NewReader(b)}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	case 404:
		// Code 404.
		return &NotFound{}, nil
	}
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response Error
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return &ErrorStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodePetGetAvatarByNameResponse(resp *http.Response, span trace.Span) (res PetGetAvatarByNameRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/octet-stream":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			response := PetGetAvatarByNameOK{Data: bytes.NewReader(b)}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	case 404:
		// Code 404.
		return &NotFound{}, nil
	}
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response Error
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return &ErrorStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodePetGetByNameResponse(resp *http.Response, span trace.Span) (res Pet, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response Pet
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodePetNameByIDResponse(resp *http.Response, span trace.Span) (res string, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response string
			if err := func() error {
				v, err := d.Str()
				response = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodePetUpdateNameAliasPostResponse(resp *http.Response, span trace.Span) (res PetUpdateNameAliasPostDef, err error) {
	// Default response.
	return PetUpdateNameAliasPostDef{
		StatusCode: resp.StatusCode,
	}, nil
}

func decodePetUpdateNamePostResponse(resp *http.Response, span trace.Span) (res PetUpdateNamePostDef, err error) {
	// Default response.
	return PetUpdateNamePostDef{
		StatusCode: resp.StatusCode,
	}, nil
}

func decodePetUploadAvatarByIDResponse(resp *http.Response, span trace.Span) (res PetUploadAvatarByIDRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		return &PetUploadAvatarByIDOK{}, nil
	case 404:
		// Code 404.
		return &NotFound{}, nil
	}
	// Default response.
	ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return res, errors.Wrap(err, "parse media type")
	}
	switch {
	case ct == "application/json":
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}

		d := jx.DecodeBytes(b)
		var response Error
		if err := func() error {
			if err := response.Decode(d); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return res, err
		}
		return &ErrorStatusCode{
			StatusCode: resp.StatusCode,
			Response:   response,
		}, nil
	default:
		return res, validate.InvalidContentType(ct)
	}
}

func decodeRecursiveArrayGetResponse(resp *http.Response, span trace.Span) (res RecursiveArray, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response RecursiveArray
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeRecursiveMapGetResponse(resp *http.Response, span trace.Span) (res RecursiveMap, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response RecursiveMap
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeSecurityTestResponse(resp *http.Response, span trace.Span) (res string, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response string
			if err := func() error {
				v, err := d.Str()
				response = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeStringIntMapGetResponse(resp *http.Response, span trace.Span) (res StringIntMap, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response StringIntMap
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeTestContentParameterResponse(resp *http.Response, span trace.Span) (res string, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response string
			if err := func() error {
				v, err := d.Str()
				response = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeTestFloatValidationResponse(resp *http.Response, span trace.Span) (res TestFloatValidationOK, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		return TestFloatValidationOK{}, nil
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeTestNullableOneofsResponse(resp *http.Response, span trace.Span) (res TestNullableOneofsRes, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response TestNullableOneofsApplicationJSONOK
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	case 201:
		// Code 201.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response TestNullableOneofsApplicationJSONCreated
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	case 202:
		// Code 202.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response OneOfBooleanSumNullables
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return &response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}

func decodeTestObjectQueryParameterResponse(resp *http.Response, span trace.Span) (res TestObjectQueryParameterOK, err error) {
	switch resp.StatusCode {
	case 200:
		// Code 200.
		ct, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return res, errors.Wrap(err, "parse media type")
		}
		switch {
		case ct == "application/json":
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return res, err
			}

			d := jx.DecodeBytes(b)
			var response TestObjectQueryParameterOK
			if err := func() error {
				if err := response.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return res, err
			}
			return response, nil
		default:
			return res, validate.InvalidContentType(ct)
		}
	}
	return res, validate.UnexpectedStatusCode(resp.StatusCode)
}
