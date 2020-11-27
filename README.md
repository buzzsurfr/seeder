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
seeds:
- name: chain
  source:
    type: ssm-parameter
    spec:
      name: /certificates/app/chain
  target:
    type: file
    spec:
      path: /certs
      name: chain.pem
- name: key
  source:
    type: ssm-parameter
    spec:
      name: /certificates/app/key
  target:
    type: file
    spec:
      path: /certs
      name: key.pem
```

## Use Cases

### Add certificate/key to Envoy container for Amazon ECS and AWS App Mesh

The first use case (which led to the creation of seeder) is for providing your own CA to enable [Transport Layer Security (TLS)](https://docs.aws.amazon.com/app-mesh/latest/userguide/tls.html) using **Local file hosting**.

When creating a Virtual Node listener and enabling TLS termination using Local file hosting, you must specify the path to the **Certificate chain** and **Private key** files on the file system where the Envoy proxy is deployed. This means _inside_ the Envoy container.

Seeder can be deployed as an additional sidecar (it doesn't have to be, but it's easier) that shares a volume with the Envoy container.

Suppose you have the following ECS Task Definition (which is already setup to work with App Mesh):

```json
{
    "executionRoleArn": "arn:aws:iam::{{ACCOUNT_ID}}:role/ecsTaskExecutionRole",
    "containerDefinitions": [
        {
            "name": "app"
            ...
        },
        {
            "environment": [
                {
                    "name": "APPMESH_VIRTUAL_NODE_NAME",
                    "value": "mesh/greeter/virtualNode/greeter-v1"
                }
            ],
            "image": "840364872350.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-envoy:v1.15.0.0-prod",
            "name": "envoy",
        },
    ],
    "taskRoleArn": "arn:aws:iam::{{ACCOUNT_ID}}:role/AppMeshEnvoyTaskRole",
    "family": "app",
    "proxyConfiguration": {
        "type": "APPMESH",
        "containerName": "envoy",
        "properties": [
            {
                "name": "ProxyIngressPort",
                "value": "15000"
            },
            {
                "name": "AppPorts",
                "value": "8080"
            },
            {
                "name": "EgressIgnoredIPs",
                "value": "169.254.170.2,169.254.169.254"
            },
            {
                "name": "IgnoredGID",
                "value": ""
            },
            {
                "name": "EgressIgnoredPorts",
                "value": ""
            },
            {
                "name": "IgnoredUID",
                "value": "1337"
            },
            {
                "name": "ProxyEgressPort",
                "value": "15001"
            }
        ]
    }
}
```

Add `seeder` to the task definition and make envoy share a volume and be dependent on the seeder container.

```json
{
    "executionRoleArn": "arn:aws:iam::{{ACCOUNT_ID}}:role/ecsTaskExecutionRole",
    "containerDefinitions": [
        {
            "name": "app"
            ...
        },
        {
            "mountPoints": [
                {
                    "sourceVolume": "certificates",
                    "containerPath": "/certs",
                    "readOnly": true
                }
            ],
            "image": "840364872350.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-envoy:v1.15.0.0-prod",
            "dependsOn": [
                {
                    "containerName": "seeder",
                    "condition": "COMPLETE"
                }
            ],
            "name": "envoy",
        },
        {
            "name": "seeder",
            "image": "buzzsurfr/seeder:v0.1.0",
            "memoryReservation": "16",
            "essential": false,
            "environment": [
                {
                    "name": "CHAIN_S3URI",
                    "value": "s3://mycertificates/greeter_server/chain.pem"
                },
                {
                    "name": "KEY_S3URI",
                    "value": "s3://mycertificates/greeter_server/key.pem"
                },
                {
                    "name": "OUTPUT_DIR",
                    "value": "/tmp/certificates"
                }
            ],
            "user": "1337",
            "mountPoints": [
                {
                    "sourceVolume": "certificates",
                    "containerPath": "/tmp/certificates",
                    "readOnly": ""
                }
            ]
        }
    ],
    "taskRoleArn": "arn:aws:iam::{{ACCOUNT_ID}}:role/AppMeshEnvoyTaskRole",
    "family": "app",
    "proxyConfiguration": ...,
    "volumes": [
        {
            "host": {},
            "name": "certificates"
        }
    ]
}
```
