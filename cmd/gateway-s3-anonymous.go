/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"encoding/hex"
	"io"

	minio "github.com/minio/minio-go"
)

// AnonPutObject creates a new object anonymously with the incoming data,
func (l *s3Gateway) AnonPutObject(bucket string, object string, size int64, data io.Reader, metadata map[string]string, sha256sum string) (ObjectInfo, error) {
	var sha256sumBytes []byte

	var err error
	if sha256sum != "" {
		sha256sumBytes, err = hex.DecodeString(sha256sum)
		if err != nil {
			return ObjectInfo{}, s3ToObjectError(traceError(err), bucket, object)
		}
	}

	var md5sumBytes []byte
	md5sum := metadata["md5Sum"]
	if md5sum != "" {
		md5sumBytes, err = hex.DecodeString(md5sum)
		if err != nil {
			return ObjectInfo{}, s3ToObjectError(traceError(err), bucket, object)
		}
		delete(metadata, "md5Sum")
	}

	oi, err := l.anonClient.PutObject(bucket, object, size, data, md5sumBytes, sha256sumBytes, toMinioClientMetadata(metadata))
	if err != nil {
		return ObjectInfo{}, s3ToObjectError(traceError(err), bucket, object)
	}

	return fromMinioClientObjectInfo(bucket, oi), nil
}

// AnonGetObject - Get object anonymously
func (l *s3Gateway) AnonGetObject(bucket string, key string, startOffset int64, length int64, writer io.Writer) error {
	r := minio.NewGetReqHeaders()
	if err := r.SetRange(startOffset, startOffset+length-1); err != nil {
		return s3ToObjectError(traceError(err), bucket, key)
	}
	object, _, err := l.anonClient.GetObject(bucket, key, r)
	if err != nil {
		return s3ToObjectError(traceError(err), bucket, key)
	}

	defer object.Close()

	if _, err := io.CopyN(writer, object, length); err != nil {
		return s3ToObjectError(traceError(err), bucket, key)
	}

	return nil
}

// AnonGetObjectInfo - Get object info anonymously
func (l *s3Gateway) AnonGetObjectInfo(bucket string, object string) (ObjectInfo, error) {
	r := minio.NewHeadReqHeaders()
	oi, err := l.anonClient.StatObject(bucket, object, r)
	if err != nil {
		return ObjectInfo{}, s3ToObjectError(traceError(err), bucket, object)
	}

	return fromMinioClientObjectInfo(bucket, oi), nil
}

// AnonListObjects - List objects anonymously
func (l *s3Gateway) AnonListObjects(bucket string, prefix string, marker string, delimiter string, maxKeys int) (ListObjectsInfo, error) {
	result, err := l.anonClient.ListObjects(bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		return ListObjectsInfo{}, s3ToObjectError(traceError(err), bucket)
	}

	return fromMinioClientListBucketResult(bucket, result), nil
}

// AnonGetBucketInfo - Get bucket metadata anonymously.
func (l *s3Gateway) AnonGetBucketInfo(bucket string) (BucketInfo, error) {
	if exists, err := l.anonClient.BucketExists(bucket); err != nil {
		return BucketInfo{}, s3ToObjectError(traceError(err), bucket)
	} else if !exists {
		return BucketInfo{}, traceError(BucketNotFound{Bucket: bucket})
	}

	buckets, err := l.anonClient.ListBuckets()
	if err != nil {
		return BucketInfo{}, s3ToObjectError(traceError(err), bucket)
	}

	for _, bi := range buckets {
		if bi.Name != bucket {
			continue
		}

		return BucketInfo{
			Name:    bi.Name,
			Created: bi.CreationDate,
		}, nil
	}

	return BucketInfo{}, traceError(BucketNotFound{Bucket: bucket})
}
