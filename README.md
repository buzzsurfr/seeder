# seeder
Certificate loader for Envoy local file hosting

## Sources

seeder downloads a certificate chain and private key for use locally.

seeder supports the following sources:
* Amazon S3
* AWS Systems Manager Parameter Store

The certificate chain will be written to the file **chain.pem** and the private key will be written to the file **key.pem**. The default directory to write both files is `/tmp/certificates`, but can be changed by setting the `OUTPUT_DIR` parameter.

### Certificate Chain

* `CHAIN_PARAMETER_STORE_NAME`: The name of the Parameter Store parameter with the certificate chain. For example, `/certificates/sample/chain`.
* `CHAIN_S3URI`: The S3Uri of the certificate chain. For example: `s3://mybucket/certificate-path/chain.pem`.

### Private Key

* `KEY_PARAMETER_STORE_NAME`: The name of the Parameter store parameters with the private key. For example, `/certificates/sample/key`.
* `KEY_S3URI`: The S3Uri of the private key. For example: `s3://mybucket/certificate-path/key.pem`.

## TO DO

* Rename to `seeder`
* Create a CLI called `plant` that _plants_ seeds into configurations.
* Support a config file
* Add a watcher