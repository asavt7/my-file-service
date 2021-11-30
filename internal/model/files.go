package model

import "io"

const (
	AvatarsBucket = "avatars"
)

type FileToStore struct {
	Body io.ReadSeekCloser

	Name   string
	Bucket string
}

type StoredFile struct {
	Name   string
	Bucket string
}

type FileToDownload struct {
	Name   string
	Bucket string
}

type LoadedFile struct {
	Name string
	Body []byte
}
