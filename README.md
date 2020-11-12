# seeder
blog loader for local file hosting

The basic unit of seeder is a **seed**, which contains a source and target, then creates the pipe between them. The different sources and targets are setup in packages to support growth later.

## Sources

seeder supports the following sources:
* [AWS Systems Manager Parameter Store](#aws-systems-manager-parameter-store)
* [Amazon S3](#amazon-s3)

### AWS Systems Manager Parameter Store

Seeds can be loaded from parameters stored in AWS Systems Manager Parameter Store by specifying the name of the parameter.

If the parameter is stored as a _SecureString_ (encrypted), it will be decrypted using KMS.

#### Permissions

Parameter seeds require the `ssm:GetParameter` permission, optionally specifying the parameter ARN as a resource.

If the parameter is stored as a _SecureString_ (encrypted), then you must also have the `kms:Decrypt` permission for the key used by the parameter, optionally specifying the key ARN as a resource. If you do not have the IAM permissions, you can optionally add the IAM user/role for seeder to the key policy for the key.

### Amazon S3

Seeds can be loaded from Amazon S3 by specifying the bucket and key as a S3 URI (similar to `aws s3` commands).

#### Permissions

Object seeds require the `s3:GetObject` permission, optionally specifying the bucket/key ARN as a resource. If you do not have the IAM permissions, you can optionally add the IAM user/role for seeder to the bucket policy for the bucket.

## Targets

seeder supports the following targets:
* [Local File](#local-file)

### Local File

Seeds can be stored locally as a file by specifying the path and file name. Optionally, set a default path to store all files from all seeds in the same location.

## Examples

### Certificate chain/private key

We can pull a TLS certificate and key stored in Parameter Store and put into a local `/certs` folder. This is useful for attaching seeder as a sidecar to an Envoy container to load the certificate and key into Envoy without having Envoy pull the certificate and key from another source.

In this example, `chain` is a _String_ parameter and `key` is a _SecureString_ parameter.

```yaml
apiVersion: v1alpha1
default:
  target:
    path: /certs
seeds:
- name: chain
  source:
    type: ssm-parameter
    spec:
      name: /certificates/greeter_server/chain
  target:
    type: file
    spec:
      name: chain.pem
- name: key
  source:
    type: ssm-parameter
    spec:
      name: /certificates/greeter_server/key
  target:
    type: file
    spec:
      name: key.pem
```
