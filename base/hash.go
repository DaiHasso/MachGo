package base

import (
    "github.com/pkg/errors"
    "github.com/cespare/xxhash"
    "github.com/mitchellh/hashstructure"
)

func HashObject(object Base) (uint64, error) {
    hasher := xxhash.New()
    resHash, err := hashstructure.Hash(
        object,
        &hashstructure.HashOptions{
            Hasher: hasher,
            TagName: "hash",
            ZeroNil: false,
        },
    )

    if err != nil {
        return 0, errors.Wrap(err, "Failed to generate hash for object")
    }

    return resHash, nil
}
