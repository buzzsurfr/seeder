#!/usr/bin/env python3

import os
import sys
from urllib.parse import urlparse
import boto3

class S3Uri(object):
    def __init__(self, url):
        self._parsed = urlparse(url, allow_fragments=False)

    @property
    def bucket(self):
        return self._parsed.netloc

    @property
    def key(self):
        if self._parsed.query:
            return self._parsed.path.lstrip('/') + '?' + self._parsed.query
        else:
            return self._parsed.path.lstrip('/')

    @property
    def url(self):
        return self._parsed.geturl()

chain_s3uri = os.getenv("CHAIN_S3URI")
key_s3uri = os.getenv("KEY_S3URI")
output_dir = os.getenv("OUTPUT_DIR", "/tmp/certificates")

if not chain_s3uri or not key_s3uri:
    print("Error: must provide CHAIN_S3URI and KEY_S3URI as environment variables.\n")
    sys.exit(1)

if not os.path.exists(output_dir):
    os.makedirs(output_dir)

chain = S3Uri(chain_s3uri)
key = S3Uri(key_s3uri)

s3 = boto3.resource('s3')

# Get certificate chain from S3 and save to file
s3.meta.client.download_file(Bucket=chain.bucket,
                             Key=chain.key,
                             Filename=os.path.join(output_dir,"chain.pem"))

# Get private key from S3 and save to file
s3.meta.client.download_file(Bucket=key.bucket,
                             Key=key.key,
                             Filename=os.path.join(output_dir,"key.pem"))
