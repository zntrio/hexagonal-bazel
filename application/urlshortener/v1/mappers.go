package v1

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/pkg/types"
)

func fromLink(d link.Link) *urlshortener.Link {
	var expiresAt *timestamppb.Timestamp
	if d.GetExpiresAt() != nil {
		expiresAt = timestamppb.New(*d.GetExpiresAt())
	}

	return &urlshortener.Link{
		Id:             string(d.GetID()),
		Url:            types.AsRef(d.GetURL()),
		SecretRequired: d.IsProtected(),
		ExpiresAt:      expiresAt,
	}
}
