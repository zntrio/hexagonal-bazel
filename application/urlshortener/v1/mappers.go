package v1

import (
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/pkg/types"
)

func fromLink(d link.Link) *urlshortener.Link {
	return &urlshortener.Link{
		Id:             string(d.GetID()),
		Url:            types.AsRef(d.GetURL()),
		SecretRequired: d.GetSecretHash() != "",
	}
}
