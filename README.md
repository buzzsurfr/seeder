# cert-seeder (python3)

**Author**: buzzsurfr (Theo Salvo)

**PROVIDED AS-IS WITHOUT WARRANTY. SAMPLE NOT AFFILIATED WITH AMAZON IN ANY WAY.**

## Usage

### Environment variables

* `CHAIN_S3URI`: The S3Uri to the certificate chain. For example, `s3://mycertificates/app/chain.pem`
* `KEY_S3URI`: The S3Uri to the private key. For example, `s3://mycertificates/app/key.pem`
* `OUTPUT_DIR`: The location to store the chain and key. Default: `/tmp/certificates`

### Run as script

```
python3 main.py
```

### Build/run container

```
docker build -t <tag> .
docker run -e "CHAIN_S3URI=s3://mycertificates/app/chain.pem" -e "KEY_S3URI=s3://mycertificates/app/key.pem" <tag>
```

This has been built and stored in [Docker Hub](https://hub.docker.com/repository/docker/buzzsurfr/cert-seeder) as `buzzsurfr/cert-seeder:v0.1-python3`.

When testing locally, it may be necessary to provide `AWS_DEFAULT_REGION` as an environment variable and associate your credentials to the container. To do so using `us-east-1` as the region, run:

```
docker run -v ~/.aws:/root/aws -e "AWS_DEFAULT_REGION=us-east-1" -e "CHAIN_S3URI=s3://mycertificates/app/chain.pem" -e "KEY_S3URI=s3://mycertificates/app/key.pem" <tag>
```

### Add to Envoy container in ECS Task Definition

A sample task definition is provided at [ecs-task-definition.json](ecs-task-definition.json).

#### Steps

1. Add a volume to your task definiton.

    ```json
    "volumes": [
        {
            "host": {},
            "name": "certificates"
        }
    ]
    ```

1. Add a new container definition for the **cert-seeder**. Note that the **mountPoint** for our created volume matches the `OUTPUT_DIR` environment variable. This does not need to match in the App Mesh Virtual Node configuration or the Envoy container configuration. The user `1337` is so that Envoy ignores this container for proxy (which it shouldn't see anyways).,

    ```json
    {
        "name": "cert-seeder",
        "image": "buzzsurfr/cert-seeder:v0.1-python3",
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
    ```

1. Modify the **envoy** container definition adding the volume and setting a dependency on the **cert-seeder** container. The **containerPath** used here should match the **Local file hosting** path in the App Mesh Virtual Node configuration.

    ```json
    "mountPoints": [
        {
            "sourceVolume": "certificates",
            "containerPath": "/cert",
            "readOnly": true
        }
    ],
    "dependsOn": [
        {
            "containerName": "cert-seeder",
            "condition": "COMPLETE"
        }
    ]
    ```